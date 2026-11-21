package conflict

import (
	"context"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
)

// Repository is the persistence the Service needs.
type Repository interface {
	Active(ctx context.Context, campaignID string) (*Conflict, error)
	Save(ctx context.Context, c Conflict) (Conflict, error)
	Close(ctx context.Context, campaignID string) error
	// Assign records the roll as spent; ErrAlreadyAssigned when it was.
	Assign(ctx context.Context, conflictID, rollID, objectiveID string, successes int) error
}

// Rolls is the part of rolls.Service used here.
type Rolls interface {
	Final(ctx context.Context, campaignID, rollID string) (rolls.Final, error)
	Threat(ctx context.Context, campaignID string, sender realtime.Sender, name string, pool rolls.ThreatPool) (symbols, error)
}

// Characters is the part of characters.Service used here.
type Characters interface {
	Get(ctx context.Context, m campaigns.Member, id string) (characters.View, error)
}

// Events is the part of realtime.Hub used to notify clients.
type Events interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the conflict rules; writes are serialized (single API instance).
type Service struct {
	repo   Repository
	rolls  Rolls
	chars  Characters
	events Events
	mu     sync.Mutex
}

// NewService builds a Service.
func NewService(repo Repository, r Rolls, chars Characters, ev Events) *Service {
	return &Service{repo: repo, rolls: r, chars: chars, events: ev}
}

func member(campaignID string, s realtime.Sender) campaigns.Member {
	return campaigns.Member{CampaignID: campaignID, UserID: s.UserID, Role: s.Role}
}

func (s *Service) checkNPC(ctx context.Context, m campaigns.Member, id *string) error {
	if id == nil {
		return nil
	}
	v, err := s.chars.Get(ctx, m, *id)
	if err != nil || v.Kind != characters.KindNPC {
		return bad("O personagem vinculado à ameaça precisa ser um NPC da campanha.")
	}
	return nil
}

// View returns the active conflict as the member may see it (nil when none).
func (s *Service) View(ctx context.Context, m campaigns.Member) (*Conflict, error) {
	c, err := s.repo.Active(ctx, m.CampaignID)
	if err != nil || c == nil {
		return nil, err
	}
	return forRole(*c, m.Role), nil
}

func forRole(c Conflict, role string) *Conflict {
	if role != campaigns.RoleGM {
		c = c.ForPlayers()
	}
	return &c
}

func (s *Service) publish(campaignID string, c *Conflict) {
	if c == nil {
		s.events.Broadcast(campaignID, realtime.ToAll, EventUpdated, map[string]any{"conflict": nil})
		return
	}
	s.events.Broadcast(campaignID, realtime.ToGM, EventUpdated, map[string]any{"conflict": c})
	s.events.Broadcast(campaignID, func(x realtime.Sender) bool { return !realtime.ToGM(x) }, EventUpdated,
		map[string]any{"conflict": c.ForPlayers()})
}

// Save creates or replaces the editable part of the active conflict (Mestre only, enforced by the guard).
func (s *Service) Save(ctx context.Context, campaignID string, sender realtime.Sender, in Input) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, err := s.repo.Active(ctx, campaignID)
	if err != nil {
		return err
	}
	c := Conflict{CampaignID: campaignID}
	if cur != nil {
		c = *cur
	}
	doc, title, err := s.normalize(ctx, member(campaignID, sender), in, c.Doc)
	if err != nil {
		return err
	}
	c.Doc, c.Title, c.SceneID = doc, title, in.SceneID
	if c.SceneID != nil && *c.SceneID == "" {
		c.SceneID = nil
	}
	saved, err := s.repo.Save(ctx, c)
	if err != nil {
		return err
	}
	s.publish(campaignID, &saved)
	return nil
}

// Close deactivates the active conflict.
func (s *Service) Close(ctx context.Context, campaignID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.repo.Close(ctx, campaignID); err != nil {
		return err
	}
	s.publish(campaignID, nil)
	return nil
}
