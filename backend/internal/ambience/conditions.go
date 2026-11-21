package ambience

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// SheetEdit rewrites a stored sheet (same shape as characters.Service.EditSheet's callback).
type SheetEdit = func(json.RawMessage) (json.RawMessage, error)

// ownerConditions are the conditions a player may toggle on their own character.
var ownerConditions = []string{"hidden"}

// SetCondition adds or removes a manual condition on a character's sheet and announces character.updated.
func (s *Service) SetCondition(ctx context.Context, snd realtime.Sender, campaignID, charID, cond string, active bool) error {
	if !slices.Contains(assimilacao.Conditions, cond) {
		return realtime.UserError("Condição inválida")
	}
	m := campaigns.Member{CampaignID: campaignID, UserID: snd.UserID, Role: snd.Role}
	v, err := s.chars.Get(ctx, m, charID)
	if err != nil {
		return realtime.UserError("Personagem não encontrado")
	}
	mine := v.OwnerID != nil && *v.OwnerID == snd.UserID
	if snd.Role != campaigns.RoleGM && (!mine || !slices.Contains(ownerConditions, cond)) {
		return realtime.UserError("Você não pode aplicar essa condição")
	}
	_, err = s.chars.EditSheet(ctx, campaignID, charID, func(raw json.RawMessage) (json.RawMessage, error) {
		return toggleCondition(raw, cond, active)
	})
	return err
}

func toggleCondition(raw json.RawMessage, cond string, active bool) (json.RawMessage, error) {
	sheet := map[string]json.RawMessage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &sheet); err != nil {
			return nil, err
		}
	}
	var list []string
	_ = json.Unmarshal(sheet["conditions"], &list)
	list = slices.DeleteFunc(list, func(c string) bool { return c == cond })
	if active {
		list = append(list, cond)
	}
	if len(list) == 0 {
		delete(sheet, "conditions")
	} else {
		b, _ := json.Marshal(list)
		sheet["conditions"] = b
	}
	return json.Marshal(sheet)
}
