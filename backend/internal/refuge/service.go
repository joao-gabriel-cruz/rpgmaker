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
	Mutate(ctx context.Context, campaignID string, fn MutateFunc) (Refuge, error)
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

// Get returns the campaign's refuge with its derived values.
func (s *Service) Get(ctx context.Context, campaignID string) (Refuge, error) {
	r, err := s.repo.Get(ctx, campaignID)
	return withDerived(r), err
}

func withDerived(r Refuge) Refuge {
	r.Derived = r.Derive()
	return r
}

func (s *Service) publish(r Refuge) Refuge {
	r = withDerived(r)
	s.events.Broadcast(r.CampaignID, realtime.ToAll, EventUpdated, map[string]Refuge{"refuge": r})
	return r
}

// Mutate edits the refuge transactionally, validates the result and broadcasts refuge.updated.
// Later deliveries (buildings, week, crises, expeditions) change state only through this.
func (s *Service) Mutate(ctx context.Context, campaignID string, fn func(*Refuge) error) (Refuge, error) {
	r, err := s.repo.Mutate(ctx, campaignID, func(r *Refuge) error {
		if err := fn(r); err != nil {
			return err
		}
		if err := r.validate(); err != nil {
			return err
		}
		r.normalize()
		return nil
	})
	if err != nil {
		return Refuge{}, err
	}
	return s.publish(r), nil
}

// Update replaces the editable fields (text and sheet); only the gm may.
func (s *Service) Update(ctx context.Context, m campaigns.Member, e Editable) (Refuge, error) {
	if m.Role != campaigns.RoleGM {
		return Refuge{}, realtime.UserError("Apenas o Mestre pode fazer isso")
	}
	if err := e.validate(); err != nil {
		return Refuge{}, err
	}
	return s.Mutate(ctx, m.CampaignID, func(r *Refuge) error { r.Editable = e; return nil })
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
	return s.publish(r), nil
}

// AddItem appends an item (sheet item format, kept verbatim) to the shared inventory.
func (s *Service) AddItem(ctx context.Context, campaignID string, item json.RawMessage) (Refuge, error) {
	return s.Mutate(ctx, campaignID, func(r *Refuge) error {
		if len(r.Items) >= maxItems {
			return errRefugeMax
		}
		r.Items = append(r.Items, item)
		return nil
	})
}

// AddResource stocks any of the 12 resources, discarding what exceeds the reserves cap.
func (s *Service) AddResource(ctx context.Context, campaignID, resource string, amount int) (Refuge, error) {
	r, _, err := s.AddResourceReport(ctx, campaignID, resource, amount)
	return r, err
}

// AddResourceReport is AddResource that also returns how many units were discarded for lack of room.
func (s *Service) AddResourceReport(ctx context.Context, campaignID, resource string, amount int) (Refuge, int, error) {
	if amount <= 0 || !IsResource(resource) {
		return Refuge{}, 0, ErrInvalid
	}
	discarded := 0
	r, err := s.Mutate(ctx, campaignID, func(r *Refuge) error {
		discarded = r.AddResource(resource, amount)
		return nil
	})
	return r, discarded, err
}
