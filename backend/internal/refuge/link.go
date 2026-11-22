package refuge

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
)

// Rolls is the part of rolls.Service used by crises and expeditions.
type Rolls interface {
	RequestEntry(ctx context.Context, campaignID string, sender realtime.Sender, in rolls.RequestInput) (rolls.Entry, error)
	Final(ctx context.Context, campaignID, rollID string) (rolls.Final, error)
	PoolEntry(ctx context.Context, campaignID string, sender realtime.Sender, name string, pool rolls.ThreatPool) (rolls.Entry, error)
}

// Suggester is the part of effects.Service used to suggest damage to the Mestre.
type Suggester interface {
	Suggest(ctx context.Context, in effects.Suggestion) (effects.Suggestion, error)
}

// Lookup reads characters (residents, expedition members).
type Lookup interface {
	Get(ctx context.Context, m campaigns.Member, id string) (characters.View, error)
}

// ItemSource returns a compendium item in the sheet item format.
type ItemSource func(ctx context.Context, campaignID, itemID string) (json.RawMessage, error)

// Links are the collaborators of crises, expeditions and residents, set after construction.
type Links struct {
	Rolls      Rolls
	Effects    Suggester
	Chars      Lookup
	Compendium ItemSource
}

// Link wires the collaborators; commands that need a missing one fail with a user error.
func (s *Service) Link(l Links) { s.links = l }
