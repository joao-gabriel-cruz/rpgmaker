package vision

import (
	"encoding/json"
	"log/slog"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

// Tap is a realtime notifier for the scenes and tokens services: it forwards every event to the hub,
// filters token events per player by vision and recomputes visibility when terrain, objects or tokens change.
type Tap struct{ svc *Service }

// Tap returns the notifier to hand to the scenes and tokens services.
func (s *Service) Tap() *Tap { return &Tap{svc: s} }

type event struct {
	SceneID string        `json:"sceneId"`
	TokenID string        `json:"tokenId"`
	X       int           `json:"x"`
	Z       int           `json:"z"`
	FromX   int           `json:"fromX"`
	FromZ   int           `json:"fromZ"`
	Token   *tokens.Token `json:"token"`
}

func parse(payload any) event {
	var ev event
	raw, err := json.Marshal(payload)
	if err == nil {
		_ = json.Unmarshal(raw, &ev)
	}
	return ev
}

// Broadcast implements the services' Notifier.
func (t *Tap) Broadcast(campaignID string, to realtime.Filter, typ string, payload any) {
	s := t.svc
	switch typ {
	case tokens.EventMoved:
		s.onMoved(campaignID, payload)
	case tokens.EventPlaced:
		s.onPlaced(campaignID, payload)
	case tokens.EventRemoved, scenes.EventPatched, scenes.EventObjectsPatched, scenes.EventReplaced:
		s.hub.Broadcast(campaignID, to, typ, payload)
		s.refresh(campaignID, parse(payload).SceneID)
	default:
		s.hub.Broadcast(campaignID, to, typ, payload)
	}
}

// refresh recomputes the scene for everyone after terrain, objects or tokens changed.
func (s *Service) refresh(campaignID, sceneID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(bg(), campaignID, sceneID, campaigns.RoleGM)
	if err != nil {
		return
	}
	s.invalidate(sceneID, "")
	if st.fog.Mode != ModeOff {
		s.pushAll(st)
	}
}

func (s *Service) others(st *state, owner string) []realtime.Sender {
	var out []realtime.Sender
	for _, u := range s.hub.Online(st.campaignID) {
		if u.Role != campaigns.RoleGM && u.UserID != owner {
			out = append(out, u)
		}
	}
	return out
}

func owner(o *string) string {
	if o == nil {
		return ""
	}
	return *o
}

func (s *Service) ownerOf(st *state, tokenID string) string {
	for _, p := range st.pos {
		if p.ID == tokenID {
			return owner(p.OwnerID)
		}
	}
	return ""
}

// guarded loads the scene for a token event; ok=false means the event was already handled (fail closed).
func (s *Service) guarded(campaignID, typ, sceneID string, payload any) (*state, bool) {
	st, err := s.load(bg(), campaignID, sceneID, campaigns.RoleGM)
	if err != nil {
		slog.Error("vision load failed; token event sent to the Mestre only", "err", err)
		s.hub.Broadcast(campaignID, realtime.ToGM, typ, payload)
		return nil, false
	}
	if st.fog.Mode == ModeOff {
		s.hub.Broadcast(campaignID, realtime.ToAll, typ, payload)
		return nil, false
	}
	return st, true
}

func (s *Service) onMoved(campaignID string, payload any) {
	ev := parse(payload)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.guarded(campaignID, tokens.EventMoved, ev.SceneID, payload)
	if !ok {
		return
	}
	own := s.ownerOf(st, ev.TokenID)
	s.invalidate(st.sc.ID, own)
	s.hub.Broadcast(campaignID, realtime.Or(realtime.ToGM, realtime.ToUser(own)), tokens.EventMoved, payload)
	gone := map[string]string{"tokenId": ev.TokenID, "sceneId": ev.SceneID}
	for _, u := range s.others(st, own) {
		vis := s.visible(st, u.UserID)
		if vis.Has(ev.X, ev.Z) {
			s.hub.Broadcast(campaignID, realtime.ToUser(u.UserID), tokens.EventMoved, payload)
		} else if vis.Has(ev.FromX, ev.FromZ) {
			s.hub.Broadcast(campaignID, realtime.ToUser(u.UserID), tokens.EventRemoved, gone)
		}
	}
	if own != "" {
		s.push(st, own)
	}
}

func (s *Service) onPlaced(campaignID string, payload any) {
	ev := parse(payload)
	if ev.Token == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.guarded(campaignID, tokens.EventPlaced, ev.Token.SceneID, payload)
	if !ok {
		return
	}
	own := owner(ev.Token.OwnerID)
	s.invalidate(st.sc.ID, own)
	s.hub.Broadcast(campaignID, realtime.Or(realtime.ToGM, realtime.ToUser(own)), tokens.EventPlaced, payload)
	for _, u := range s.others(st, own) {
		if s.visible(st, u.UserID).Has(ev.Token.X, ev.Token.Z) {
			s.hub.Broadcast(campaignID, realtime.ToUser(u.UserID), tokens.EventPlaced, payload)
		}
	}
	if own != "" {
		s.push(st, own)
	}
}
