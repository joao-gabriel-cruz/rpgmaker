package ambience

import (
	"context"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

// Event names broadcast to every member.
const (
	EventUpdated = "ambience.updated"
	EventSaid    = "chat.said"
	EventReacted = "chat.reacted"
)

// Scenes is the part of scenes.Service used to check scene ownership.
type Scenes interface {
	Get(ctx context.Context, campaignID, id, role string) (scenes.Scene, error)
}

// Characters is the part of characters.Service used by conditions and chat.
type Characters interface {
	Get(ctx context.Context, m campaigns.Member, id string) (characters.View, error)
	EditSheet(ctx context.Context, campaignID, id string, fn SheetEdit) (characters.View, error)
}

// Events is the part of the realtime hub used to notify clients.
type Events interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the ambience, conditions and chat rules.
type Service struct {
	store  Store
	scenes Scenes
	chars  Characters
	events Events
}

// NewService builds a Service.
func NewService(store Store, sc Scenes, chars Characters, ev Events) *Service {
	return &Service{store: store, scenes: sc, chars: chars, events: ev}
}

func (s *Service) checkScene(ctx context.Context, campaignID, sceneID string) error {
	_, err := s.scenes.Get(ctx, campaignID, sceneID, campaigns.RoleGM)
	return err
}

// Get returns the ambience of a scene of the campaign.
func (s *Service) Get(ctx context.Context, campaignID, sceneID string) (Ambience, error) {
	if err := s.checkScene(ctx, campaignID, sceneID); err != nil {
		return Ambience{}, err
	}
	return s.store.Get(ctx, sceneID)
}

// Set validates, persists and announces a new ambience (Mestre only, enforced by the command guard).
func (s *Service) Set(ctx context.Context, campaignID string, a Ambience) error {
	if !oneOf(timesOfDay, a.TimeOfDay) || !oneOf(weathers, a.Weather) {
		return realtime.UserError("Ambiente inválido")
	}
	if err := s.checkScene(ctx, campaignID, a.SceneID); err != nil {
		if errors.Is(err, scenes.ErrNotFound) {
			return realtime.UserError("Cena não encontrada")
		}
		return err
	}
	if err := s.store.Save(ctx, a); err != nil {
		return err
	}
	s.events.Broadcast(campaignID, realtime.ToAll, EventUpdated, a)
	return nil
}
