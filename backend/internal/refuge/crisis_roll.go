package refuge

import (
	"context"
	"fmt"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

const severityDice = 4

// severity is what a count of symbols on the 3 kept dice means.
type severity struct {
	Kind string
	Text string
}

func severityFor(symbols int) severity {
	switch {
	case symbols <= 1:
		return severity{"", "Sem efeito: o Refúgio contorna a ameaça por ora."}
	case symbols == 2:
		return severity{"internal_fight", "Briga interna: a metade da População migra levando metade das Reservas."}
	case symbols == 3:
		return severity{"disease", "Doença: -1 dado em todos os testes até a cura."}
	case symbols == 4:
		return severity{"creature", "Uma criatura assimilada grande aparece."}
	case symbols == 5:
		return severity{"rebellion", "Rebelião: o Refúgio se divide em duas facções."}
	}
	return severity{"invasion", "Ataque de invasores."}
}

func symbolsOf(faces []int, discard int) int {
	total := 0
	for i, f := range faces {
		if i == discard {
			continue
		}
		sym := assimilacao.FaceSymbols(assimilacao.D6, f)
		total += sym.Success + sym.Adaptation + sym.Pressure
	}
	return total
}

// RollSeverity rolls 4d6 on the tray for a crisis (a new custom one when crisisID is empty); the Mestre then discards one.
func (s *Service) RollSeverity(ctx context.Context, campaignID string, sender realtime.Sender, crisisID string) error {
	if s.links.Rolls == nil {
		return bad("Rolagens indisponíveis")
	}
	if crisisID == "" {
		made, err := s.AddCrisis(ctx, campaignID, sender.UserID, CrisisInput{Name: "Crise do Refúgio", Kind: "custom"})
		if err != nil {
			return err
		}
		crisisID = made.ID
	}
	if err := s.requireConfirmed(ctx, campaignID, crisisID); err != nil {
		return err
	}
	e, err := s.links.Rolls.PoolEntry(ctx, campaignID, sender, "Gravidade da crise", rolls.ThreatPool{D6: severityDice})
	if err != nil {
		return err
	}
	_, err = s.Mutate(ctx, campaignID, func(r *Refuge) error {
		c, err := findCrisis(r, crisisID)
		if err == nil {
			c.Roll = &SeverityRoll{RollID: e.RollID, Faces: facesOf(e.Dice)}
		}
		return err
	})
	return err
}

func facesOf(dice []assimilacao.Die) []int {
	faces := make([]int, len(dice))
	for i, d := range dice {
		faces[i] = d.Face
	}
	return faces
}

func (s *Service) requireConfirmed(ctx context.Context, campaignID, id string) error {
	r, err := s.Get(ctx, campaignID)
	if err != nil {
		return err
	}
	c, err := findCrisis(&r, id)
	if err == nil && c.Status == StatusSuggested {
		return bad("Confirme a crise sugerida antes de rolar a gravidade")
	}
	return err
}

// DiscardDie settles the severity roll: it drops die index, counts the other three and updates the crisis.
func (s *Service) DiscardDie(ctx context.Context, campaignID, crisisID string, index int) error {
	var done Crisis
	var sev severity
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		c, err := findCrisis(r, crisisID)
		if err != nil {
			return err
		}
		if c.Roll == nil || c.Roll.Discarded != nil || index < 0 || index >= len(c.Roll.Faces) {
			return bad("Não há rolagem de gravidade esperando o descarte de um dado")
		}
		n := symbolsOf(c.Roll.Faces, index)
		sev = severityFor(n)
		c.Roll.Discarded, c.Roll.Symbols = &index, &n
		applySeverity(c, r, sev, n)
		done = *c
		return nil
	})
	if err != nil {
		return err
	}
	s.announceRoll(campaignID, done, sev)
	return nil
}

func applySeverity(c *Crisis, r *Refuge, sev severity, symbols int) {
	c.Grade = min(symbols, maxGrade)
	if sev.Kind != "" {
		c.Kind = sev.Kind
	}
	text := sev.Text
	if d := c.Description; d != "" && len(d)+len(text)+2 <= maxCrisisTx {
		text = d + "\n\n" + text
	}
	c.Description = text
	if sev.Kind == "internal_fight" {
		c.Suggestions = append(c.Suggestions, CrisisSuggestion{ID: newID(), Kind: "migration", Status: "pending", Text: migrationText(r)})
	}
}

func migrationText(r *Refuge) string {
	return fmt.Sprintf("Metade da População migra (%d de %d níveis) levando metade das Reservas.", r.Population.Level/2, r.Population.Level)
}

func (s *Service) announceRoll(campaignID string, c Crisis, sev severity) {
	s.events.Broadcast(campaignID, realtime.ToAll, EventCrisisRolled, map[string]any{
		"crisisId": c.ID, "name": c.Name, "faces": c.Roll.Faces, "discarded": *c.Roll.Discarded,
		"symbols": *c.Roll.Symbols, "kind": c.Kind, "grade": c.Grade, "text": sev.Text})
}

// ResolveSuggestion applies (or dismisses) a pending numeric consequence of a crisis.
func (s *Service) ResolveSuggestion(ctx context.Context, campaignID, crisisID, sugID string, apply bool) error {
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		c, err := findCrisis(r, crisisID)
		if err != nil {
			return err
		}
		for i := range c.Suggestions {
			sg := &c.Suggestions[i]
			if sg.ID != sugID || sg.Status != "pending" {
				continue
			}
			sg.Status = "dismissed"
			if apply {
				sg.Status = "applied"
				migrate(r)
			}
			return nil
		}
		return bad("Sugestão não encontrada ou já resolvida")
	})
	return err
}

// migrate sends half the population and half of every stock away.
func migrate(r *Refuge) {
	r.Population.Level -= r.Population.Level / 2
	for k, v := range r.Resources {
		r.Resources[k] = v - v/2
	}
}
