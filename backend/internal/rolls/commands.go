package rolls

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// Registrar is the part of the realtime hub used to register commands.
type Registrar interface {
	Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard)
}

// RegisterCommands wires roll.request, roll.keep and roll.generic.
func RegisterCommands(r Registrar, svc *Service) {
	handle(r, "roll.request", func(ctx context.Context, cmd realtime.Command, in RequestInput) error {
		return svc.Request(ctx, cmd.CampaignID, cmd.Sender, in)
	})
	handle(r, "roll.keep", func(ctx context.Context, cmd realtime.Command, in struct {
		RollID string `json:"rollId"`
		Kept   []int  `json:"kept"`
	}) error {
		return svc.Keep(ctx, cmd.CampaignID, cmd.Sender, in.RollID, in.Kept)
	})
	r.Handle("roll.difficulty.set", handler(func(ctx context.Context, cmd realtime.Command, in struct {
		CharacterID string `json:"characterId"`
		D12         int    `json:"d12"`
	}) error {
		return svc.SetDifficulty(ctx, cmd.CampaignID, cmd.Sender, in.CharacterID, in.D12)
	}), realtime.GMOnly)
	handle(r, "roll.generic", func(ctx context.Context, cmd realtime.Command, in struct {
		Expression string `json:"expression"`
	}) error {
		return svc.Generic(ctx, cmd.CampaignID, cmd.Sender, in.Expression)
	})
}

func handle[T any](r Registrar, typ string, fn func(context.Context, realtime.Command, T) error) {
	r.Handle(typ, handler(fn))
}

func handler[T any](fn func(context.Context, realtime.Command, T) error) realtime.HandlerFunc {
	return func(ctx context.Context, cmd realtime.Command) error {
		var in T
		if err := json.Unmarshal(cmd.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		return commandError(fn(ctx, cmd, in))
	}
}

// commandError turns domain errors into messages shown to the roller.
func commandError(err error) error {
	var ruleA assimilacao.RuleError
	var ruleG RuleError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &ruleA), errors.As(err, &ruleG):
		return realtime.UserError(err.Error())
	case errors.Is(err, ErrNotFound):
		return realtime.UserError("Rolagem não encontrada.")
	case errors.Is(err, ErrAlreadyFinal):
		return realtime.UserError("Esta rolagem já foi finalizada.")
	case errors.Is(err, characters.ErrNotFound), errors.Is(err, campaigns.ErrNotFound):
		return realtime.UserError("Personagem não encontrado.")
	}
	return err
}

// RegisterHTTP mounts GET /api/campaigns/{id}/rolls.
func RegisterHTTP(mux *http.ServeMux, svc *Service, requireMember func(http.Handler) http.Handler) {
	mux.Handle("GET /api/campaigns/{id}/rolls/state", requireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		st, err := svc.State(r.Context(), m, r.URL.Query().Get("characterId"))
		writeJSON(w, st, err)
	})))
	mux.Handle("GET /api/campaigns/{id}/rolls", requireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		out, err := svc.Recent(r.Context(), m, limit)
		writeJSON(w, out, err)
	})))
}

func writeJSON(w http.ResponseWriter, v any, err error) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case errors.Is(err, characters.ErrNotFound), errors.Is(err, campaigns.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Personagem não encontrado."}`))
	case err != nil:
		slog.Error("rolls request failed", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"Erro interno. Tente novamente"}`))
	default:
		_ = json.NewEncoder(w).Encode(v)
	}
}
