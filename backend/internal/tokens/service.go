package tokens

import (
	"context"
	"errors"
	"math"
	"regexp"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Events emitted to clients.
const (
	EventPlaced  = "token.placed"
	EventRemoved = "token.removed"
	EventMoved   = "token.moved"
	EventPinged  = "map.pinged"
	EventFocused = "camera.focused"
)

// Notifier is the part of the realtime hub the service needs.
type Notifier interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the token rules. Moves are serialized so validation and write never interleave.
type Service struct {
	repo   Repo
	scenes SceneSource
	notif  Notifier
	mu     sync.Mutex
}

// NewService builds the service.
func NewService(repo Repo, sc SceneSource, notif Notifier) *Service {
	return &Service{repo: repo, scenes: sc, notif: notif}
}

func deny(msg string) error { return realtime.UserError(msg) }

func isGM(s realtime.Sender) bool { return s.Role == campaigns.RoleGM }

func owns(s realtime.Sender, ownerID *string) bool { return ownerID != nil && *ownerID == s.UserID }

// loadScene maps a missing scene to a user message.
func (s *Service) loadScene(ctx context.Context, campaignID, id string, snd realtime.Sender) (scenes.Scene, error) {
	if !uuidRe.MatchString(id) {
		return scenes.Scene{}, deny("Cena não encontrada")
	}
	sc, err := s.scenes.Get(ctx, campaignID, id, snd.Role)
	if errors.Is(err, scenes.ErrNotFound) {
		return scenes.Scene{}, deny("Cena não encontrada ou não está ativa")
	}
	return sc, err
}

// List returns the tokens of a scene visible to the member.
func (s *Service) List(ctx context.Context, m campaigns.Member, sceneID string) ([]Token, error) {
	if _, err := s.scenes.Get(ctx, m.CampaignID, sceneID, m.Role); err != nil {
		return nil, mapScene(err)
	}
	return s.repo.List(ctx, m.CampaignID, sceneID)
}

func mapScene(err error) error {
	if errors.Is(err, scenes.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func grid(sc scenes.Scene) (*maps.Grid, error) { return maps.Deserialize(sc.Grid) }

// Place puts a character on a scene: the Mestre any character, a player only their own on the active scene.
func (s *Service) Place(ctx context.Context, snd realtime.Sender, campaignID, sceneID, charID string, x, z int) error {
	if !uuidRe.MatchString(charID) {
		return deny("Personagem não encontrado")
	}
	sc, err := s.loadScene(ctx, campaignID, sceneID, snd)
	if err != nil {
		return err
	}
	ch, err := s.repo.Character(ctx, campaignID, charID)
	if errors.Is(err, ErrNotFound) {
		return deny("Personagem não encontrado")
	}
	if err != nil {
		return err
	}
	if err := s.checkPlace(snd, sc, ch, x, z); err != nil {
		return err
	}
	t, err := s.repo.Place(ctx, sc.ID, ch.ID, x, z)
	if errors.Is(err, ErrExists) {
		return deny("Este personagem já está na cena")
	}
	if err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, realtime.ToAll, EventPlaced, map[string]any{"token": t})
	return nil
}

func (s *Service) checkPlace(snd realtime.Sender, sc scenes.Scene, ch Char, x, z int) error {
	if !isGM(snd) && (!owns(snd, ch.OwnerID) || !sc.IsActive) {
		return deny("Você só pode colocar o token do seu personagem na cena ativa")
	}
	return s.checkCell(sc, x, z)
}

func (s *Service) checkCell(sc scenes.Scene, x, z int) error {
	g, err := grid(sc)
	if err != nil {
		return err
	}
	c := maps.Coord{X: x, Z: z}
	if ok, why := maps.CanEnter(g, sc.Objects, c, c, 0); !ok {
		return deny(why)
	}
	return nil
}

// Remove takes a token off its scene (Mestre only; enforced by the command guard too).
func (s *Service) Remove(ctx context.Context, snd realtime.Sender, campaignID, tokenID string) error {
	if !isGM(snd) {
		return deny("Você não tem permissão para fazer isso")
	}
	t, err := s.token(ctx, campaignID, tokenID)
	if err != nil {
		return err
	}
	if err := s.repo.Remove(ctx, t.ID); err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, realtime.ToAll, EventRemoved, map[string]string{"tokenId": t.ID, "sceneId": t.SceneID})
	return nil
}

func (s *Service) token(ctx context.Context, campaignID, id string) (Token, error) {
	if !uuidRe.MatchString(id) {
		return Token{}, deny("Token não encontrado")
	}
	t, err := s.repo.Get(ctx, campaignID, id)
	if errors.Is(err, ErrNotFound) {
		return Token{}, deny("Token não encontrado")
	}
	return t, err
}

// Move validates and applies a move; the Mestre may skip the path rules with force.
func (s *Service) Move(ctx context.Context, snd realtime.Sender, campaignID, tokenID string, x, z int, force bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, err := s.token(ctx, campaignID, tokenID)
	if err != nil {
		return err
	}
	if !isGM(snd) && !owns(snd, t.OwnerID) {
		return deny("Você só pode mover os tokens dos seus personagens")
	}
	sc, err := s.loadScene(ctx, campaignID, t.SceneID, snd)
	if err != nil {
		return err
	}
	if err := s.checkMove(sc, t, x, z, force && isGM(snd)); err != nil {
		return err
	}
	if err := s.repo.Move(ctx, t.ID, x, z); err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, realtime.ToAll, EventMoved,
		map[string]any{"tokenId": t.ID, "x": x, "z": z, "by": snd.UserID})
	return nil
}

func (s *Service) checkMove(sc scenes.Scene, t Token, x, z int, force bool) error {
	g, err := grid(sc)
	if err != nil {
		return err
	}
	to := maps.Coord{X: x, Z: z}
	if !g.InBounds(x, z) {
		return deny("Destino fora do mapa.")
	}
	if force {
		return nil
	}
	if ok, why := maps.CanEnter(g, sc.Objects, to, to, 0); !ok {
		return deny(why)
	}
	if maps.PathBlocked(g, sc.Objects, maps.Coord{X: t.X, Z: t.Z}, to) {
		return deny("O caminho até lá está bloqueado.")
	}
	return nil
}

// Ping shows a marker to everyone.
func (s *Service) Ping(campaignID string, snd realtime.Sender, sceneID string, x, z float64) error {
	if err := checkPoint(sceneID, x, z); err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, realtime.ToAll, EventPinged,
		map[string]any{"sceneId": sceneID, "x": x, "z": z, "userName": snd.Name})
	return nil
}

// FocusAll moves everyone's camera (the command guard restricts it to the Mestre).
func (s *Service) FocusAll(campaignID string, snd realtime.Sender, sceneID string, x, z float64) error {
	if !isGM(snd) {
		return deny("Você não tem permissão para fazer isso")
	}
	if err := checkPoint(sceneID, x, z); err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, realtime.ToAll, EventFocused, map[string]any{"sceneId": sceneID, "x": x, "z": z})
	return nil
}

func checkPoint(sceneID string, x, z float64) error {
	bad := func(v float64) bool { return math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1000 }
	if !uuidRe.MatchString(sceneID) || bad(x) || bad(z) {
		return deny("Posição inválida")
	}
	return nil
}
