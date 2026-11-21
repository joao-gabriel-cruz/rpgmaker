// Package rounds tracks the per-campaign conflict round and phase, shared live with every member.
package rounds

import "errors"

// Phase is one step of a round.
type Phase string

// Round phases in play order.
const (
	PhasePreparation Phase = "preparation"
	PhaseInfected    Phase = "infected"
	PhaseAssimilator Phase = "assimilator"
	PhaseEnd         Phase = "end"
)

var phaseOrder = []Phase{PhasePreparation, PhaseInfected, PhaseAssimilator, PhaseEnd}

// State is the round tracker of a campaign. Inactive means no conflict is running.
type State struct {
	Active bool  `json:"active"`
	Round  int   `json:"round"`
	Phase  Phase `json:"phase"`
}

var (
	errInactive = errors.New("no conflict in progress")
	errAtStart  = errors.New("conflict is at its start")
)

// Idle is the state of a campaign without a running conflict.
func Idle() State { return State{Phase: PhasePreparation} }

// Start begins a conflict at round 1, Preparação.
func Start() State { return State{Active: true, Round: 1, Phase: PhasePreparation} }

func index(p Phase) int {
	for i, q := range phaseOrder {
		if q == p {
			return i
		}
	}
	return 0
}

// Next advances one phase; after Fim it opens Preparação of the following round.
func (s State) Next() (State, error) {
	if !s.Active {
		return s, errInactive
	}
	i := index(s.Phase)
	if i == len(phaseOrder)-1 {
		return State{Active: true, Round: s.Round + 1, Phase: PhasePreparation}, nil
	}
	s.Phase = phaseOrder[i+1]
	return s, nil
}

// Prev goes back one phase; before Preparação it returns to Fim of the previous round.
func (s State) Prev() (State, error) {
	if !s.Active {
		return s, errInactive
	}
	i := index(s.Phase)
	if i > 0 {
		s.Phase = phaseOrder[i-1]
		return s, nil
	}
	if s.Round <= 1 {
		return s, errAtStart
	}
	return State{Active: true, Round: s.Round - 1, Phase: PhaseEnd}, nil
}
