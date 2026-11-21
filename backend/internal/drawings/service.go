package drawings

import (
	"context"
	"errors"
	"math"
	"regexp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

var (
	uuidRe  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	colorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

// Events emitted to clients.
const (
	EventAdded   = "drawing.added"
	EventRemoved = "drawing.removed"
	EventCleared = "drawings.cleared"
	EventRuler   = "ruler.moved"
)

// Notifier is the part of the realtime hub the service needs.
type Notifier interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the drawing rules.
type Service struct {
	repo   Repo
	scenes SceneSource
	notif  Notifier
}

// NewService builds the service.
func NewService(repo Repo, sc SceneSource, notif Notifier) *Service {
	return &Service{repo: repo, scenes: sc, notif: notif}
}

func deny(msg string) error { return realtime.UserError(msg) }

func isGM(s realtime.Sender) bool { return s.Role == campaigns.RoleGM }

func (s *Service) checkScene(ctx context.Context, campaignID, id string, snd realtime.Sender) error {
	if !uuidRe.MatchString(id) {
		return deny("Cena não encontrada")
	}
	if _, err := s.scenes.Get(ctx, campaignID, id, snd.Role); err != nil {
		if errors.Is(err, scenes.ErrNotFound) {
			return deny("Cena não encontrada ou não está ativa")
		}
		return err
	}
	return nil
}

// List returns the strokes of a scene visible to the member.
func (s *Service) List(ctx context.Context, m campaigns.Member, sceneID string) ([]Drawing, error) {
	if _, err := s.scenes.Get(ctx, m.CampaignID, sceneID, m.Role); err != nil {
		if errors.Is(err, scenes.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.repo.List(ctx, m.CampaignID, sceneID)
}

func validPoint(p Point) bool {
	bad := func(v float64) bool { return math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > maxCoord }
	return !bad(p.X) && !bad(p.Z)
}

func validStroke(color string, width int, pts []Point) error {
	if !colorRe.MatchString(color) {
		return deny("Cor inválida")
	}
	if width < MinWidth || width > MaxWidth {
		return deny("Espessura inválida")
	}
	if len(pts) < 2 {
		return deny("O traço é curto demais")
	}
	if len(pts) > MaxPoints {
		return deny("O traço tem pontos demais")
	}
	for _, p := range pts {
		if !validPoint(p) {
			return deny("Posição inválida")
		}
	}
	return nil
}

// Add stores a stroke and tells everyone.
func (s *Service) Add(ctx context.Context, snd realtime.Sender, campaignID, sceneID, color string, width int, pts []Point) error {
	if err := s.checkScene(ctx, campaignID, sceneID, snd); err != nil {
		return err
	}
	if err := validStroke(color, width, pts); err != nil {
		return err
	}
	d, err := s.repo.Add(ctx, Drawing{SceneID: sceneID, UserID: snd.UserID, Color: color, Width: width, Points: pts})
	if err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, realtime.ToAll, EventAdded, map[string]any{"drawing": d})
	return nil
}

// Remove deletes a stroke: its author or the Mestre.
func (s *Service) Remove(ctx context.Context, snd realtime.Sender, campaignID, id string) error {
	if !uuidRe.MatchString(id) {
		return deny("Desenho não encontrado")
	}
	d, err := s.repo.Get(ctx, campaignID, id)
	if errors.Is(err, ErrNotFound) {
		return deny("Desenho não encontrado")
	}
	if err != nil {
		return err
	}
	if !isGM(snd) && d.UserID != snd.UserID {
		return deny("Você só pode apagar os seus desenhos")
	}
	if err := s.repo.Remove(ctx, d.ID); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	s.notif.Broadcast(campaignID, realtime.ToAll, EventRemoved, map[string]string{"drawingId": d.ID, "sceneId": d.SceneID})
	return nil
}

// Clear wipes every stroke of a scene (Mestre only).
func (s *Service) Clear(ctx context.Context, snd realtime.Sender, campaignID, sceneID string) error {
	if !isGM(snd) {
		return deny("Você não tem permissão para fazer isso")
	}
	if err := s.checkScene(ctx, campaignID, sceneID, snd); err != nil {
		return err
	}
	if err := s.repo.Clear(ctx, sceneID); err != nil {
		return err
	}
	s.notif.Broadcast(campaignID, realtime.ToAll, EventCleared, map[string]string{"sceneId": sceneID})
	return nil
}

// RulerMove relays a measurement in progress to the others; to == nil means the ruler was released. Never persisted.
func (s *Service) RulerMove(campaignID string, snd realtime.Sender, sceneID string, from, to *Point) error {
	if !uuidRe.MatchString(sceneID) || (to != nil && (from == nil || !validPoint(*from) || !validPoint(*to))) {
		return deny("Posição inválida")
	}
	s.notif.Broadcast(campaignID, realtime.ExceptConn(snd.ConnID), EventRuler, map[string]any{
		"sceneId": sceneID, "userId": snd.UserID, "userName": snd.Name, "from": from, "to": to,
	})
	return nil
}
