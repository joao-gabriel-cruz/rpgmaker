package effects

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rounds"
)

// Registrar is the part of the realtime hub used to register commands.
type Registrar interface {
	Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard)
}

// RegisterCommands wires the Mestre-only effect.confirm and effect.dismiss.
func RegisterCommands(r Registrar, svc *Service) {
	idCmd := func(fn func(context.Context, realtime.Command, string) error) realtime.HandlerFunc {
		return func(ctx context.Context, c realtime.Command) error {
			var in struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(c.Payload, &in); err != nil {
				return realtime.UserError("Payload inválido")
			}
			return fn(ctx, c, in.ID)
		}
	}
	r.Handle("effect.confirm", idCmd(func(ctx context.Context, c realtime.Command, id string) error {
		return svc.Confirm(ctx, c.CampaignID, c.Sender, id)
	}), realtime.GMOnly)
	r.Handle("effect.dismiss", idCmd(func(ctx context.Context, c realtime.Command, id string) error {
		return svc.Dismiss(ctx, c.CampaignID, id)
	}), realtime.GMOnly)
}

// StartOfTurn is the rounds hook: Preparação of a new round (conflict start or the step after Fim) opens the turn.
func (s *Service) StartOfTurn(ctx context.Context, campaignID string, prev, next rounds.State) {
	if next.Active && next.Phase == rounds.PhasePreparation && (!prev.Active || next.Round != prev.Round) {
		s.OnTurnStart(ctx, campaignID)
	}
}

// RegisterHTTP mounts the Mestre-only REST: GET rules, PUT/DELETE one rule set, GET pending suggestions.
func RegisterHTTP(mux *http.ServeMux, svc *Service, cmp *campaigns.Handlers) {
	base := "/api/campaigns/{id}/effects"
	mux.Handle("GET "+base+"/rules", cmp.RequireGM(handler(func(r *http.Request, m campaigns.Member) (any, error) {
		list, err := svc.Rules(r.Context(), m.CampaignID)
		return map[string]any{"rules": list}, err
	})))
	mux.Handle("PUT "+base+"/rules/{kind}/{key}", cmp.RequireGM(handler(func(r *http.Request, m campaigns.Member) (any, error) {
		var in struct {
			Rules []Rule `json:"rules"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<16)).Decode(&in); err != nil {
			return nil, badInput("Corpo inválido")
		}
		return svc.SetRules(r.Context(), m.CampaignID, r.PathValue("kind"), r.PathValue("key"), in.Rules)
	})))
	mux.Handle("DELETE "+base+"/rules/{kind}/{key}", cmp.RequireGM(handler(func(r *http.Request, m campaigns.Member) (any, error) {
		return svc.RestoreRules(r.Context(), m.CampaignID, r.PathValue("kind"), r.PathValue("key"))
	})))
	mux.Handle("GET "+base+"/pending", cmp.RequireGM(handler(func(r *http.Request, m campaigns.Member) (any, error) {
		list, err := svc.Pending(r.Context(), m.CampaignID)
		return map[string]any{"suggestions": list}, err
	})))
}

func handler(fn func(*http.Request, campaigns.Member) (any, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		v, err := fn(r, m)
		w.Header().Set("Content-Type", "application/json")
		var bad inputError
		switch {
		case err == nil:
			_ = json.NewEncoder(w).Encode(v)
		case errors.As(err, &bad):
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		default:
			slog.Error("effects http failed", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Erro interno. Tente novamente"})
		}
	})
}
