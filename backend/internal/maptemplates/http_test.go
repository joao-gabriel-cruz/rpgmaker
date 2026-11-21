package maptemplates

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func gmServer() *http.ServeMux {
	mux := http.NewServeMux()
	RegisterHTTP(mux, func(h http.Handler) http.Handler { return h })
	return mux
}

func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	gmServer().ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
	return rec
}

func TestListEndpoint(t *testing.T) {
	rec := get(t, "/api/campaigns/c1/map-templates")
	var out []Info
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &out) != nil || len(out) < 2 {
		t.Fatalf("list: %d %s", rec.Code, rec.Body.String())
	}
	if out[0].ID != "vila-abandonada" || out[0].Width != 64 || out[0].Ambience.TimeOfDay != "dusk" || len(out[0].Tags) == 0 {
		t.Errorf("first: %+v", out[0])
	}
}

func TestPreviewEndpoint(t *testing.T) {
	rec := get(t, "/api/campaigns/c1/map-templates/rodovia-engarrafada/preview?seed=5")
	var p Preview
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &p) != nil {
		t.Fatalf("preview: %d", rec.Code)
	}
	g, err := maps.Deserialize(p.Grid)
	if err != nil || g.Width != 96 || g.Depth != 64 || len(p.Objects) == 0 {
		t.Fatalf("payload: %v %dx%d objs=%d", err, g.Width, g.Depth, len(p.Objects))
	}
	other := get(t, "/api/campaigns/c1/map-templates/rodovia-engarrafada/preview?seed=6")
	if other.Body.String() == rec.Body.String() {
		t.Error("another seed must change the preview")
	}
	if rec.Body.Len() > 120<<10 {
		t.Errorf("preview payload too large: %d bytes", rec.Body.Len())
	}
}

func TestPreviewErrors(t *testing.T) {
	if c := get(t, "/api/campaigns/c1/map-templates/nope/preview").Code; c != 404 {
		t.Errorf("unknown template: %d", c)
	}
	if c := get(t, "/api/campaigns/c1/map-templates/vila-abandonada/preview?seed=abc").Code; c != 400 {
		t.Errorf("bad seed: %d", c)
	}
}
