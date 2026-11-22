package refuge

import (
	"context"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// EventWeekAdvanced is broadcast to everyone when the Mestre confirms a week; payload is {report}.
const EventWeekAdvanced = "refuge.week.advanced"

// EventWeekPreview answers refuge.week.preview to the Mestre only; payload is {report}.
const EventWeekPreview = "refuge.week.preview"

var errGMOnly = realtime.UserError("Apenas o Mestre pode fazer isso")

// OnWeekAdvanced registers fn to run after each confirmed week, once the change is committed.
func (s *Service) OnWeekAdvanced(fn func(ctx context.Context, campaignID string, report WeekReport)) {
	s.hooks = append(s.hooks, fn)
}

// PreviewWeek computes the week without writing anything.
func (s *Service) PreviewWeek(ctx context.Context, m campaigns.Member, adjust map[string]int) (WeekReport, error) {
	if m.Role != campaigns.RoleGM {
		return WeekReport{}, errGMOnly
	}
	if err := ValidateAdjust(adjust); err != nil {
		return WeekReport{}, err
	}
	r, err := s.repo.Get(ctx, m.CampaignID)
	if err != nil {
		return WeekReport{}, err
	}
	_, rep := Simulate(r, adjust)
	return rep, nil
}

// ConfirmWeek writes the week, stores the report in the history, announces it and runs the hooks.
func (s *Service) ConfirmWeek(ctx context.Context, m campaigns.Member, adjust map[string]int) (WeekReport, error) {
	if m.Role != campaigns.RoleGM {
		return WeekReport{}, errGMOnly
	}
	if err := ValidateAdjust(adjust); err != nil {
		return WeekReport{}, err
	}
	var rep WeekReport
	_, err := s.Mutate(ctx, m.CampaignID, func(r *Refuge) error {
		var next Refuge
		next, rep = Simulate(*r, adjust)
		next.History = append(append([]WeekReport{}, r.History...), rep)
		if len(next.History) > historyKeep {
			next.History = next.History[len(next.History)-historyKeep:]
		}
		*r = next
		return nil
	})
	if err != nil {
		return WeekReport{}, err
	}
	s.events.Broadcast(m.CampaignID, realtime.ToAll, EventWeekAdvanced, map[string]WeekReport{"report": rep})
	for _, fn := range s.hooks {
		fn(ctx, m.CampaignID, rep)
	}
	return rep, nil
}
