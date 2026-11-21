package scenes

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

// Handlers exposes the scene REST endpoints.
type Handlers struct {
	svc *Service
	cmp *campaigns.Handlers
}

// NewHandlers builds the HTTP layer using the campaign membership middlewares.
func NewHandlers(svc *Service, cmp *campaigns.Handlers) *Handlers {
	return &Handlers{svc: svc, cmp: cmp}
}

// Register mounts the routes on mux.
func (h *Handlers) Register(mux *http.ServeMux) {
	const base = "/api/campaigns/{id}/scenes"
	mux.Handle("GET "+base, h.cmp.RequireMember(http.HandlerFunc(h.list)))
	mux.Handle("POST "+base, h.cmp.RequireGM(http.HandlerFunc(h.create)))
	mux.Handle("GET "+base+"/{sceneId}", h.cmp.RequireMember(http.HandlerFunc(h.get)))
	mux.Handle("PATCH "+base+"/{sceneId}", h.cmp.RequireGM(http.HandlerFunc(h.rename)))
	mux.Handle("DELETE "+base+"/{sceneId}", h.cmp.RequireGM(http.HandlerFunc(h.remove)))
	mux.Handle("POST "+base+"/{sceneId}/activate", h.cmp.RequireGM(http.HandlerFunc(h.activate)))
	h.registerSnapshots(mux, base)
}

func (h *Handlers) list(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.List(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handlers) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name       string `json:"name"`
		Width      int    `json:"width"`
		Depth      int    `json:"depth"`
		TemplateID string `json:"templateId"`
		Seed       int64  `json:"seed"`
	}
	if !decode(w, r, &in) {
		return
	}
	campaignID := r.PathValue("id")
	var sc Scene
	var err error
	if in.TemplateID != "" {
		sc, err = h.svc.CreateFromTemplate(r.Context(), campaignID, in.Name, in.TemplateID, in.Seed)
	} else {
		sc, err = h.svc.Create(r.Context(), campaignID, in.Name, in.Width, in.Depth)
	}
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sc)
}

func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {
	m, _ := campaigns.MemberFrom(r)
	sc, err := h.svc.Get(r.Context(), m.CampaignID, r.PathValue("sceneId"), m.Role)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sc)
}

func (h *Handlers) rename(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &in) {
		return
	}
	h.done(w, h.svc.Rename(r.Context(), r.PathValue("id"), r.PathValue("sceneId"), in.Name))
}

func (h *Handlers) remove(w http.ResponseWriter, r *http.Request) {
	h.done(w, h.svc.Delete(r.Context(), r.PathValue("id"), r.PathValue("sceneId")))
}

func (h *Handlers) activate(w http.ResponseWriter, r *http.Request) {
	h.done(w, h.svc.Activate(r.Context(), r.PathValue("id"), r.PathValue("sceneId")))
}

func (h *Handlers) done(w http.ResponseWriter, err error) {
	if err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida")
		return false
	}
	return true
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "Cena não encontrada")
	case errors.Is(err, ErrTooManySnapshots):
		writeError(w, http.StatusConflict, "Limite de estados salvos desta cena atingido")
	case errors.Is(err, ErrInvalidName):
		writeError(w, http.StatusBadRequest, "Informe um nome para a cena (até 80 caracteres)")
	case errors.Is(err, ErrUnknownTemplate):
		writeError(w, http.StatusBadRequest, "Modelo de mapa não encontrado")
	case errors.Is(err, ErrInvalidSize):
		writeError(w, http.StatusBadRequest, "O tamanho da cena deve ser de 16 a 128 células por lado")
	default:
		slog.Error("scenes request failed", "err", err)
		writeError(w, http.StatusInternalServerError, "Erro interno. Tente novamente")
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
