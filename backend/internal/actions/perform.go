package actions

import (
	"context"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

func deny(msg string) error { return realtime.UserError(msg) }

// Perform validates an action, rolls for it through the rolls flow and announces it to the table.
func (s *Service) Perform(ctx context.Context, campaignID string, sender realtime.Sender, in PerformInput) error {
	view, err := s.authorize(ctx, campaignID, sender, in.CharacterID)
	if err != nil {
		return err
	}
	act, err := s.choose(ctx, campaignID, in)
	if err != nil {
		return err
	}
	sp, err := s.locate(ctx, campaignID, view.ID)
	if err != nil {
		return err
	}
	tgt, victim, err := sp.check(act, in.Target)
	if err != nil {
		return err
	}
	bonus, variant := weaponFor(view.Sheet, act)
	entry, err := s.rolls.RequestEntry(ctx, campaignID, sender, requestFor(view.ID, act.Rolls[in.RollChoice], in.Modifiers, bonus))
	if err != nil {
		return rollError(err)
	}
	r := &run{id: newRunID(), rollID: entry.RollID, act: act, char: view, actor: sp.actor, target: tgt, victim: victim,
		sceneID: sp.sc.ID, safe: sp.safe(), variant: variant}
	s.remember(r)
	s.events.Broadcast(campaignID, realtime.ToAll, EventStarted, startedPayload(r))
	return nil
}

// choose finds the enabled action and checks the chosen roll index.
func (s *Service) choose(ctx context.Context, campaignID string, in PerformInput) (Action, error) {
	act, err := s.Find(ctx, campaignID, in.ActionID)
	if errors.Is(err, ErrNotFound) {
		return act, deny("Ação não encontrada ou desativada.")
	}
	if err == nil && (in.RollChoice < 0 || in.RollChoice >= len(act.Rolls)) {
		return act, deny("Rolagem inválida para esta ação.")
	}
	return act, err
}

func (s *Service) authorize(ctx context.Context, campaignID string, sender realtime.Sender, charID string) (characters.View, error) {
	view, err := s.chars.Get(ctx, campaigns.Member{CampaignID: campaignID, UserID: sender.UserID, Role: sender.Role}, charID)
	if errors.Is(err, characters.ErrNotFound) {
		return view, deny("Personagem não encontrado.")
	}
	if err != nil {
		return view, err
	}
	if sender.Role != campaigns.RoleGM && (view.OwnerID == nil || *view.OwnerID != sender.UserID) {
		return view, deny("Você só pode agir com os seus personagens.")
	}
	return view, nil
}

func requestFor(charID string, opt RollOption, m Modifiers, bonus int) rolls.RequestInput {
	return rolls.RequestInput{CharacterID: charID, Instinct: opt.Instinct, Aptitude: opt.Aptitude,
		SecondInstinct: m.SecondInstinct, Effort: m.Effort, UseOccupation: m.UseOccupation,
		UseMarkingEvent: m.UseMarkingEvent, BonusKept: bonus}
}

// rollError turns rule failures of the roll into messages for the player.
func rollError(err error) error {
	var a assimilacao.RuleError
	var g rolls.RuleError
	if errors.As(err, &a) || errors.As(err, &g) {
		return deny(err.Error())
	}
	return err
}

func startedPayload(r *run) map[string]any {
	p := map[string]any{"actionRunId": r.id, "actionId": r.act.ID, "actionName": r.act.Name, "animation": r.act.Animation,
		"actorCharacterId": r.char.ID, "actorName": r.char.Name, "actorTokenId": r.actor.ID, "target": r.target, "rollId": r.rollID,
		"actor": map[string]int{"x": r.actor.X, "z": r.actor.Z}}
	if r.variant != "" {
		p["variant"] = r.variant
	}
	return p
}
