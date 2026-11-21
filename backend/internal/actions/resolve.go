package actions

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/items"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// Outcome kinds sent in action.resolved.
const (
	KindNone      = "none"
	KindDamage    = "damage_suggested"
	KindHeal      = "heal_suggested"
	KindLoot      = "loot_found"
	KindRest      = "rest"
	KindMorale    = "morale"
	KindNotice    = "notice"
	triggerAction = "action"
)

type outcome struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}

// Finalized is the rolls.OnFinalized hook: it settles the consequence of the action tied to the roll.
func (s *Service) Finalized(ctx context.Context, campaignID, rollID string) {
	r := s.take(rollID)
	if r == nil {
		return
	}
	fin, err := s.rolls.Final(ctx, campaignID, rollID)
	if err != nil {
		slog.Error("actions: read final roll failed", "err", err)
		return
	}
	out := s.consequence(ctx, campaignID, r, fin.Symbols)
	s.events.Broadcast(campaignID, realtime.ToAll, EventResolved, map[string]any{
		"actionRunId": r.id, "actionId": r.act.ID, "actionName": r.act.Name, "actorName": r.char.Name,
		"animation": r.act.Animation, "actorTokenId": r.actor.ID,
		"target": r.target, "rollId": rollID, "outcome": out,
		"result": map[string]int{"success": fin.Symbols.Success, "adaptation": fin.Symbols.Adaptation, "pressure": fin.Symbols.Pressure}})
}

func (s *Service) consequence(ctx context.Context, campaignID string, r *run, sym assimilacao.Symbols) outcome {
	var out outcome
	var err error
	switch r.act.Outcome {
	case OutcomeDamage:
		out, err = s.hit(ctx, r, sym.Success, effects.EffectDamage, KindDamage)
	case OutcomeHeal:
		out, err = s.hit(ctx, r, sym.Success, effects.EffectHeal, KindHeal)
	case OutcomeLoot:
		out, err = s.search(ctx, campaignID, r, sym.Success)
	case OutcomeRest:
		out, err = s.rest(ctx, campaignID, r, sym)
	case OutcomeMorale:
		out, err = s.morale(ctx, r, sym.Success)
	case OutcomeNotice:
		out = outcome{KindNotice, fmt.Sprintf("%s usou %s (%d Sucesso(s)).", r.char.Name, r.act.Name, sym.Success)}
	default:
		out = outcome{Kind: KindNone}
	}
	if err != nil {
		slog.Error("actions: consequence failed", "action", r.act.ID, "err", err)
		return outcome{KindNone, "Não foi possível aplicar a consequência."}
	}
	return out
}

func (s *Service) suggest(ctx context.Context, r *run, charID, charName, effect string, amount int, msg string) error {
	_, err := s.effects.Suggest(ctx, effects.Suggestion{CampaignID: r.char.CampaignID, CharacterID: charID, CharacterName: charName,
		SceneID: r.sceneID, X: r.actor.X, Z: r.actor.Z, Source: r.act.Name, Trigger: triggerAction,
		Effect: effect, Amount: amount, Message: msg})
	return err
}

// hit suggests damage or healing on the targeted token: the Successes, at least 1 when there was any.
func (s *Service) hit(ctx context.Context, r *run, successes int, effect, kind string) (outcome, error) {
	if r.victim == nil || successes < 1 {
		return outcome{KindNone, fmt.Sprintf("%s falhou.", r.act.Name)}, nil
	}
	amount := min(max(1, successes), effects.MaxDamage)
	msg := fmt.Sprintf("%s: %s em %s (%d)", r.char.Name, r.act.Name, r.victim.Name, amount)
	if err := s.suggest(ctx, r, r.victim.CharacterID, r.victim.Name, effect, amount, msg); err != nil {
		return outcome{}, err
	}
	return outcome{kind, fmt.Sprintf("%s deu certo em %s!", r.act.Name, r.victim.Name)}, nil
}

// search reveals hidden loot: Perceber scans a radius of Successes+2 around the actor; a targeted
// search looks only at the object or cell.
func (s *Service) search(ctx context.Context, campaignID string, r *run, successes int) (outcome, error) {
	found, err := s.loot.TryFind(ctx, campaignID, r.char.ID, r.sceneID, searchTarget(r, successes), successes)
	if err != nil {
		return outcome{}, err
	}
	if len(found) == 0 {
		return outcome{KindNone, "Nada de interessante por aqui."}, nil
	}
	return outcome{KindLoot, fmt.Sprintf("%d item(ns) encontrado(s)!", len(found))}, nil
}

func searchTarget(r *run, successes int) items.Target {
	switch r.target.Kind {
	case TargetObject:
		return items.Target{ObjectID: r.target.ObjectID, Cells: []maps.Coord{{X: *r.target.X, Z: *r.target.Z}}}
	case TargetCell:
		return items.Target{Cells: []maps.Coord{{X: *r.target.X, Z: *r.target.Z}}}
	}
	radius := successes + lootBonus
	var cells []maps.Coord
	for z := r.actor.Z - radius; z <= r.actor.Z+radius; z++ {
		for x := r.actor.X - radius; x <= r.actor.X+radius; x++ {
			cells = append(cells, maps.Coord{X: x, Z: z})
		}
	}
	return items.Target{Cells: cells}
}

func (s *Service) morale(ctx context.Context, r *run, successes int) (outcome, error) {
	if successes < 1 {
		return outcome{KindNone, "A dança não animou ninguém."}, nil
	}
	msg := fmt.Sprintf("%s animou o grupo dançando: sugestão de +1 de Moral no Refúgio.", r.char.Name)
	if err := s.suggest(ctx, r, r.char.ID, r.char.Name, effects.EffectNotice, 0, msg); err != nil {
		return outcome{}, err
	}
	return outcome{KindMorale, fmt.Sprintf("%s animou o grupo!", r.char.Name)}, nil
}
