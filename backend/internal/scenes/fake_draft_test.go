package scenes

import (
	"context"
	"fmt"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func (f *fakeRepo) SaveDraft(_ context.Context, id string, d *Draft) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if d == nil {
		f.scenes[id].Draft = nil
		return nil
	}
	cp := *d
	f.scenes[id].Draft = &cp
	return nil
}

func (f *fakeRepo) Replace(_ context.Context, id, grid string, objects []maps.Object) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	sc := f.scenes[id]
	sc.Grid.Data, sc.Objects, sc.Draft = grid, objects, nil
	return nil
}

func (f *fakeRepo) CreateSnapshot(_ context.Context, sceneID, name, grid string, objects []maps.Object) (Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	sn := Snapshot{ID: fmt.Sprintf("00000000-0000-0000-0001-%012d", len(f.snaps)+1), SceneID: sceneID, Name: name, CreatedAt: time.Now()}
	f.snaps[sn.ID] = SnapshotData{Snapshot: sn, Grid: grid, Objects: append([]maps.Object{}, objects...)}
	return sn, nil
}

func (f *fakeRepo) ListSnapshots(_ context.Context, sceneID string) ([]Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Snapshot{}
	for _, d := range f.snaps {
		if d.SceneID == sceneID {
			out = append(out, d.Snapshot)
		}
	}
	return out, nil
}

func (f *fakeRepo) GetSnapshot(_ context.Context, sceneID, id string) (SnapshotData, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d, ok := f.snaps[id]
	if !ok || d.SceneID != sceneID {
		return SnapshotData{}, ErrNotFound
	}
	return d, nil
}

func (f *fakeRepo) DeleteSnapshot(_ context.Context, sceneID, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if d, ok := f.snaps[id]; !ok || d.SceneID != sceneID {
		return ErrNotFound
	}
	delete(f.snaps, id)
	return nil
}
