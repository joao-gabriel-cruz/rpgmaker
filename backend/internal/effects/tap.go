package effects

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

// Notifier is what the tokens service broadcasts through.
type Notifier interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Tap decorates a Notifier: it forwards everything, then triggers on_enter for token placements and moves.
type Tap struct {
	inner Notifier
	svc   *Service
}

// NewTap wraps inner (the vision tap) so token movements are observed without changing what is sent.
func NewTap(inner Notifier, svc *Service) *Tap { return &Tap{inner: inner, svc: svc} }

type moveEvent struct {
	TokenID string        `json:"tokenId"`
	X       int           `json:"x"`
	Z       int           `json:"z"`
	FromX   int           `json:"fromX"`
	FromZ   int           `json:"fromZ"`
	Token   *tokens.Token `json:"token"`
}

// Broadcast implements the tokens service's Notifier.
func (t *Tap) Broadcast(campaignID string, to realtime.Filter, typ string, payload any) {
	t.inner.Broadcast(campaignID, to, typ, payload)
	if typ != tokens.EventMoved && typ != tokens.EventPlaced {
		return
	}
	var ev moveEvent
	raw, err := json.Marshal(payload)
	if err != nil || json.Unmarshal(raw, &ev) != nil {
		return
	}
	ctx := context.Background()
	if typ == tokens.EventPlaced && ev.Token != nil {
		t.svc.OnEntered(ctx, campaignID, ev.Token.ID, ev.Token.X, ev.Token.Z, nil)
	} else if typ == tokens.EventMoved {
		t.svc.OnEntered(ctx, campaignID, ev.TokenID, ev.X, ev.Z, &maps.Coord{X: ev.FromX, Z: ev.FromZ})
	}
}
