package systems

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultRegistryHasAssimilacao(t *testing.T) {
	s, ok := Default().Get("assimilacao")
	if !ok || s.Name() != "Assimilação RPG" {
		t.Fatalf("assimilacao missing: %v %v", s, ok)
	}
	if _, ok := Default().Get("nope"); ok {
		t.Error("unknown system found")
	}
}

func TestSystemsEndpointLists(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, Default(), func(h http.Handler) http.Handler { return h })
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/systems", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"id":"assimilacao"`) {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}
