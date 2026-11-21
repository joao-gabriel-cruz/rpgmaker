package scenes

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Events emitted to clients.
const (
	EventActivated      = "scene.activated"
	EventListChanged    = "scenes.changed"
	EventPatched        = "map.patched"
	EventObjectsPatched = "map.objects.patched"
)

// Notifier is the part of the realtime hub the service needs.
type Notifier interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the scene rules. Edits are serialized per scene so concurrent patches are never lost.
type Service struct {
	repo  Repo
	notif Notifier
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// NewService builds the service.
func NewService(repo Repo, notif Notifier) *Service {
	return &Service{repo: repo, notif: notif, locks: map[string]*sync.Mutex{}}
}

func (s *Service) lockFor(id string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.locks[id]
	if !ok {
		l = &sync.Mutex{}
		s.locks[id] = l
	}
	return l
}

func (s *Service) changed(campaignID string) {
	s.notif.Broadcast(campaignID, realtime.ToAll, EventListChanged, struct{}{})
}

// Create makes an inactive, empty scene (floor, height 1).
func (s *Service) Create(ctx context.Context, campaignID, name string, w, d int) (Scene, error) {
	name, err := cleanName(name)
	if err != nil {
		return Scene{}, err
	}
	if w < MinSize || w > MaxSize || d < MinSize || d > MaxSize {
		return Scene{}, ErrInvalidSize
	}
	g, err := maps.NewGrid(w, d)
	if err != nil {
		return Scene{}, ErrInvalidSize
	}
	sc, err := s.repo.Create(ctx, campaignID, name, w, d, g.Serialize().Data)
	if err == nil {
		s.changed(campaignID)
	}
	return sc, err
}

func cleanName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxNameLen {
		return "", ErrInvalidName
	}
	return name, nil
}

// List returns every scene's summary (names and which one is active) to any member.
func (s *Service) List(ctx context.Context, campaignID string) ([]Summary, error) {
	return s.repo.List(ctx, campaignID)
}

// Get returns a scene; players may only load the active one.
func (s *Service) Get(ctx context.Context, campaignID, id, role string) (Scene, error) {
	if !uuidRe.MatchString(id) {
		return Scene{}, ErrNotFound
	}
	sc, err := s.repo.Get(ctx, campaignID, id)
	if err != nil {
		return Scene{}, err
	}
	if role != campaigns.RoleGM && !sc.IsActive {
		return Scene{}, ErrNotFound
	}
	return sc, nil
}

// Rename changes a scene's name.
func (s *Service) Rename(ctx context.Context, campaignID, id, name string) error {
	name, err := cleanName(name)
	if err != nil {
		return err
	}
	if !uuidRe.MatchString(id) {
		return ErrNotFound
	}
	if err := s.repo.Rename(ctx, campaignID, id, name); err != nil {
		return err
	}
	s.changed(campaignID)
	return nil
}

// Delete removes a scene (the active one included: the campaign is then left without one).
func (s *Service) Delete(ctx context.Context, campaignID, id string) error {
	if !uuidRe.MatchString(id) {
		return ErrNotFound
	}
	l := s.lockFor(id)
	l.Lock()
	defer l.Unlock()
	if err := s.repo.Delete(ctx, campaignID, id); err != nil {
		return err
	}
	s.changed(campaignID)
	return nil
}

// Activate makes the scene the only active one and tells everyone.
func (s *Service) Activate(ctx context.Context, campaignID, id string) error {
	if !uuidRe.MatchString(id) {
		return ErrNotFound
	}
	if err := s.repo.Activate(ctx, campaignID, id); err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, realtime.ToAll, EventActivated, map[string]string{"sceneId": id})
	s.changed(campaignID)
	return nil
}
