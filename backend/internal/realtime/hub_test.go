package realtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// members maps a test token (header X-User) to its sender; campaign "c1" has all of them.
var members = map[string]Sender{
	"gm":    {UserID: "u-gm", Name: "Mestre", Role: "gm"},
	"ana":   {UserID: "u-ana", Name: "Ana", Role: "player"},
	"bruno": {UserID: "u-bruno", Name: "Bruno", Role: "player"},
}

func newServer(t *testing.T, h *Hub) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/campaigns/{id}/ws", func(w http.ResponseWriter, r *http.Request) {
		s, ok := members[r.Header.Get("X-User")]
		if !ok || r.PathValue("id") != "c1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.Serve(w, r, "c1", s)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// client reads in a goroutine because a cancelled Read closes a coder/websocket connection.
type client struct {
	ws   *websocket.Conn
	msgs chan Message
}

func dial(t *testing.T, srv *httptest.Server, user string) *client {
	t.Helper()
	ws, resp, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http")+"/api/campaigns/c1/ws",
		&websocket.DialOptions{HTTPHeader: http.Header{"X-User": {user}}})
	if err != nil {
		t.Fatalf("dial %s: %v", user, err)
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	t.Cleanup(func() { _ = ws.CloseNow() })
	c := &client{ws: ws, msgs: make(chan Message, 64)}
	go func() {
		for {
			_, data, err := ws.Read(context.Background())
			if err != nil {
				close(c.msgs)
				return
			}
			var m Message
			_ = json.Unmarshal(data, &m)
			c.msgs <- m
		}
	}()
	return c
}

func read(t *testing.T, c *client) Message {
	t.Helper()
	select {
	case m, ok := <-c.msgs:
		if !ok {
			t.Fatal("connection closed")
		}
		return m
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
	return Message{}
}

func expectSilence(t *testing.T, c *client) {
	t.Helper()
	select {
	case m := <-c.msgs:
		t.Fatalf("unexpected message: %+v", m)
	case <-time.After(150 * time.Millisecond):
	}
}

func send(t *testing.T, c *client, m Message) {
	t.Helper()
	data, _ := json.Marshal(m)
	if err := c.ws.Write(context.Background(), websocket.MessageText, data); err != nil {
		t.Fatal(err)
	}
}

func presenceNames(t *testing.T, m Message) []string {
	t.Helper()
	if m.Type != EventPresence {
		t.Fatalf("want presence, got %s", m.Type)
	}
	var p struct{ Users []Sender }
	if err := json.Unmarshal(m.Payload, &p); err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, u := range p.Users {
		names = append(names, u.Name)
	}
	return names
}

func eq(a, b []string) bool { return strings.Join(a, ",") == strings.Join(b, ",") }

func TestNonMemberRefusedAtHandshake(t *testing.T) {
	srv := newServer(t, NewHub(Options{}))
	_, resp, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http")+"/api/campaigns/c1/ws",
		&websocket.DialOptions{HTTPHeader: http.Header{"X-User": {"stranger"}}})
	if err == nil {
		t.Fatal("expected handshake refusal")
	}
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404, got %v", resp)
	}
}

