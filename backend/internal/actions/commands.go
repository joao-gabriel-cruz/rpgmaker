package actions

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Registrar is the part of the realtime hub used to register commands.
type Registrar interface {
	Handle(typ string, fn realtime.HandlerFunc, guards ...realtime.Guard)
}

// RegisterCommands wires action.perform (owner or Mestre) and action.card.accept (Mestre).
func RegisterCommands(r Registrar, svc *Service) {
	r.Handle("action.perform", func(ctx context.Context, c realtime.Command) error {
		var in PerformInput
		if err := json.Unmarshal(c.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		return svc.Perform(ctx, c.CampaignID, c.Sender, in)
	})
	r.Handle("action.card.accept", func(ctx context.Context, c realtime.Command) error {
		var in struct {
			CharacterID string `json:"characterId"`
			Card
		}
		if err := json.Unmarshal(c.Payload, &in); err != nil {
			return realtime.UserError("Payload inválido")
		}
		return svc.AcceptCard(ctx, c.CampaignID, in.CharacterID, in.Card)
	}, realtime.GMOnly)
}

// AcceptCard adds a confirmed rest card to the character's list of Assimilações.
func (s *Service) AcceptCard(ctx context.Context, campaignID, charID string, c Card) error {
	if _, ok := suitLabel[c.Suit]; !ok || strings.TrimSpace(c.Name) == "" ||
		utf8.RuneCountInString(c.Name) > 80 || utf8.RuneCountInString(c.Description) > 300 {
		return realtime.UserError("Carta inválida")
	}
	_, err := s.chars.EditSheet(ctx, campaignID, charID, func(sheet json.RawMessage) (json.RawMessage, error) {
		doc := map[string]json.RawMessage{}
		if err := json.Unmarshal(sheet, &doc); err != nil {
			return nil, err
		}
		list := []Card{}
		if raw, ok := doc["assimilations"]; ok {
			if err := json.Unmarshal(raw, &list); err != nil {
				return nil, err
			}
		}
		doc["assimilations"], _ = json.Marshal(append(list, c))
		return json.Marshal(doc)
	})
	if errors.Is(err, characters.ErrNotFound) {
		return realtime.UserError("Personagem não encontrado")
	}
	return err
}
