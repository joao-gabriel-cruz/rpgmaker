package items

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

// RegisterHTTP mounts the Mestre-only compendium CRUD and the role-filtered loot list of a scene.
func RegisterHTTP(mux *http.ServeMux, svc *Service, h *campaigns.Handlers) {
	base := "/api/campaigns/{id}/items"
	mux.Handle("GET "+base, h.RequireGM(serve(func(r *http.Request, m campaigns.Member) (any, int, error) {
		list, err := svc.List(r.Context(), m.CampaignID)
		return map[string]any{"items": list}, http.StatusOK, err
	})))
	mux.Handle("POST "+base, h.RequireGM(serve(func(r *http.Request, m campaigns.Member) (any, int, error) {
		in, err := decode(r)
		if err != nil {
			return nil, 0, err
		}
		out, err := svc.Create(r.Context(), m.CampaignID, in)
		return out, http.StatusCreated, err
	})))
	mux.Handle("PUT "+base+"/{itemId}", h.RequireGM(serve(func(r *http.Request, m campaigns.Member) (any, int, error) {
		in, err := decode(r)
		if err != nil {
			return nil, 0, err
		}
		out, err := svc.Update(r.Context(), m.CampaignID, r.PathValue("itemId"), in)
		return out, http.StatusOK, err
	})))
	mux.Handle("DELETE "+base+"/{itemId}", h.RequireGM(serve(func(r *http.Request, m campaigns.Member) (any, int, error) {
		return nil, http.StatusNoContent, svc.Delete(r.Context(), m.CampaignID, r.PathValue("itemId"))
	})))
	mux.Handle("GET /api/campaigns/{id}/scenes/{sceneId}/loot", h.RequireMember(serve(func(r *http.Request, m campaigns.Member) (any, int, error) {
		list, err := svc.ListLoot(r.Context(), m, r.PathValue("sceneId"))
		return map[string]any{"loot": list}, http.StatusOK, err
	})))
}

func decode(r *http.Request) (Item, error) {
	var in Item
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
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Item não encontrado"})
		default:
			slog.Error("items http failed", "err", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro interno. Tente novamente"})
		}
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
