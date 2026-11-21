package tokens

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

// Register mounts GET /api/campaigns/{id}/scenes/{sceneId}/tokens.
func Register(mux *http.ServeMux, svc *Service, cmp *campaigns.Handlers) {
	mux.Handle("GET /api/campaigns/{id}/scenes/{sceneId}/tokens", cmp.RequireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		out, err := svc.List(r.Context(), m, r.PathValue("sceneId"))
		switch {
		case errors.Is(err, ErrNotFound):
			write(w, http.StatusNotFound, map[string]string{"error": "Cena não encontrada"})
		case err != nil:
			slog.Error("list tokens failed", "err", err)
			write(w, http.StatusInternalServerError, map[string]string{"error": "Erro interno. Tente novamente"})
		default:
			write(w, http.StatusOK, out)
		}
	})))
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
