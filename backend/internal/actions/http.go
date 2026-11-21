package actions

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

// RegisterHTTP mounts the catalog: GET for members (the Mestre also sees disabled ones) and Mestre-only edits.
func RegisterHTTP(mux *http.ServeMux, svc *Service, h *campaigns.Handlers) {
	base := "/api/campaigns/{id}/actions"
	mux.Handle("GET "+base, h.RequireMember(serve(func(r *http.Request, m campaigns.Member) (any, int, error) {
		list, err := svc.Catalog(r.Context(), m.CampaignID, m.Role == campaigns.RoleGM)
		return map[string]any{"actions": list}, http.StatusOK, err
	})))
	mux.Handle("POST "+base, h.RequireGM(serve(func(r *http.Request, m campaigns.Member) (any, int, error) {
		in, err := decode(r)
		if err != nil {
			return nil, 0, err
		}
		out, err := svc.Create(r.Context(), m.CampaignID, in)
		return out, http.StatusCreated, err
	})))
	mux.Handle("PUT "+base+"/{actionId}", h.RequireGM(serve(func(r *http.Request, m campaigns.Member) (any, int, error) {
		in, err := decode(r)
		if err != nil {
			return nil, 0, err
		}
		out, err := svc.Update(r.Context(), m.CampaignID, r.PathValue("actionId"), in)
		return out, http.StatusOK, err
	})))
	mux.Handle("DELETE "+base+"/{actionId}", h.RequireGM(serve(func(r *http.Request, m campaigns.Member) (any, int, error) {
		return nil, http.StatusNoContent, svc.Delete(r.Context(), m.CampaignID, r.PathValue("actionId"))
	})))
}

func decode(r *http.Request) (Action, error) {
	var in Action
	if err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<16)).Decode(&in); err != nil {
		return in, bad("Corpo inválido")
	}
	return in, nil
}

func serve(fn func(*http.Request, campaigns.Member) (any, int, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		v, status, err := fn(r, m)
		var in inputError
		switch {
		case err == nil && status == http.StatusNoContent:
			w.WriteHeader(status)
		case err == nil:
			writeJSON(w, status, v)
		case errors.As(err, &in):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Ação não encontrada"})
		default:
			slog.Error("actions http failed", "err", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro interno. Tente novamente"})
		}
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
