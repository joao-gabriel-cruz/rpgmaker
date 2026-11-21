package systems

import (
	"encoding/json"
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

func TestAssimilacaoModuleValidatesAndDerives(t *testing.T) {
	sys, _ := Default().Get("assimilacao")
	v, d := sys.(SheetValidator), sys.(SheetDeriver)
	sheet := json.RawMessage(`{"instincts":{"potency":2,"resolution":1}}`)
	out, err := d.DeriveSheet(sheet)
	if err != nil || !strings.Contains(string(out), `"boxesPerLevel":4`) {
		t.Errorf("derive = %s %v", out, err)
	}
	if _, err := d.DeriveSheet(json.RawMessage(`{"instincts":"x"}`)); err == nil {
		t.Error("expected derive error")
	}
	bad := v.ValidateSheet(json.RawMessage(`{"instincts":"x"}`), nil)
	if len(bad) != 1 || bad[0].Severity != SeverityBlocking {
		t.Errorf("bad = %+v", bad)
	}
	over := v.ValidateSheet(json.RawMessage(`{"instincts":{"influence":3,"perception":3,"potency":3,"reaction":3,"resolution":3,"sagacity":3}}`), nil)
	if len(over) == 0 || over[0].Severity != SeverityLimit {
		t.Errorf("over = %+v", over)
	}
}
