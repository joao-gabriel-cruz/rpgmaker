package scenes

import (
	"context"
)

const maxSnapshotsPerScene = 50

// SaveSnapshot stores the scene's current real state under a name.
func (s *Service) SaveSnapshot(ctx context.Context, campaignID, sceneID, name string) (Snapshot, error) {
	name, err := cleanName(name)
	if err != nil {
		return Snapshot{}, err
	}
	sc, err := s.load(ctx, campaignID, sceneID)
	if err != nil {
		return Snapshot{}, err
	}
	list, err := s.repo.ListSnapshots(ctx, sceneID)
	if err != nil {
		return Snapshot{}, err
	}
	if len(list) >= maxSnapshotsPerScene {
		return Snapshot{}, ErrTooManySnapshots
	}
	return s.repo.CreateSnapshot(ctx, sceneID, name, sc.Grid.Data, sc.Objects)
}

// ListSnapshots returns the scene's saved states, newest first.
func (s *Service) ListSnapshots(ctx context.Context, campaignID, sceneID string) ([]Snapshot, error) {
	if _, err := s.load(ctx, campaignID, sceneID); err != nil {
		return nil, err
	}
	return s.repo.ListSnapshots(ctx, sceneID)
}

// DeleteSnapshot removes a saved state.
func (s *Service) DeleteSnapshot(ctx context.Context, campaignID, sceneID, id string) error {
	if _, err := s.load(ctx, campaignID, sceneID); err != nil {
		return err
	}
	if !uuidRe.MatchString(id) {
		return ErrNotFound
	}
	return s.repo.DeleteSnapshot(ctx, sceneID, id)
}

// ApplySnapshot turns the scene into the snapshot and broadcasts map.replaced; tokens are untouched.
func (s *Service) ApplySnapshot(ctx context.Context, campaignID, sceneID, id string) error {
	l := s.lockFor(sceneID)
	l.Lock()
	defer l.Unlock()
	sc, err := s.load(ctx, campaignID, sceneID)
	if err != nil {
		return err
	}
	if sc.Draft != nil {
		return ErrDraftState
	}
	if !uuidRe.MatchString(id) {
		return ErrNotFound
	}
	sn, err := s.repo.GetSnapshot(ctx, sceneID, id)
	if err != nil {
		return err
	}
	return s.replace(ctx, sc, sn.Grid, sn.Objects)
}
