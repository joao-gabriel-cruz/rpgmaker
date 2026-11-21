package items

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Registrar is the part of the realtime hub used to register commands.
type Registrar interface {
	Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard)
}

func parse[T any](c realtime.Command) (T, error) {
	var p T
	if err := json.Unmarshal(c.Payload, &p); err != nil {
		return p, realtime.UserError("Payload inválido")
	}
	return p, nil
}

func memberOf(c realtime.Command) campaigns.Member {
	return campaigns.Member{CampaignID: c.CampaignID, UserID: c.Sender.UserID, Role: c.Sender.Role}
}

type lootRef struct {
	LootID      string `json:"lootId"`
	CharacterID string `json:"characterId"`
}

type useCmd struct {
	CharacterID string `json:"characterId"`
	Place       string `json:"place"`
	Index       int    `json:"index"`
}

// on registers a command whose payload decodes into T.
func on[T any](r Registrar, typ string, run func(context.Context, realtime.Command, T) error, guards ...realtime.Guard) {
	r.Handle(typ, func(ctx context.Context, c realtime.Command) error {
		in, err := parse[T](c)
		if err != nil {
			return err
		}
		return run(ctx, c, in)
	}, guards...)
}

// RegisterCommands wires item.give, loot.place, loot.remove (Mestre) and item.use, loot.pickup (owner or Mestre).
func RegisterCommands(r Registrar, svc *Service) {
	on(r, "item.give", func(ctx context.Context, c realtime.Command, in GiveInput) error {
		return svc.Give(ctx, c.CampaignID, in)
	}, realtime.GMOnly)
	on(r, "loot.place", func(ctx context.Context, c realtime.Command, in PlaceInput) error {
		_, err := svc.PlaceLoot(ctx, c.CampaignID, in)
		return err
	}, realtime.GMOnly)
	on(r, "loot.remove", func(ctx context.Context, c realtime.Command, in lootRef) error {
		return svc.RemoveLoot(ctx, c.CampaignID, in.LootID)
	}, realtime.GMOnly)
	on(r, "item.use", func(ctx context.Context, c realtime.Command, in useCmd) error {
		return svc.Use(ctx, memberOf(c), in.CharacterID, in.Place, in.Index)
	})
	on(r, "loot.pickup", func(ctx context.Context, c realtime.Command, in lootRef) error {
		return svc.Pickup(ctx, memberOf(c), in.LootID, in.CharacterID)
	})
}
