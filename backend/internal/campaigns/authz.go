package campaigns

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
)

type memberKey struct{}

// MemberFrom returns the membership injected by RequireMember or RequireGM.
func MemberFrom(r *http.Request) (Member, bool) {
	m, ok := r.Context().Value(memberKey{}).(Member)
	return m, ok
}

// RequireMember authenticates, resolves the {id} route value and injects the caller's Member.
// Non-members get 404 so campaign existence is not leaked.
func (h *Handlers) RequireMember(next http.Handler) http.Handler {
	return h.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := auth.UserFrom(r)
		m, err := h.svc.Membership(r.Context(), r.PathValue("id"), u.ID)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}
		if err != nil {
			serverError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), memberKey{}, m)))
	}))
}

// RequireGM is RequireMember plus a gm-role check (403 for players).
func (h *Handlers) RequireGM(next http.Handler) http.Handler {
	return h.RequireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m, _ := MemberFrom(r); m.Role != RoleGM {
			writeError(w, http.StatusForbidden, "Apenas o Mestre pode fazer isso")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func serverError(w http.ResponseWriter, err error) {
	slog.Error("campaigns request failed", "err", err)
	writeError(w, http.StatusInternalServerError, "Erro interno. Tente novamente")
}
