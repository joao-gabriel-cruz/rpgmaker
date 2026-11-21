package scenes

import (
	"context"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// ErrDraftState reports a draft command invalid for the scene's current state.
var ErrDraftState = errors.New("scenes: draft state")

// working returns the scene as the Mestre edits it: the draft when one exists, else the real state.
func working(sc Scene) Scene {
	if sc.Draft != nil {
		sc.Grid, sc.Objects = sc.Draft.Grid, sc.Draft.Objects
	}
	return sc
}

// editAudience: draft edits reach only other Mestre sessions; real edits follow audience.
func editAudience(sc Scene, origin string) realtime.Filter {
	if sc.Draft == nil {
		return audience(sc, origin)
	}
	return func(s realtime.Sender) bool { return realtime.ToGM(s) && (origin == "" || s.UserID != origin) }
}

func eventName(sc Scene, live, draft string) string {
	if sc.Draft != nil {
		return draft
	}
	return live
}

func (s *Service) saveGrid(ctx context.Context, sc Scene, data string) error {
	if sc.Draft == nil {
		return s.repo.SaveGrid(ctx, sc.ID, data)
	}
	d := *sc.Draft
	d.Grid.Data = data
	return s.repo.SaveDraft(ctx, sc.ID, &d)
}

func (s *Service) saveObjects(ctx context.Context, sc Scene, objs []maps.Object) error {
	if sc.Draft == nil {
		return s.repo.SaveObjects(ctx, sc.ID, objs)
	}
	d := *sc.Draft
	d.Objects = objs
	return s.repo.SaveDraft(ctx, sc.ID, &d)
}

func (s *Service) draftChanged(campaignID, sceneID string, active bool) {
	s.notif.Broadcast(campaignID, realtime.ToGM, EventDraftChanged, map[string]any{"sceneId": sceneID, "active": active})
}

// StartDraft copies the real state into a hidden draft; later Mestre edits go to it.
func (s *Service) StartDraft(ctx context.Context, campaignID, sceneID string) error {
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
	d := &Draft{Grid: sc.Grid, Objects: append([]maps.Object{}, sc.Objects...)}
	if err := s.repo.SaveDraft(ctx, sceneID, d); err != nil {
		return err
	}
	s.draftChanged(campaignID, sceneID, true)
	return nil
}

// DiscardDraft drops the draft; players never knew about it.
func (s *Service) DiscardDraft(ctx context.Context, campaignID, sceneID string) error {
	l := s.lockFor(sceneID)
	l.Lock()
	defer l.Unlock()
	sc, err := s.load(ctx, campaignID, sceneID)
	if err != nil {
		return err
	}
	if sc.Draft == nil {
		return ErrDraftState
	}
	if err := s.repo.SaveDraft(ctx, sceneID, nil); err != nil {
		return err
	}
	s.draftChanged(campaignID, sceneID, false)
	return nil
}

// ApplyDraft makes the draft the real state and sends everyone one consolidated map.replaced.
func (s *Service) ApplyDraft(ctx context.Context, campaignID, sceneID string) error {
	l := s.lockFor(sceneID)
	l.Lock()
	defer l.Unlock()
	sc, err := s.load(ctx, campaignID, sceneID)
	if err != nil {
		return err
	}
	if sc.Draft == nil {
		return ErrDraftState
	}
	return s.replace(ctx, sc, sc.Draft.Grid.Data, sc.Draft.Objects)
}

func (s *Service) replace(ctx context.Context, sc Scene, grid string, objects []maps.Object) error {
	if err := s.repo.Replace(ctx, sc.ID, grid, objects); err != nil {
		return err
	}
	if objects == nil {
		objects = []maps.Object{}
	}
	payload := map[string]any{
		"sceneId": sc.ID,
		"grid":    maps.Serialized{Width: sc.Width, Depth: sc.Depth, Data: grid},
		"objects": objects,
	}
	s.notif.Broadcast(sc.CampaignID, audience(sc, ""), EventReplaced, payload)
	return nil
}
