package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

func handlerFixture(t *testing.T) (*http.ServeMux, string) {
	t.Helper()
	repo := newFakeRepo()
	links := NewLinks(repo, fakeMembers{role: campaigns.RoleGM})
	secret, _, err := links.Generate(t.Context(), campaigns.Member{CampaignID: "c1", UserID: "gm", Role: campaigns.RoleGM})
	if err != nil {
		t.Fatal(err)
	}
	r := newRig("s1")
	mux := http.NewServeMux()
	mux.Handle("/api/mcp/{secret}", NewHandler(links, r.t, r.j))
	return mux, secret
}

func post(mux http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestHandlerRejectsUnknownSecret(t *testing.T) {
	mux, _ := handlerFixture(t)
	rec := post(mux, "/api/mcp/nope", `{}`)
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "Link inválido") {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerServesInitialize(t *testing.T) {
	mux, secret := handlerFixture(t)
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18",` +
		`"capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	rec := post(mux, "/api/mcp/"+secret, body)
	if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("got %d %s %s", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "rpgmaker-mesa") {
		t.Fatalf("missing server name: %s", rec.Body.String())
	}
}