func TestPresenceJoinLeaveAndRoomRelease(t *testing.T) {
	h := NewHub(Options{})
	srv := newServer(t, h)
	gm := dial(t, srv, "gm")
	if got := presenceNames(t, read(t, gm)); !eq(got, []string{"Mestre"}) {
		t.Fatalf("got %v", got)
	}
	ana := dial(t, srv, "ana")
	if got := presenceNames(t, read(t, ana)); !eq(got, []string{"Ana", "Mestre"}) {
		t.Fatalf("new client must get current state, got %v", got)
	}
	if got := presenceNames(t, read(t, gm)); !eq(got, []string{"Ana", "Mestre"}) {
		t.Fatalf("got %v", got)
	}
	_ = ana.ws.Close(websocket.StatusNormalClosure, "")
	if got := presenceNames(t, read(t, gm)); !eq(got, []string{"Mestre"}) {
		t.Fatalf("got %v", got)
	}
	_ = gm.ws.Close(websocket.StatusNormalClosure, "")
	waitFor(t, func() bool { h.mu.Lock(); defer h.mu.Unlock(); return len(h.rooms) == 0 })
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	for i := 0; i < 100; i++ {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not met")
}

func TestMultipleTabsStayOnlineUntilLastCloses(t *testing.T) {
	srv := newServer(t, NewHub(Options{}))
	gm := dial(t, srv, "gm")
	read(t, gm)
	tab1, tab2 := dial(t, srv, "ana"), dial(t, srv, "ana")
	read(t, tab1)
	read(t, gm)
	read(t, gm)
	_ = tab1.ws.Close(websocket.StatusNormalClosure, "")
	expectSilence(t, gm)
	_ = tab2.ws.Close(websocket.StatusNormalClosure, "")
	if got := presenceNames(t, read(t, gm)); !eq(got, []string{"Mestre"}) {
		t.Fatalf("got %v", got)
	}
}

func drain(_ *testing.T, cs ...*client) {
	time.Sleep(150 * time.Millisecond)
	for _, c := range cs {
		for len(c.msgs) > 0 {
			<-c.msgs
		}
	}
}

func TestBroadcastFilters(t *testing.T) {
	h := NewHub(Options{})
	srv := newServer(t, h)
	gm, ana, bruno := dial(t, srv, "gm"), dial(t, srv, "ana"), dial(t, srv, "bruno")
	drain(t, gm, ana, bruno)
	cases := []struct {
		name   string
		filter Filter
		want   map[*client]bool
	}{
		{"all", ToAll, map[*client]bool{gm: true, ana: true, bruno: true}},
		{"gm only", ToGM, map[*client]bool{gm: true}},
		{"user", ToUser("u-ana"), map[*client]bool{ana: true}},
		{"except", Except("u-ana"), map[*client]bool{gm: true, bruno: true}},
		{"predicate", Or(ToGM, ToUser("u-bruno")), map[*client]bool{gm: true, bruno: true}},
	}
	for _, tc := range cases {
		h.Broadcast("c1", tc.filter, "test.event", map[string]string{"case": tc.name})
		for _, c := range []*client{gm, ana, bruno} {
			if tc.want[c] {
				if m := read(t, c); m.Type != "test.event" {
					t.Fatalf("%s: got %s", tc.name, m.Type)
				}
			} else {
				expectSilence(t, c)
			}
		}
	}
}

func TestGMOnlyCommandAndPermissionError(t *testing.T) {
	h := NewHub(Options{})
	h.Handle("secret.roll", func(_ context.Context, cmd Command) error {
		h.Broadcast(cmd.CampaignID, ToGM, "secret.rolled", map[string]string{"by": cmd.Sender.Name})
		return nil
	}, GMOnly)
	srv := newServer(t, h)
	gm, ana := dial(t, srv, "gm"), dial(t, srv, "ana")
	drain(t, gm, ana)

	send(t, ana, Message{Type: "secret.roll", ReqID: "r1"})
	m := read(t, ana)
	if m.Type != EventError || !strings.Contains(string(m.Payload), `"reqId":"r1"`) ||
		!strings.Contains(string(m.Payload), "permissão") {
		t.Fatalf("want permission error, got %+v %s", m, m.Payload)
	}
	expectSilence(t, gm)

	send(t, gm, Message{Type: "secret.roll", ReqID: "r2"})
	got := []string{read(t, gm).Type, read(t, gm).Type}
	if !eq(got, []string{"secret.rolled", EventAck}) {
		t.Fatalf("got %v", got)
	}
	expectSilence(t, ana)
}

func TestPingPongOnlyToSender(t *testing.T) {
	srv := newServer(t, NewHub(Options{}))
	gm, ana := dial(t, srv, "gm"), dial(t, srv, "ana")
	drain(t, gm, ana)
	send(t, ana, Message{Type: "ping", ReqID: "p1"})
	if read(t, ana).Type != EventPong {
		t.Fatal("want pong")
	}
	if m := read(t, ana); m.Type != EventAck {
		t.Fatalf("want ack, got %s", m.Type)
	}
	expectSilence(t, gm)
}

func TestUnknownCommandAndHandlerErrors(t *testing.T) {
	h := NewHub(Options{})
	h.Handle("bad", func(context.Context, Command) error { return UserError("Movimento inválido") })
	srv := newServer(t, h)
	ana := dial(t, srv, "ana")
	drain(t, ana)
	send(t, ana, Message{Type: "nope", ReqID: "a"})
	if m := read(t, ana); m.Type != EventError || !strings.Contains(string(m.Payload), "desconhecida") {
		t.Fatalf("got %s", m.Payload)
	}
	send(t, ana, Message{Type: "bad", ReqID: "b"})
	if m := read(t, ana); !strings.Contains(string(m.Payload), "Movimento inválido") {
		t.Fatalf("got %s", m.Payload)
	}
}

func TestSlowConsumerDroppedWithoutBlockingHub(t *testing.T) {
	h := NewHub(Options{})
	c := &conn{sender: members["ana"], send: make(chan []byte, 1), cancel: func() {}}
	dropped := false
	c.cancel = func() { dropped = true }
	h.rooms["c1"] = map[*conn]struct{}{c: {}}
	h.Broadcast("c1", ToAll, "x", nil)
	h.Broadcast("c1", ToAll, "x", nil)
	if !dropped {
		t.Fatal("full queue must drop the connection")
	}
}
