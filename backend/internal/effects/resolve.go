package effects

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Pending lists the open suggestions of the campaign.
func (s *Service) Pending(ctx context.Context, campaignID string) ([]Suggestion, error) {
	return s.store.Pending(ctx, campaignID)
}

// Dismiss closes a suggestion without changing anything.
func (s *Service) Dismiss(ctx context.Context, campaignID, sugID string) error {
	if !uuidRe.MatchString(sugID) {
		return realtime.UserError("Sugestão não encontrada")
	}
	if _, err := s.store.Resolve(ctx, campaignID, sugID, StatusDismissed); err != nil {
		return notFound(err)
	}
	s.resolved(campaignID, sugID, false)
	return nil
}

// Confirm applies a suggestion to the character and closes it; a failed application leaves it pending.
func (s *Service) Confirm(ctx context.Context, campaignID string, gm realtime.Sender, sugID string) error {
	if !uuidRe.MatchString(sugID) {
		return realtime.UserError("Sugestão não encontrada")
	}
	sg, err := s.store.Resolve(ctx, campaignID, sugID, StatusApplied)
	if err != nil {
		return notFound(err)
	}
	owner, err := s.apply(ctx, campaignID, gm, sg)
	if err != nil {
		_ = s.store.Reopen(ctx, campaignID, sugID)
		return err
	}
	s.resolved(campaignID, sugID, true)
	if owner != "" {
		s.events.Broadcast(campaignID, realtime.ToUser(owner), EventApplied, map[string]any{
			"characterId": sg.CharacterID, "characterName": sg.CharacterName, "source": sg.Source,
			"effect": sg.Effect, "amount": sg.Amount, "message": sg.Message})
	}
	return nil
}

func notFound(err error) error {
	if errors.Is(err, ErrNotFound) {
		return realtime.UserError("Sugestão não encontrada ou já resolvida")
	}
	return err
}

func (s *Service) resolved(campaignID, sugID string, applied bool) {
	s.events.Broadcast(campaignID, realtime.ToGM, EventResolved, map[string]any{"id": sugID, "applied": applied})
}

// apply performs the effect and returns the owner user id to tell (empty when none).
func (s *Service) apply(ctx context.Context, campaignID string, gm realtime.Sender, sg Suggestion) (string, error) {
	m := campaigns.Member{CampaignID: campaignID, UserID: gm.UserID, Role: campaigns.RoleGM}
	view, err := s.chars.Get(ctx, m, sg.CharacterID)
	if err != nil {
		if errors.Is(err, characters.ErrNotFound) {
			return "", realtime.UserError("Personagem não encontrado")
		}
		return "", err
	}
	switch sg.Effect {
	case EffectDamage:
		err = s.addDamage(ctx, m, view, sg.Amount)
	case EffectDifficulty:
		err = s.addDifficulty(ctx, m, gm, view.ID, sg.Amount)
	case EffectHeal:
		err = s.addDamage(ctx, m, view, -sg.Amount)
	}
	if err != nil || view.OwnerID == nil {
		return "", err
	}
	return *view.OwnerID, nil
}

func (s *Service) addDifficulty(ctx context.Context, m campaigns.Member, gm realtime.Sender, charID string, amount int) error {
	st, err := s.rolls.State(ctx, m, charID)
	if err != nil {
		return err
	}
	return s.rolls.SetDifficulty(ctx, m.CampaignID, gm, charID, min(st.DifficultyD12+amount, MaxDifficulty))
}

func (s *Service) addDamage(ctx context.Context, m campaigns.Member, view characters.View, amount int) error {
	sheet, err := withDamage(view.Sheet, amount)
	if err != nil {
		return err
	}
	_, err = s.chars.Save(ctx, m, view.ID, sheet)
	var rej *characters.RejectedError
	if errors.As(err, &rej) {
		return realtime.UserError("A ficha do personagem não pôde ser atualizada")
	}
	return err
}

// withDamage adds to health.damage (never below zero) keeping every other sheet field verbatim.
func withDamage(sheet json.RawMessage, amount int) (json.RawMessage, error) {
	doc := map[string]json.RawMessage{}
	if err := json.Unmarshal(sheet, &doc); err != nil {
		return nil, err
	}
	health := map[string]json.RawMessage{}
	if raw, ok := doc["health"]; ok {
		if err := json.Unmarshal(raw, &health); err != nil {
			return nil, err
		}
	}
	var cur int
	if raw, ok := health["damage"]; ok {
		_ = json.Unmarshal(raw, &cur)
	}
	health["damage"], _ = json.Marshal(max(cur+amount, 0))
	doc["health"], _ = json.Marshal(health)
	return json.Marshal(doc)
}
