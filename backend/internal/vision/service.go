package vision

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Service computes per-player visibility. One mutex serializes everything so the caches never go stale.
type Service struct {
	repo   Repo
	scenes SceneSource
	hub    Hub

	mu    sync.Mutex
	cache map[string]Bits   // scene|user -> visible cells
	sent  map[string]string // scene|user -> last vision.updated pushed (mode + bitset)
}

// NewService builds the service.
func NewService(repo Repo, sc SceneSource, hub Hub) *Service {
	return &Service{repo: repo, scenes: sc, hub: hub, cache: map[string]Bits{}, sent: map[string]string{}}
}

func deny(msg string) error { return realtime.UserError(msg) }

// state is a consistent snapshot of a scene's inputs.
type state struct {
	campaignID string
	sc         scenes.Scene
	grid       *maps.Grid
	fog        Fog
	revealed   Bits
	pos        []Pos
}

func (s *Service) load(ctx context.Context, campaignID, sceneID, role string) (*state, error) {
	if !uuidRe.MatchString(sceneID) {
		return nil, ErrNotFound
	}
	sc, err := s.scenes.Get(ctx, campaignID, sceneID, role)
	if errors.Is(err, scenes.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	g, err := maps.Deserialize(sc.Grid)
	if err != nil {
		return nil, err
	}
	fog, err := s.repo.Fog(ctx, sceneID)
	if err != nil {
		return nil, err
	}
	pos, err := s.repo.Positions(ctx, sceneID)
	if err != nil {
		return nil, err
	}
	return &state{campaignID, sc, g, fog, DecodeBits(fog.Revealed, g.Width, g.Depth), pos}, nil
}

// visible returns the cells the player sees; callers hold s.mu.
func (s *Service) visible(st *state, userID string) Bits {
	key := st.sc.ID + "|" + userID
	if b, ok := s.cache[key]; ok {
		return b
	}
	b := NewBits(st.grid.Width, st.grid.Depth)
	b.Or(st.revealed)
	if st.fog.Mode == ModeDynamic {
		var from []maps.Coord
		for _, p := range st.pos {
			if p.OwnerID != nil && *p.OwnerID == userID {
				from = append(from, maps.Coord{X: p.X, Z: p.Z})
			}
		}
		set := maps.VisibleForMany(st.grid, st.sc.Objects, from, eyeHeight, st.fog.Radius)
		b.Or(FromSet(set, st.grid.Width, st.grid.Depth))
	}
	s.cache[key] = b
	return b
}

func (s *Service) invalidate(sceneID, userID string) {
	if userID != "" {
		delete(s.cache, sceneID+"|"+userID)
		return
	}
	for k := range s.cache {
		if strings.HasPrefix(k, sceneID+"|") {
			delete(s.cache, k)
		}
	}
}

// push sends the player their vision when it changed since the last push.
func (s *Service) push(st *state, userID string) {
	payload := map[string]any{"sceneId": st.sc.ID, "mode": st.fog.Mode}
	enc := ""
	if st.fog.Mode != ModeOff {
		enc = s.visible(st, userID).Encode()
		payload["visible"] = enc
	}
	key := st.sc.ID + "|" + userID
	if s.sent[key] == st.fog.Mode+enc {
		return
	}
	s.sent[key] = st.fog.Mode + enc
	s.hub.Broadcast(st.campaignID, realtime.ToUser(userID), EventVision, payload)
}

func (s *Service) pushAll(st *state) {
	for _, u := range s.hub.Online(st.campaignID) {
		if u.Role != campaigns.RoleGM {
			s.push(st, u.UserID)
		}
	}
}

func (s *Service) pushFog(st *state) {
	s.hub.Broadcast(st.campaignID, realtime.ToGM, EventFog, map[string]any{
		"sceneId": st.sc.ID, "mode": st.fog.Mode, "visionRadius": st.fog.Radius, "revealed": st.fog.Revealed,
	})
}

// changed recomputes after a fog edit and tells the Mestre and every online player.
func (s *Service) changed(st *state) {
	s.invalidate(st.sc.ID, "")
	s.pushFog(st)
	s.pushAll(st)
}

// Get returns the member's view of the scene's fog.
func (s *Service) Get(ctx context.Context, m campaigns.Member, sceneID string) (View, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(ctx, m.CampaignID, sceneID, m.Role)
	if err != nil {
		return View{}, err
	}
	v := View{SceneID: st.sc.ID, Mode: st.fog.Mode}
	if m.Role == campaigns.RoleGM {
		v.VisionRadius, v.Revealed = st.fog.Radius, st.fog.Revealed
	} else if st.fog.Mode != ModeOff {
		v.Visible = s.visible(st, m.UserID).Encode()
	}
	return v, nil
}

// Filter drops the tokens a player cannot see (their own always stay); the Mestre and mode off see all.
func (s *Service) Filter(ctx context.Context, m campaigns.Member, sceneID string, list []tokens.Token) []tokens.Token {
	if m.Role == campaigns.RoleGM {
		return list
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(ctx, m.CampaignID, sceneID, m.Role)
	if err != nil || st.fog.Mode == ModeOff {
		return list
	}
	vis := s.visible(st, m.UserID)
	out := make([]tokens.Token, 0, len(list))
	for _, t := range list {
		if (t.OwnerID != nil && *t.OwnerID == m.UserID) || vis.Has(t.X, t.Z) {
			out = append(out, t)
		}
	}
	return out
}
