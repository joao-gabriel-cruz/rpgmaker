package items

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

var errNotConsumable = realtime.UserError("Este item não pode ser usado")

// consumed is the part of a used sheet item the effects need.
type consumed struct {
	Name    string      `json:"name"`
	Kind    string      `json:"kind"`
	Effects ItemEffects `json:"effects"`
}

// Use removes a consumable from the character's inventory and applies its effects: a heal suggestion
// for the Mestre (effects flow) and refuge resources added directly. Owner or Mestre only.
func (s *Service) Use(ctx context.Context, m campaigns.Member, charID, place string, index int) error {
	view, err := s.chars.Get(ctx, m, charID)
	if err != nil {
		return mapChar(err)
	}
	if m.Role != campaigns.RoleGM && (view.OwnerID == nil || *view.OwnerID != m.UserID) {
		return realtime.UserError("Você só pode usar itens dos seus personagens")
	}
	var used consumed
	_, err = s.chars.EditSheet(ctx, m.CampaignID, charID, func(sheet json.RawMessage) (json.RawMessage, error) {
		out, taken, err := takeAt(sheet, place, index)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(taken, &used); err != nil || used.Kind != KindConsumable {
			return nil, errNotConsumable
		}
		return out, nil
	})
	if err != nil {
		return mapChar(err)
	}
	return s.applyUse(ctx, m.CampaignID, view.ID, view.Name, used)
}

func (s *Service) applyUse(ctx context.Context, campaignID, charID, charName string, used consumed) error {
	s.events.Broadcast(campaignID, realtime.ToAll, EventItemUsed, map[string]string{
		"characterId": charID, "characterName": charName, "itemName": used.Name})
	if h := used.Effects.Heal; h > 0 {
		_, err := s.effects.Suggest(ctx, effects.Suggestion{CampaignID: campaignID, CharacterID: charID, CharacterName: charName,
			Source: used.Name, Trigger: effects.TriggerItemUse, Effect: effects.EffectHeal, Amount: h,
			Message: fmt.Sprintf("Usou %s: curar %d", used.Name, h)})
		if err != nil {
			slog.Error("items: heal suggestion failed", "err", err)
			return err
		}
	}
	if r := used.Effects.RefugeResource; r != nil {
		if _, err := s.refuge.AddResource(ctx, campaignID, r.Resource, r.Amount); err != nil {
			slog.Error("items: refuge resource failed", "err", err)
			return err
		}
	}
	return nil
}
