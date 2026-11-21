package ambience

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

// Register wires the commands ambience.set (Mestre), condition.set, chat.say and chat.react and
// GET /api/campaigns/{id}/scenes/{sceneId}/ambience.
func Register(mux *http.ServeMux, hub *realtime.Hub, svc *Service, cmp *campaigns.Handlers) {
	RegisterCommands(hub, svc)
	mux.Handle("GET /api/campaigns/{id}/scenes/{sceneId}/ambience", cmp.RequireMember(getHandler(svc)))
}

// RegisterCommands wires the realtime commands only.
func RegisterCommands(hub *realtime.Hub, svc *Service) {
	hub.Handle("ambience.set", decode(func(ctx context.Context, c realtime.Command, in Ambience) error {
		return svc.Set(ctx, c.CampaignID, in)
	}), realtime.GMOnly)
	hub.Handle("condition.set", decode(func(ctx context.Context, c realtime.Command, in struct {
		CharacterID string `json:"characterId"`
		Condition   string `json:"condition"`
		Active      bool   `json:"active"`
	}) error {
		return svc.SetCondition(ctx, c.Sender, c.CampaignID, in.CharacterID, in.Condition, in.Active)
	}))
	hub.Handle("chat.say", decode(func(ctx context.Context, c realtime.Command, in struct {
		CharacterID string `json:"characterId"`
		Text        string `json:"text"`
	}) error {
		return svc.Say(ctx, c.Sender, c.CampaignID, in.CharacterID, in.Text)
	}))
	hub.Handle("chat.react", decode(func(ctx context.Context, c realtime.Command, in struct {
		CharacterID string `json:"characterId"`
		Emoji       string `json:"emoji"`
	}) error {
		return svc.React(ctx, c.Sender, c.CampaignID, in.CharacterID, in.Emoji)
	}))
}

func decode[T any](fn func(context.Context, realtime.Command, T) error) realtime.HandlerFunc {
	return func(ctx context.Context, cmd realtime.Command) error {
		var in T
		if err := json.Unmarshal(cmd.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		return fn(ctx, cmd, in)
	}
}

func getHandler(svc *Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		a, err := svc.Get(r.Context(), m.CampaignID, r.PathValue("sceneId"))
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, scenes.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Cena não encontrada"})
		case err != nil:
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Erro interno. Tente novamente"})
		default:
			_ = json.NewEncoder(w).Encode(a)
		}
	})
}
