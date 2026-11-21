package mcp

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

// RegisterLinkHTTP mounts the GM-only link endpoints.
func RegisterLinkHTTP(mux *http.ServeMux, l *Links, cmp *campaigns.Handlers) {
	const path = "/api/campaigns/{id}/mcp-link"
	h := &linkHandlers{links: l}
	mux.Handle("GET "+path, cmp.RequireGM(http.HandlerFunc(h.status)))
	mux.Handle("POST "+path, cmp.RequireGM(http.HandlerFunc(h.generate)))
	mux.Handle("DELETE "+path, cmp.RequireGM(http.HandlerFunc(h.revoke)))
}

type linkHandlers struct{ links *Links }

func (h *linkHandlers) status(w http.ResponseWriter, r *http.Request) {
	m, _ := campaigns.MemberFrom(r)
	link, ok, err := h.links.Status(r.Context(), m)
	if err != nil {
		failLink(w, err)
		return
	}
	out := map[string]any{"exists": ok}
	if ok {
		out["createdAt"] = link.CreatedAt
		if link.LastUsedAt != nil {
			out["lastUsedAt"] = link.LastUsedAt
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *linkHandlers) generate(w http.ResponseWriter, r *http.Request) {
	m, _ := campaigns.MemberFrom(r)
	secret, link, err := h.links.Generate(r.Context(), m)
	if err != nil {
		failLink(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, struct {
		URL       string    `json:"url"`
		CreatedAt time.Time `json:"createdAt"`
	}{requestScheme(r) + "://" + r.Host + "/api/mcp/" + secret, link.CreatedAt})
}

func (h *linkHandlers) revoke(w http.ResponseWriter, r *http.Request) {
	m, _ := campaigns.MemberFrom(r)
	if err := h.links.Revoke(r.Context(), m); err != nil {
		failLink(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func requestScheme(r *http.Request) string {
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		return p
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func failLink(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrForbidden) {
		writeError(w, http.StatusForbidden, "Apenas o Mestre pode gerenciar este link")
		return
	}
	slog.Error("mcp link", "err", err)
	writeError(w, http.StatusInternalServerError, "Erro interno")
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
