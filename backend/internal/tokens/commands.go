package tokens

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Registrar is the part of the realtime hub used to register commands.
type Registrar interface {
	Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard)
}

type placeIn struct {
	SceneID     string `json:"sceneId"`
	CharacterID string `json:"characterId"`
	X           int    `json:"x"`
	Z           int    `json:"z"`
}

type moveIn struct {
	TokenID string `json:"tokenId"`
	X       int    `json:"x"`
	Z       int    `json:"z"`
	Force   bool   `json:"force"`
}

type pointIn struct {
	SceneID string  `json:"sceneId"`
	X       float64 `json:"x"`
	Z       float64 `json:"z"`
}

// RegisterCommands wires token.place, token.remove, token.move, map.ping and camera.focusAll.
func RegisterCommands(r Registrar, svc *Service) {
	handle(r, "token.place", func(ctx context.Context, c realtime.Command, in placeIn) error {
		return svc.Place(ctx, c.Sender, c.CampaignID, in.SceneID, in.CharacterID, in.X, in.Z)
	})
	handle(r, "token.remove", func(ctx context.Context, c realtime.Command, in struct {
		TokenID string `json:"tokenId"`
	}) error {
		return svc.Remove(ctx, c.Sender, c.CampaignID, in.TokenID)
	}, realtime.GMOnly)
	handle(r, "token.move", func(ctx context.Context, c realtime.Command, in moveIn) error {
		return svc.Move(ctx, c.Sender, c.CampaignID, in.TokenID, in.X, in.Z, in.Force)
	})
	handle(r, "map.ping", func(_ context.Context, c realtime.Command, in pointIn) error {
		return svc.Ping(c.CampaignID, c.Sender, in.SceneID, in.X, in.Z)
	})
	handle(r, "camera.focusAll", func(_ context.Context, c realtime.Command, in pointIn) error {
		return svc.FocusAll(c.CampaignID, c.Sender, in.SceneID, in.X, in.Z)
	}, realtime.GMOnly)
}

func handle[T any](r Registrar, typ string, fn func(context.Context, realtime.Command, T) error, guards ...realtime.Guard) {
	r.Handle(typ, func(ctx context.Context, cmd realtime.Command) error {
		var in T
		if err := json.Unmarshal(cmd.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		return fn(ctx, cmd, in)
	}, guards...)
}
