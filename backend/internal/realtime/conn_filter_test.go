package realtime

import (
	"context"
	"testing"

	"github.com/coder/websocket"
)

func TestExceptConnSkipsOnlyOriginTab(t *testing.T) {
	h := NewHub(Options{})
	h.Handle("echo.test", func(_ context.Context, cmd Command) error {
		h.Broadcast(cmd.CampaignID, ExceptConn(cmd.ConnID), "echo.patched", nil)
		return nil
	}, GMOnly)
	srv := newServer(t, h)
	tabA, tabB, ana := dial(t, srv, "gm"), dial(t, srv, "gm"), dial(t, srv, "ana")
	drain(t, tabA, tabB, ana)
	if err := tabA.ws.Write(context.Background(), websocket.MessageText, []byte(`{"type":"echo.test","reqId":"1"}`)); err != nil {
		t.Fatal(err)
	}
	if m := read(t, tabA); m.Type != EventAck {
		t.Fatalf("sender got %s", m.Type)
	}
	for _, c := range []*client{tabB, ana} {
		if m := read(t, c); m.Type != "echo.patched" {
			t.Fatalf("got %s", m.Type)
		}
	}
	expectSilence(t, tabA)
}
