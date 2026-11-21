package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

const (
	queueSize    = 64
	writeTimeout = 10 * time.Second
	pingEvery    = 30 * time.Second
	readLimit    = 64 << 10
)

type conn struct {
	ws         *websocket.Conn
	hub        *Hub
	campaignID string
	sender     Sender
	send       chan []byte
	cancel     context.CancelFunc
}

// Serve upgrades the request and runs the connection until it closes. The caller must
// have authenticated the user and verified campaign membership already.
func (h *Hub) Serve(w http.ResponseWriter, r *http.Request, campaignID string, s Sender) {
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: h.opts.OriginPatterns})
	if err != nil {
		return
	}
	ws.SetReadLimit(readLimit)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	c := &conn{ws: ws, hub: h, campaignID: campaignID, sender: s, send: make(chan []byte, queueSize), cancel: cancel}
	go c.writeLoop(ctx)
	h.join(campaignID, c)
	defer h.leave(campaignID, c)
	c.readLoop(ctx)
	_ = ws.CloseNow()
}

// enqueue never blocks; a client that cannot keep up is disconnected.
func (c *conn) enqueue(data []byte) {
	select {
	case c.send <- data:
	default:
		c.cancel()
	}
}

func (c *conn) writeLoop(ctx context.Context) {
	tick := time.NewTicker(pingEvery)
	defer tick.Stop()
	defer c.cancel()
	for {
		var err error
		select {
		case <-ctx.Done():
			return
		case data := <-c.send:
			err = c.write(ctx, data)
		case <-tick.C:
			err = c.ping(ctx)
		}
		if err != nil {
			return
		}
	}
}

func (c *conn) write(ctx context.Context, data []byte) error {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return c.ws.Write(ctx, websocket.MessageText, data)
}

func (c *conn) ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return c.ws.Ping(ctx)
}

func (c *conn) readLoop(ctx context.Context) {
	for {
		_, data, err := c.ws.Read(ctx)
		if err != nil {
			return
		}
		var m Message
		if json.Unmarshal(data, &m) != nil || m.Type == "" {
			c.reply(EventError, "", map[string]string{"message": "Mensagem inválida"})
			continue
		}
		c.dispatch(ctx, m)
	}
}

func (c *conn) dispatch(ctx context.Context, m Message) {
	rt, ok := c.hub.lookup(m.Type)
	if !ok {
		c.fail(m.ReqID, "Ação desconhecida")
		return
	}
	for _, g := range rt.guards {
		if err := g(c.sender); err != nil {
			c.failErr(m.ReqID, err)
			return
		}
	}
	cmd := Command{CampaignID: c.campaignID, Sender: c.sender, Type: m.Type, Payload: m.Payload, ReqID: m.ReqID}
	if err := rt.fn(ctx, cmd); err != nil {
		c.failErr(m.ReqID, err)
		return
	}
	if m.ReqID != "" {
		c.reply(EventAck, m.ReqID, nil)
	}
}

func (c *conn) failErr(reqID string, err error) {
	msg, ok := userMessage(err)
	if !ok {
		slog.Error("realtime command failed", "err", err)
		msg = "Erro interno. Tente novamente"
	}
	c.fail(reqID, msg)
}

func (c *conn) fail(reqID, msg string) {
	c.reply(EventError, reqID, map[string]string{"message": msg})
}

// reply sends to this connection only; error/ack payloads carry the reqId.
func (c *conn) reply(typ, reqID string, payload map[string]string) {
	if payload == nil {
		payload = map[string]string{}
	}
	if reqID != "" {
		payload["reqId"] = reqID
	}
	data, err := encode(typ, payload)
	if err == nil {
		c.enqueue(data)
	}
}
