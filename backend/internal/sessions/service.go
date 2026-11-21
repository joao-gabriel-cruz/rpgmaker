package sessions

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// EventUpdated is broadcast to everyone when a session starts or ends.
const EventUpdated = "session.updated"

// State is the session state sent to clients.
type State struct {
	Active    bool       `json:"active"`
	StartedAt *time.Time `json:"startedAt"`
}

// Events is the part of realtime.Hub used to notify clients.
type Events interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the session rules.
type Service struct {
	store  Store
	events Events
}

// NewService builds the service.
func NewService(store Store, ev Events) *Service { return &Service{store: store, events: ev} }

// ActiveID returns the running session id, or "" when there is none.
func (s *Service) ActiveID(ctx context.Context, campaignID string) (string, error) {
	ses, err := s.store.Active(ctx, campaignID)
	if errors.Is(err, ErrNoneActive) {
		return "", nil
	}
	return ses.ID, err
}

// State returns the current state.
func (s *Service) State(ctx context.Context, campaignID string) (State, error) {
	ses, err := s.store.Active(ctx, campaignID)
	if errors.Is(err, ErrNoneActive) {
		return State{}, nil
	}
	return State{Active: err == nil, StartedAt: &ses.StartedAt}, err
}

// Start opens a session and tells the table.
func (s *Service) Start(ctx context.Context, campaignID string) error {
	if _, err := s.store.Start(ctx, campaignID); err != nil {
		return err
	}
	return s.publish(ctx, campaignID)
}

// End closes the session and tells the table.
func (s *Service) End(ctx context.Context, campaignID string) error {
	if err := s.store.End(ctx, campaignID); err != nil {
		return err
	}
	return s.publish(ctx, campaignID)
}

func (s *Service) publish(ctx context.Context, campaignID string) error {
	st, err := s.State(ctx, campaignID)
	if err == nil {
		s.events.Broadcast(campaignID, realtime.ToAll, EventUpdated, st)
	}
	return err
}

func userFacing(err error) error {
	switch {
	case errors.Is(err, ErrAlreadyActive):
		return realtime.UserError("Já existe uma sessão em andamento.")
	case errors.Is(err, ErrNoneActive):
		return realtime.UserError("Nenhuma sessão em andamento.")
	}
	return err
}

// Register wires session.start/session.end (GM only) and GET /api/campaigns/{id}/session.
func Register(mux *http.ServeMux, hub *realtime.Hub, svc *Service, requireMember func(http.Handler) http.Handler) {
	cmds := map[string]func(context.Context, string) error{"session.start": svc.Start, "session.end": svc.End}
	for typ, fn := range cmds {
		hub.Handle(typ, func(ctx context.Context, cmd realtime.Command) error {
			return userFacing(fn(ctx, cmd.CampaignID))
		}, realtime.GMOnly)
	}
	mux.Handle("GET /api/campaigns/{id}/session", requireMember(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := campaigns.MemberFrom(r)
		st, err := svc.State(r.Context(), m.CampaignID)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Erro interno. Tente novamente"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(st)
	})))
}
