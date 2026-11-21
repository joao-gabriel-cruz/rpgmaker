package refuge

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// EventUpdated is broadcast to everyone whenever the refuge changes; payload is {refuge}.
const EventUpdated = "refuge.updated"

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Repository is the persistence the Service needs.
type Repository interface {
	Get(ctx context.Context, campaignID string) (Refuge, error)
	Update(ctx context.Context, campaignID string, d Details) (Refuge, error)
	Move(ctx context.Context, campaignID, charID string, fn MoveFunc) (Refuge, error)
}

// Characters announces a changed character (character.updated) through the characters service.
type Characters interface {
	Announce(ctx context.Context, campaignID, charID string) error
}

// Events is the part of realtime.Hub used to notify clients.
type Events interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the refuge rules.
type Service struct {
	repo   Repository
	chars  Characters
	events Events
}

// NewService builds a Service.
func NewService(repo Repository, chars Characters, ev Events) *Service {
	return &Service{repo: repo, chars: chars, events: ev}
}

// Get returns the campaign's refuge.
func (s *Service) Get(ctx context.Context, campaignID string) (Refuge, error) {
	return s.repo.Get(ctx, campaignID)
}

// Update replaces the editable fields; only the gm may.
func (s *Service) Update(ctx context.Context, m campaigns.Member, d Details) (Refuge, error) {
	if m.Role != campaigns.RoleGM {
		return Refuge{}, realtime.UserError("Apenas o Mestre pode fazer isso")
	}
	if err := d.validate(); err != nil {
		return Refuge{}, err
	}
	r, err := s.repo.Update(ctx, m.CampaignID, d)
	if err != nil {
		return Refuge{}, err
	}
	s.events.Broadcast(m.CampaignID, realtime.ToAll, EventUpdated, map[string]Refuge{"refuge": r})
	return r, nil
}

// Deposit moves a character's item from body or backpack into the refuge.
func (s *Service) Deposit(ctx context.Context, m campaigns.Member, charID, from string, index int) (Refuge, error) {
	return s.move(ctx, m, charID, func(c CharState, items []json.RawMessage) (json.RawMessage, []json.RawMessage, error) {
		return deposit(c, items, from, index)
	})
}

// Withdraw moves a refuge item into a character's body or backpack.
func (s *Service) Withdraw(ctx context.Context, m campaigns.Member, charID, to string, index int) (Refuge, error) {
	return s.move(ctx, m, charID, func(c CharState, items []json.RawMessage) (json.RawMessage, []json.RawMessage, error) {
		return withdraw(c, items, to, index)
	})
}

func (s *Service) move(ctx context.Context, m campaigns.Member, charID string, fn MoveFunc) (Refuge, error) {
	if !uuidRe.MatchString(charID) {
		return Refuge{}, errNoChar
	}
	guarded := func(c CharState, items []json.RawMessage) (json.RawMessage, []json.RawMessage, error) {
		if m.Role != campaigns.RoleGM && (c.Kind != "pc" || c.OwnerID == nil || *c.OwnerID != m.UserID) {
			return nil, nil, errNotYours
		}
		return fn(c, items)
	}
	r, err := s.repo.Move(ctx, m.CampaignID, charID, guarded)
	if errors.Is(err, ErrCharNotFound) {
		return Refuge{}, errNoChar
	}
	if err != nil {
		return Refuge{}, err
	}
	if err := s.chars.Announce(ctx, m.CampaignID, charID); err != nil {
		return Refuge{}, err
	}
	s.events.Broadcast(m.CampaignID, realtime.ToAll, EventUpdated, map[string]Refuge{"refuge": r})
	return r, nil
}
