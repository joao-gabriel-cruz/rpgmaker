package refuge

import (
	"context"
	"log/slog"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// triggerName is the display name of a crisis born from a week trigger.
func triggerName(t CrisisTrigger) string {
	switch t.Kind {
	case "overpopulation":
		return "Superlotação"
	case "breach":
		return "Defesa insuficiente"
	}
	if l, ok := ResourceLabels[t.Resource]; ok {
		return "Falta de " + l
	}
	return "Falta de recursos"
}

func triggerKey(t CrisisTrigger) string { return t.Kind + ":" + t.Resource }

// onWeekAdvanced turns week triggers into crises suggested to the Mestre (confirming costs 1 Moral).
func (s *Service) onWeekAdvanced(ctx context.Context, campaignID string, report WeekReport) {
	if len(report.Triggers) == 0 {
		return
	}
	var created []Crisis
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		created = created[:0]
		for _, t := range report.Triggers {
			if hasOpenTrigger(r.Crises, triggerKey(t)) || len(r.Crises) >= maxCrises || !isTriggerKind(t.Kind) {
				continue
			}
			c := Crisis{ID: newID(), Name: triggerName(t), Kind: t.Kind, Status: StatusSuggested, Description: t.Detail,
				Week: r.Week, CreatedBy: "system", Trigger: triggerKey(t), Suggestions: []CrisisSuggestion{}}
			r.Crises = append(r.Crises, c)
			created = append(created, c)
		}
		return nil
	})
	if err != nil {
		slog.Error("refuge: crisis suggestions failed", "err", err)
		return
	}
	for _, c := range created {
		s.events.Broadcast(campaignID, realtime.ToGM, EventCrisisSuggested, map[string]Crisis{"crisis": c})
	}
}

func isTriggerKind(k string) bool { return k == "overpopulation" || k == "breach" || k == "shortage" }

func hasOpenTrigger(crises []Crisis, key string) bool {
	for _, c := range crises {
		if c.Trigger == key && (c.Status == StatusSuggested || c.Status == StatusActive) {
			return true
		}
	}
	return false
}
