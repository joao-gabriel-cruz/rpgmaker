package actions

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
)

// Catalog merges the defaults with the campaign's overrides and custom actions; players get enabled ones only.
func (s *Service) Catalog(ctx context.Context, campaignID string, all bool) ([]Action, error) {
	stored, err := s.store.List(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	byID := map[string]Action{}
	for _, a := range stored {
		byID[a.ID] = a
	}
	out := make([]Action, 0, len(defaults)+len(stored))
	for _, d := range defaults {
		if o, ok := byID[d.ID]; ok {
			delete(byID, d.ID)
			o.Builtin = true
			out = append(out, o)
			continue
		}
		d.Enabled, d.Builtin = true, true
		out = append(out, d)
	}
	for _, a := range stored {
		if _, custom := byID[a.ID]; custom {
			out = append(out, a)
		}
	}
	return filterEnabled(out, all), nil
}

func filterEnabled(list []Action, all bool) []Action {
	if all {
		return list
	}
	out := list[:0:0]
	for _, a := range list {
		if a.Enabled {
			out = append(out, a)
		}
	}
	return out
}

// Find returns one enabled action of the campaign.
func (s *Service) Find(ctx context.Context, campaignID, id string) (Action, error) {
	list, err := s.Catalog(ctx, campaignID, false)
	if err != nil {
		return Action{}, err
	}
	for _, a := range list {
		if a.ID == id {
			return a, nil
		}
	}
	return Action{}, ErrNotFound
}

// Create stores a new custom action.
func (s *Service) Create(ctx context.Context, campaignID string, in Action) (Action, error) {
	a, err := Clean(in)
	if err != nil {
		return Action{}, err
	}
	var b [6]byte
	if _, err := crand.Read(b[:]); err != nil {
		return Action{}, err
	}
	a.ID, a.Builtin = "custom-"+hex.EncodeToString(b[:]), false
	return a, s.store.Save(ctx, campaignID, a)
}

// Update replaces a default (as an override) or a custom action.
func (s *Service) Update(ctx context.Context, campaignID, id string, in Action) (Action, error) {
	a, err := Clean(in)
	if err != nil {
		return Action{}, err
	}
	_, builtin := defaultByID(id)
	if !builtin && !s.exists(ctx, campaignID, id) {
		return Action{}, ErrNotFound
	}
	a.ID, a.Builtin = id, builtin
	return a, s.store.Save(ctx, campaignID, a)
}

func (s *Service) exists(ctx context.Context, campaignID, id string) bool {
	list, err := s.store.List(ctx, campaignID)
	if err != nil {
		return false
	}
	for _, a := range list {
		if a.ID == id {
			return true
		}
	}
	return false
}

// Delete removes a custom action or restores a default to its system definition.
func (s *Service) Delete(ctx context.Context, campaignID, id string) error {
	if _, builtin := defaultByID(id); builtin && !s.exists(ctx, campaignID, id) {
		return nil
	}
	return s.store.Delete(ctx, campaignID, id)
}
