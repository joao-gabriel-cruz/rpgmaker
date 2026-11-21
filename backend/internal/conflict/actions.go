package conflict

import (
	"context"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

type symbols = assimilacao.Symbols

func (s *Service) active(ctx context.Context, campaignID string) (*Conflict, error) {
	c, err := s.repo.Active(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, bad("Não há partitura de conflito ativa.")
	}
	return c, nil
}

func findObjective(c *Conflict, id string) (*Objective, error) {
	for i := range c.Objectives {
		if c.Objectives[i].ID == id {
			return &c.Objectives[i], nil
		}
	}
	return nil, bad("Objetivo não encontrado.")
}

// finalOf loads a finalized character roll the sender may spend.
func (s *Service) finalOf(ctx context.Context, campaignID string, sender realtime.Sender, rollID string) (rolls.Final, error) {
	fin, err := s.rolls.Final(ctx, campaignID, rollID)
	if err != nil {
		return fin, err
	}
	if fin.Kind != rolls.KindAssimilacao {
		return fin, bad("Apenas rolagens de personagens podem ser atribuídas.")
	}
	if fin.UserID != sender.UserID && sender.Role != campaigns.RoleGM {
		return fin, bad("Você só pode atribuir as suas rolagens.")
	}
	return fin, nil
}

// Assign adds the Successes of a finalized roll to an objective, once per roll.
func (s *Service) Assign(ctx context.Context, campaignID string, sender realtime.Sender, rollID, objectiveID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := s.active(ctx, campaignID)
	if err != nil {
		return err
	}
	o, err := findObjective(c, objectiveID)
	if err != nil {
		return err
	}
	fin, err := s.finalOf(ctx, campaignID, sender, rollID)
	if err != nil {
		return err
	}
	if err := s.repo.Assign(ctx, c.ID, rollID, objectiveID, fin.Symbols.Success); err != nil {
		return err
	}
	return s.credit(ctx, c, o, fin.Symbols)
}

// credit adds the Successes to the objective, saves and announces completion and activations.
func (s *Service) credit(ctx context.Context, c *Conflict, o *Objective, sym symbols) error {
	was := o.Progress >= o.Target
	o.Progress += sym.Success
	fired := tally(c.Activations, SourcePlayers, sym)
	saved, err := s.repo.Save(ctx, *c)
	if err != nil {
		return err
	}
	s.publish(c.CampaignID, &saved)
	if !was && o.Progress >= o.Target {
		s.events.Broadcast(c.CampaignID, realtime.ToAll, EventObjectiveHit, map[string]string{"objectiveId": o.ID, "title": o.Title})
	}
	s.notify(c.CampaignID, fired)
	return nil
}

func (s *Service) notify(campaignID string, fired []Triggered) {
	for _, f := range fired {
		s.events.Broadcast(campaignID, realtime.ToGM, EventActivation, f)
	}
}

func findThreat(c *Conflict, id string) (Threat, error) {
	for _, t := range c.Threats {
		if t.ID == id {
			return t, nil
		}
	}
	return Threat{}, bad("Ameaça não encontrada.")
}

// RollThreat draws the pool of a threat (every die counts) and logs it as a public Mestre roll.
func (s *Service) RollThreat(ctx context.Context, campaignID string, sender realtime.Sender, threatID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := s.active(ctx, campaignID)
	if err != nil {
		return err
	}
	t, err := findThreat(c, threatID)
	if err != nil {
		return err
	}
	sym, err := s.rolls.Threat(ctx, campaignID, sender, t.Name, *t.Pool)
	if err != nil {
		return err
	}
	s.events.Broadcast(campaignID, realtime.ToAll, EventThreatRolled,
		map[string]any{"threatId": t.ID, "name": t.Name, "npcCharacterId": t.NpcCharacterID})
	fired := tally(c.Activations, SourceThreat+t.ID, sym)
	if len(fired) == 0 {
		return nil
	}
	saved, err := s.repo.Save(ctx, *c)
	if err != nil {
		return err
	}
	s.publish(campaignID, &saved)
	s.notify(campaignID, fired)
	return nil
}

// commandError maps domain errors to user-facing messages.
func commandError(err error) error {
	switch {
	case errors.Is(err, ErrAlreadyAssigned):
		return bad("Esta rolagem já foi atribuída a um objetivo.")
	case errors.Is(err, rolls.ErrNotFinal):
		return bad("Finalize a rolagem antes de atribuí-la.")
	case errors.Is(err, rolls.ErrNotFound):
		return bad("Rolagem não encontrada.")
	}
	return err
}
