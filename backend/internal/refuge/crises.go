package refuge

import (
	"context"
	"slices"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Crisis statuses; a suggested crisis waits for the Mestre to confirm it.
const (
	StatusSuggested = "suggested"
	StatusActive    = "active"
	StatusContained = "contained"
	StatusResolved  = "resolved"
)

// Events about crises, all announced by the server.
const (
	EventCrisisSuggested = "refuge.crisis.suggested"
	EventCrisisRolled    = "refuge.crisis.rolled"
)

const (
	maxCrises   = 100
	maxGrade    = 10
	maxCrisisTx = 2000
)

var (
	crisisKinds    = []string{"overpopulation", "breach", "shortage", "custom", "internal_fight", "disease", "creature", "rebellion", "invasion"}
	crisisStatuses = []string{StatusSuggested, StatusActive, StatusContained, StatusResolved}
)

// ResourceLabels are the PT-BR names of the 12 resources.
var ResourceLabels = map[string]string{
	"water": "Água", "food": "Alimento", "medicine": "Remédios", "clothing": "Vestuário", "plants": "Plantas", "animals": "Animais",
	"wood": "Madeira", "biomass": "Biomassa", "minerals": "Minerais", "fuel": "Combustível", "ammo": "Munição", "buildingMaterial": "Material de Construção",
}

// CrisisSettings are the Mestre's options for crises; the Moral penalty is on by default.
type CrisisSettings struct {
	MoralePenalty bool `json:"moralePenalty"`
}

// DefaultCrisisSettings turns the -1 Moral per confirmed crisis on.
func DefaultCrisisSettings() CrisisSettings { return CrisisSettings{MoralePenalty: true} }

// SeverityRoll is the 4d6 severity roll of a crisis: faces on the tray and the Mestre's discard.
type SeverityRoll struct {
	RollID    string `json:"rollId"`
	Faces     []int  `json:"faces"`
	Discarded *int   `json:"discarded"`
	Symbols   *int   `json:"symbols"`
}

// CrisisSuggestion is a numeric consequence the Mestre may apply with one click.
type CrisisSuggestion struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Text   string `json:"text"`
	Status string `json:"status"`
}

// Crisis is a trouble of the Refúgio with a grade (0-10) and a status.
type Crisis struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Kind        string             `json:"kind"`
	Grade       int                `json:"grade"`
	Status      string             `json:"status"`
	Description string             `json:"description"`
	Week        int                `json:"week"`
	CreatedBy   string             `json:"createdBy"`
	Trigger     string             `json:"trigger,omitempty"`
	Roll        *SeverityRoll      `json:"roll,omitempty"`
	Suggestions []CrisisSuggestion `json:"suggestions"`
}

// CrisisInput creates a crisis.
type CrisisInput struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Grade       int    `json:"grade"`
	Description string `json:"description"`
}

// CrisisPatch edits the given fields of a crisis.
type CrisisPatch struct {
	ID          string  `json:"id"`
	Name        *string `json:"name"`
	Kind        *string `json:"kind"`
	Grade       *int    `json:"grade"`
	Status      *string `json:"status"`
	Description *string `json:"description"`
}

func bad(msg string) error { return realtime.UserError(msg) }

func (c Crisis) validate() error {
	if len(c.Name) == 0 || len(c.Name) > maxName || len(c.Description) > maxCrisisTx {
		return bad("Nome (até 100) ou descrição (até 2000) inválidos")
	}
	if !slices.Contains(crisisKinds, c.Kind) || !slices.Contains(crisisStatuses, c.Status) || c.Grade < 0 || c.Grade > maxGrade {
		return bad("Tipo, status ou gravidade (0 a 10) inválidos")
	}
	return nil
}

func findCrisis(r *Refuge, id string) (*Crisis, error) {
	for i := range r.Crises {
		if r.Crises[i].ID == id {
			return &r.Crises[i], nil
		}
	}
	return nil, bad("Crise não encontrada")
}

// conditions lists the active alerts: an active Doença crisis costs -1 die on every test.
func (r Refuge) conditions() []string {
	out := []string{}
	for _, c := range r.Crises {
		if c.Kind == "disease" && c.Status == StatusActive {
			return append(out, "disease")
		}
	}
	return out
}

// AddCrisis creates an active crisis chosen by the Mestre.
func (s *Service) AddCrisis(ctx context.Context, campaignID, userID string, in CrisisInput) (Crisis, error) {
	var made Crisis
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		made = Crisis{ID: newID(), Name: in.Name, Kind: in.Kind, Grade: in.Grade, Status: StatusActive,
			Description: in.Description, Week: r.Week, CreatedBy: userID, Suggestions: []CrisisSuggestion{}}
		if err := made.validate(); err != nil {
			return err
		}
		if len(r.Crises) >= maxCrises {
			return bad("Limite de crises atingido")
		}
		r.Crises = append(r.Crises, made)
		return nil
	})
	return made, err
}

// UpdateCrisis edits name, kind, grade, status or description of a crisis.
func (s *Service) UpdateCrisis(ctx context.Context, campaignID string, p CrisisPatch) error {
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		c, err := findCrisis(r, p.ID)
		if err != nil {
			return err
		}
		next := *c
		patchString(&next.Name, p.Name)
		patchString(&next.Kind, p.Kind)
		patchString(&next.Status, p.Status)
		patchString(&next.Description, p.Description)
		if p.Grade != nil {
			next.Grade = *p.Grade
		}
		if err := next.validate(); err != nil {
			return err
		}
		*c = next
		return nil
	})
	return err
}

func patchString(dst *string, v *string) {
	if v != nil {
		*dst = *v
	}
}

// DeleteCrisis removes a crisis (also used to dismiss a suggested one).
func (s *Service) DeleteCrisis(ctx context.Context, campaignID, id string) error {
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		if _, err := findCrisis(r, id); err != nil {
			return err
		}
		r.Crises = slices.DeleteFunc(r.Crises, func(c Crisis) bool { return c.ID == id })
		return nil
	})
	return err
}

// ConfirmCrisis turns a suggested crisis into an active one, costing 1 Moral when the option is on.
func (s *Service) ConfirmCrisis(ctx context.Context, campaignID, id string) error {
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		c, err := findCrisis(r, id)
		if err != nil {
			return err
		}
		if c.Status != StatusSuggested {
			return bad("Esta crise já foi confirmada")
		}
		c.Status = StatusActive
		if r.CrisisSettings.MoralePenalty {
			r.Morale = max(r.Morale-1, 0)
		}
		return nil
	})
	return err
}

// SetCrisisSettings stores the crisis options.
func (s *Service) SetCrisisSettings(ctx context.Context, campaignID string, st CrisisSettings) error {
	_, err := s.Mutate(ctx, campaignID, func(r *Refuge) error { r.CrisisSettings = st; return nil })
	return err
}
