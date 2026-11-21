package campaigns

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
)

// Handlers exposes the campaign, member and invite endpoints.
type Handlers struct {
	svc         *Service
	requireAuth func(http.Handler) http.Handler
}

// NewHandlers builds the HTTP layer; requireAuth is auth.Handlers.RequireAuth.
func NewHandlers(svc *Service, requireAuth func(http.Handler) http.Handler) *Handlers {
	return &Handlers{svc: svc, requireAuth: requireAuth}
}

// Register mounts the routes on mux.
func (h *Handlers) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/campaigns", h.requireAuth(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/campaigns", h.requireAuth(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/campaigns/{id}", h.RequireMember(http.HandlerFunc(h.detail)))
	mux.Handle("PATCH /api/campaigns/{id}/settings", h.RequireGM(http.HandlerFunc(h.updateSettings)))
	mux.Handle("GET /api/campaigns/{id}/members", h.RequireMember(http.HandlerFunc(h.members)))
	mux.Handle("DELETE /api/campaigns/{id}/members/{userId}", h.RequireGM(http.HandlerFunc(h.removeMember)))
	mux.Handle("POST /api/campaigns/{id}/invites", h.RequireGM(http.HandlerFunc(h.createInvite)))
	mux.HandleFunc("GET /api/invites/{token}", h.invite)
	mux.HandleFunc("POST /api/invites/{token}/signup", h.signUp)
	mux.Handle("POST /api/invites/{token}/accept", h.requireAuth(http.HandlerFunc(h.accept)))
}

func (h *Handlers) create(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name, SystemID string }
	if !decode(w, r, &in) {
		return
	}
	u, _ := auth.UserFrom(r)
	c, err := h.svc.Create(r.Context(), u.ID, in.Name, in.SystemID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, Summary{Campaign: c, Role: RoleGM, SystemName: h.svc.systemName(c.SystemID)})
}

func (h *Handlers) list(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r)
	list, err := h.svc.List(r.Context(), u.ID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handlers) detail(w http.ResponseWriter, r *http.Request) {
	m, _ := MemberFrom(r)
	d, err := h.svc.Detail(r.Context(), m)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *Handlers) updateSettings(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CreationLimits string `json:"creationLimits"`
	}
	if !decode(w, r, &in) {
		return
	}
	m, _ := MemberFrom(r)
	d, err := h.svc.UpdateSettings(r.Context(), m, in.CreationLimits)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *Handlers) members(w http.ResponseWriter, r *http.Request) {
	m, _ := MemberFrom(r)
	list, err := h.svc.Members(r.Context(), m.CampaignID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handlers) removeMember(w http.ResponseWriter, r *http.Request) {
	m, _ := MemberFrom(r)
	if err := h.svc.RemoveMember(r.Context(), m, r.PathValue("userId")); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) createInvite(w http.ResponseWriter, r *http.Request) {
	m, _ := MemberFrom(r)
	inv, err := h.svc.CreateInvite(r.Context(), m.CampaignID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, inv)
}

func (h *Handlers) invite(w http.ResponseWriter, r *http.Request) {
	inv, err := h.svc.Invite(r.Context(), r.PathValue("token"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func (h *Handlers) signUp(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, DisplayName, Password string }
	if !decode(w, r, &in) {
		return
	}
	u, token, expires, err := h.svc.SignUp(r.Context(), r.PathValue("token"), in.Email, in.DisplayName, in.Password)
	if err != nil {
		fail(w, err)
		return
	}
	auth.SetSessionCookie(w, r, token, expires)
	writeJSON(w, http.StatusCreated, u)
}

func (h *Handlers) accept(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r)
	m, err := h.svc.Accept(r.Context(), r.PathValue("token"), u.ID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
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
	case errors.Is(err, ErrInvalidInvite):
		writeError(w, http.StatusNotFound, "Convite inválido ou expirado")
	case errors.Is(err, ErrAlreadyMember):
		writeError(w, http.StatusConflict, "Você já participa desta campanha")
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "Preencha os campos corretamente (senha com no mínimo 8 caracteres)")
	case errors.Is(err, ErrUnknownSystem):
		writeError(w, http.StatusBadRequest, "Sistema de jogo inválido")
	case errors.Is(err, ErrCannotRemove):
		writeError(w, http.StatusBadRequest, "Não é possível remover este membro")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "Não encontrado")
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, http.StatusConflict, "E-mail já cadastrado. Entre na sua conta e aceite o convite")
	default:
		serverError(w, err)
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
