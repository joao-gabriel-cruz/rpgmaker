package ambience

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

var senders = map[string]realtime.Sender{
	"gm":  {UserID: "u-gm", Name: "Mestre", Role: "gm"},
	"ana": {UserID: "u-ana", Name: "Ana", Role: "player"},
}

type peer struct {
	ws   *websocket.Conn
	msgs chan realtime.Message
}

func dial(t *testing.T, srv *httptest.Server, user string) *peer {
	t.Helper()
	ws, resp, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http"),
		&websocket.DialOptions{HTTPHeader: http.Header{"X-User": {user}}})
	if err != nil {
		t.Fatal(err)
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	t.Cleanup(func() { _ = ws.CloseNow() })
	p := &peer{ws: ws, msgs: make(chan realtime.Message, 32)}
	go func() {
		for {
			_, data, err := ws.Read(context.Background())
			if err != nil {
				return
			}
			var m realtime.Message
			_ = json.Unmarshal(data, &m)
			p.msgs <- m
		}
	}()
	return p
}

func (p *peer) next(t *testing.T, typ string) realtime.Message {
	t.Helper()
	for {
		select {
		case m := <-p.msgs:
			if m.Type == typ {
				return m
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("no %q message", typ)
		}
	}
}

func (p *peer) send(t *testing.T, typ, payload, reqID string) {
	t.Helper()
	b, _ := json.Marshal(realtime.Message{Type: typ, Payload: json.RawMessage(payload), ReqID: reqID})
	if err := p.ws.Write(context.Background(), websocket.MessageText, b); err != nil {
		t.Fatal(err)
	}
}

func TestAmbienceSetOverTheHubIsGMOnlyAndReachesEveryone(t *testing.T) {
	hub := realtime.NewHub(realtime.Options{})
	st := memStore{}
	chars := &fakeChars{}
	RegisterCommands(hub, NewService(st, fakeScenes{}, chars, hub))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(w, r, "c1", senders[r.Header.Get("X-User")])
	}))
	t.Cleanup(srv.Close)
	gm, ana := dial(t, srv, "gm"), dial(t, srv, "ana")

	ana.send(t, "ambience.set", `{"sceneId":"s1","timeOfDay":"night","weather":"rain"}`, "r1")
	if m := ana.next(t, "error"); !strings.Contains(string(m.Payload), `"reqId":"r1"`) {
		t.Fatalf("player must be refused: %+v", m)
	}
	if len(st) != 0 {
		t.Fatalf("player change was persisted: %v", st)
	}
	gm.send(t, "ambience.set", `{"sceneId":"s1","timeOfDay":"night","weather":"storm"}`, "r2")
	gm.next(t, "ack")
	var got Ambience
	_ = json.Unmarshal(ana.next(t, EventUpdated).Payload, &got)
	if got != (Ambience{SceneID: "s1", TimeOfDay: "night", Weather: "storm"}) || st["s1"] != got {
		t.Fatalf("broadcast %+v stored %+v", got, st)
	}
}
