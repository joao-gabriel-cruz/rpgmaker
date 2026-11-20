package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// CookieName is the session cookie.
const CookieName = "rpg_session"

type ctxKey struct{}

// Handlers exposes the auth and admin endpoints.
type Handlers struct{ svc *Service }

// NewHandlers builds the HTTP layer over a Service.
func NewHandlers(svc *Service) *Handlers { return &Handlers{svc: svc} }

// Register mounts the routes on mux.
func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.Handle("POST /api/auth/logout", h.RequireAuth(http.HandlerFunc(h.logout)))
	mux.Handle("GET /api/auth/me", h.RequireAuth(http.HandlerFunc(h.me)))
	mux.Handle("GET /api/admin/users", h.RequireAdmin(http.HandlerFunc(h.listUsers)))
	mux.Handle("POST /api/admin/users/{id}/reset-password", h.RequireAdmin(http.HandlerFunc(h.resetPassword)))
}

// UserFrom returns the authenticated user injected by RequireAuth.
func UserFrom(r *http.Request) (User, bool) {
	u, ok := r.Context().Value(ctxKey{}).(User)
	return u, ok
}

// RequireAuth rejects requests without a valid session and injects the user into the context.
func (h *Handlers) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(CookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Faça login para continuar")
			return
		}
		u, err := h.svc.Authenticate(r.Context(), c.Value)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				slog.Error("authenticate", "err", err)
			}
			writeError(w, http.StatusUnauthorized, "Faça login para continuar")
			return
		}
		next.ServeHTTP(w, r.WithContext(contextWithUser(r, u)))
	})
}

// RequireAdmin is RequireAuth plus an admin check.
func (h *Handlers) RequireAdmin(next http.Handler) http.Handler {
	return h.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, _ := UserFrom(r); !u.IsAdmin {
			writeError(w, http.StatusForbidden, "Acesso restrito a administradores")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "Requisição inválida")
		return
	}
	u, token, expires, err := h.svc.Login(r.Context(), in.Email, in.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "E-mail ou senha inválidos")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	http.SetCookie(w, sessionCookie(token, expires, isSecure(r)))
	writeJSON(w, http.StatusOK, u)
}

func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie(CookieName)
	if err := h.svc.Logout(r.Context(), c.Value); err != nil {
		serverError(w, err)
		return
	}
	http.SetCookie(w, sessionCookie("", time.Unix(0, 0), isSecure(r)))
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) me(w http.ResponseWriter, r *http.Request) {
	u, _ := UserFrom(r)
	writeJSON(w, http.StatusOK, u)
}

func (h *Handlers) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handlers) resetPassword(w http.ResponseWriter, r *http.Request) {
	temp, err := h.svc.ResetPassword(r.Context(), r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "Usuário não encontrado")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"temporaryPassword": temp})
}

func sessionCookie(value string, expires time.Time, secure bool) *http.Cookie {
	//nolint:gosec // Secure is set per request: TLS or X-Forwarded-Proto https
	c := &http.Cookie{Name: CookieName, Value: value, Path: "/", Expires: expires, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: secure}
	if value == "" {
		c.MaxAge = -1
	}
	return c
}

func serverError(w http.ResponseWriter, err error) {
	slog.Error("auth request failed", "err", err)
	writeError(w, http.StatusInternalServerError, "Erro interno. Tente novamente")
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func contextWithUser(r *http.Request, u User) context.Context {
	return context.WithValue(r.Context(), ctxKey{}, u)
}

func isSecure(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}
