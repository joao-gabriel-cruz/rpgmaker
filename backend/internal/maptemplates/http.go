package maptemplates

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// Preview is the payload that draws a gallery thumbnail.
type Preview struct {
	Grid     maps.Serialized `json:"grid"`
	Objects  []maps.Object   `json:"objects"`
	Ambience Ambience        `json:"ambience"`
}

// Build generates the template for a seed, ready to send or persist.
func (t Template) Build(seed int64) Preview {
	g, objs := t.Generate(seed)
	if objs == nil {
		objs = []maps.Object{}
	}
	return Preview{Grid: g.Serialize(), Objects: objs, Ambience: t.Ambience}
}

// RegisterHTTP mounts the Mestre-only gallery endpoints; requireGM is campaigns.Handlers.RequireGM.
func RegisterHTTP(mux *http.ServeMux, requireGM func(http.Handler) http.Handler) {
	const base = "/api/campaigns/{id}/map-templates"
	mux.Handle("GET "+base, requireGM(http.HandlerFunc(list)))
	mux.Handle("GET "+base+"/{tid}/preview", requireGM(http.HandlerFunc(preview)))
}

func list(w http.ResponseWriter, _ *http.Request) {
	all := All()
	out := make([]Info, len(all))
	for i, t := range all {
		out[i] = t.Info()
	}
	writeJSON(w, http.StatusOK, out)
}

func preview(w http.ResponseWriter, r *http.Request) {
	t, ok := Get(r.PathValue("tid"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Modelo de mapa não encontrado"})
		return
	}
	seed := int64(1)
	if raw := r.URL.Query().Get("seed"); raw != "" {
		var err error
		if seed, err = strconv.ParseInt(raw, 10, 64); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Semente inválida"})
			return
		}
	}
	writeJSON(w, http.StatusOK, t.Build(seed))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
