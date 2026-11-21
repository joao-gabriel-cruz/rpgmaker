package rounds

import (
	"context"
	"sync"
)

// PhaseChangeFunc reacts to a state transition (start, next, prev, end) after it was persisted.
type PhaseChangeFunc func(ctx context.Context, campaignID string, prev, next State)

// Service applies round transitions and notifies hooks.
type Service struct {
	store Store
	mu    sync.Mutex
	hooks []PhaseChangeFunc
}

// NewService builds the service over a store.
func NewService(store Store) *Service { return &Service{store: store} }

// OnPhaseChange registers a hook called synchronously after every transition; register before serving.
func (s *Service) OnPhaseChange(fn PhaseChangeFunc) { s.hooks = append(s.hooks, fn) }

// Get returns the current state.
func (s *Service) Get(ctx context.Context, campaignID string) (State, error) {
	return s.store.Get(ctx, campaignID)
}

// Start begins a conflict at round 1.
func (s *Service) Start(ctx context.Context, campaignID string) (State, error) {
	return s.apply(ctx, campaignID, func(State) (State, error) { return Start(), nil })
}

// Next advances the phase.
func (s *Service) Next(ctx context.Context, campaignID string) (State, error) {
	return s.apply(ctx, campaignID, State.Next)
}

// Prev goes back one phase.
func (s *Service) Prev(ctx context.Context, campaignID string) (State, error) {
	return s.apply(ctx, campaignID, State.Prev)
}

// End closes the conflict.
func (s *Service) End(ctx context.Context, campaignID string) (State, error) {
	return s.apply(ctx, campaignID, func(State) (State, error) { return Idle(), nil })
}

func (s *Service) apply(ctx context.Context, campaignID string, fn func(State) (State, error)) (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev, err := s.store.Get(ctx, campaignID)
	if err != nil {
		return prev, err
	}
	next, err := fn(prev)
	if err != nil {
		return prev, err
	}
	if err := s.store.Save(ctx, campaignID, next); err != nil {
		return prev, err
	}
	for _, h := range s.hooks {
		h(ctx, campaignID, prev, next)
	}
	return next, nil
}
