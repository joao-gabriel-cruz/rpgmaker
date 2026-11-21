package characters

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

// Handlers exposes the character endpoints.
type Handlers struct {
	svc           *Service
	requireMember func(http.Handler) http.Handler
}

// NewHandlers builds the HTTP layer; requireMember is campaigns.Handlers.RequireMember.
func NewHandlers(svc *Service, requireMember func(http.Handler) http.Handler) *Handlers {
	return &Handlers{svc: svc, requireMember: requireMember}
}

// Register mounts the routes on mux.
func (h *Handlers) Register(mux *http.ServeMux) {
	const base = "/api/campaigns/{id}/characters"
	mux.Handle("GET "+base, h.requireMember(http.HandlerFunc(h.list)))
	mux.Handle("POST "+base, h.requireMember(http.HandlerFunc(h.create)))
	mux.Handle("GET "+base+"/{charId}", h.requireMember(http.HandlerFunc(h.get)))
	mux.Handle("PUT "+base+"/{charId}", h.requireMember(http.HandlerFunc(h.save)))
	mux.Handle("DELETE "+base+"/{charId}", h.requireMember(http.HandlerFunc(h.delete)))
}

func (h *Handlers) list(w http.ResponseWriter, r *http.Request) {
	m, _ := campaigns.MemberFrom(r)
	out, err := h.svc.List(r.Context(), m)
	respond(w, http.StatusOK, out, err)
}

func (h *Handlers) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		OwnerID string `json:"ownerId"`
		Name    string `json:"name"`
		Kind    string `json:"kind"`
	}
	if r.ContentLength != 0 && !decode(w, r, &in) {
		return
	}
	m, _ := campaigns.MemberFrom(r)
	out, err := h.svc.Create(r.Context(), m, CreateInput{OwnerID: in.OwnerID, Name: in.Name, Kind: in.Kind})
	respond(w, http.StatusCreated, out, err)
}

func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {
	m, _ := campaigns.MemberFrom(r)
	out, err := h.svc.Get(r.Context(), m, r.PathValue("charId"))
	respond(w, http.StatusOK, out, err)
}

func (h *Handlers) save(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Sheet json.RawMessage `json:"sheet"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.Sheet) == 0 || in.Sheet[0] != '{' {
		writeError(w, http.StatusBadRequest, "Requisição inválida")
		return
	}
	m, _ := campaigns.MemberFrom(r)
	out, err := h.svc.Save(r.Context(), m, r.PathValue("charId"), in.Sheet)
	respond(w, http.StatusOK, out, err)
}

func (h *Handlers) delete(w http.ResponseWriter, r *http.Request) {
	m, _ := campaigns.MemberFrom(r)
	if err := h.svc.Delete(r.Context(), m, r.PathValue("charId")); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func respond(w http.ResponseWriter, status int, v any, err error) {
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, status, v)
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida")
		return false
	}
	return true
}

func fail(w http.ResponseWriter, err error) {
	var rej *RejectedError
	switch {
	case errors.As(err, &rej):
		writeJSON(w, http.StatusBadRequest, rejection{Error: "A ficha não pôde ser salva: corrija os pontos indicados", Issues: rej.Issues})
	case errors.Is(err, ErrNotFound), errors.Is(err, campaigns.ErrNotFound):
		writeError(w, http.StatusNotFound, "Personagem não encontrado")
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "Você não pode alterar este personagem")
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "Dados inválidos")
	case errors.Is(err, ErrNoSheetSupport):
		writeError(w, http.StatusBadRequest, "Este sistema não suporta fichas")
	default:
		slog.Error("characters request failed", "err", err)
		writeError(w, http.StatusInternalServerError, "Erro interno. Tente novamente")
	}
}

type rejection struct {
	Error  string          `json:"error"`
	Issues []systems.Issue `json:"issues"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
