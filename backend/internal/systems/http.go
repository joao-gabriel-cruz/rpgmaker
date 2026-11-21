package systems

import (
	"encoding/json"
	"net/http"
)

type systemDTO struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Register mounts GET /api/systems behind the given auth middleware.
func Register(mux *http.ServeMux, reg *Registry, requireAuth func(http.Handler) http.Handler) {
	mux.Handle("GET /api/systems", requireAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		out := []systemDTO{}
		for _, s := range reg.List() {
			out = append(out, systemDTO{s.ID(), s.Name(), s.Version()})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})))
}
