package vision

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

func bg() context.Context { return context.Background() }

// RegisterHTTP mounts GET /api/campaigns/{id}/scenes/{sceneId}/vision.
func RegisterHTTP(mux *http.ServeMux, svc *Service, cmp *campaigns.Handlers) {
	mux.Handle("GET /api/campaigns/{id}/scenes/{sceneId}/vision", cmp.RequireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		v, err := svc.Get(r.Context(), m, r.PathValue("sceneId"))
		switch {
		case errors.Is(err, ErrNotFound):
			write(w, http.StatusNotFound, map[string]string{"error": "Cena não encontrada"})
		case err != nil:
			slog.Error("get vision failed", "err", err)
			write(w, http.StatusInternalServerError, map[string]string{"error": "Erro interno. Tente novamente"})
		default:
			write(w, http.StatusOK, v)
		}
	})))
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
