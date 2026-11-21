package scenes

import (
	"context"
	"fmt"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

const camp = "00000000-0000-0000-0000-00000000000c"

type fakeRepo struct {
	mu     sync.Mutex
	scenes map[string]*Scene
	order  []string
	snaps  map[string]SnapshotData
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{scenes: map[string]*Scene{}, snaps: map[string]SnapshotData{}}
}

func (f *fakeRepo) Create(_ context.Context, campaignID, name string, w, d int, grid string) (Scene, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := fmt.Sprintf("00000000-0000-0000-0000-%012d", len(f.order)+1)
	sc := &Scene{ID: id, CampaignID: campaignID, Name: name, Width: w, Depth: d,
		Grid: maps.Serialized{Width: w, Depth: d, Data: grid}, Objects: []maps.Object{}}
	f.scenes[id] = sc
	f.order = append(f.order, id)
	return *sc, nil
}

func (f *fakeRepo) find(campaignID, id string) (*Scene, error) {
	if sc, ok := f.scenes[id]; ok && sc.CampaignID == campaignID {
		return sc, nil
	}
	return nil, ErrNotFound
}

func (f *fakeRepo) List(_ context.Context, campaignID string) ([]Summary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Summary{}
	for _, id := range f.order {
		if sc := f.scenes[id]; sc != nil && sc.CampaignID == campaignID {
			out = append(out, Summary{ID: sc.ID, Name: sc.Name, Width: sc.Width, Depth: sc.Depth, IsActive: sc.IsActive})
		}
	}
	return out, nil
}

func (f *fakeRepo) Get(_ context.Context, campaignID, id string) (Scene, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	sc, err := f.find(campaignID, id)
	if err != nil {
		return Scene{}, err
	}
	cp := *sc
	cp.Objects = append([]maps.Object{}, sc.Objects...)
	if sc.Draft != nil {
		d := *sc.Draft
		d.Objects = append([]maps.Object{}, d.Objects...)
		cp.Draft = &d
	}
	return cp, nil
}

func (f *fakeRepo) Rename(_ context.Context, campaignID, id, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	sc, err := f.find(campaignID, id)
	if err == nil {
		sc.Name = name
	}
	return err
}

func (f *fakeRepo) Delete(_ context.Context, campaignID, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, err := f.find(campaignID, id); err != nil {
		return err
	}
	delete(f.scenes, id)
	return nil
}

func (f *fakeRepo) Activate(_ context.Context, campaignID, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, err := f.find(campaignID, id); err != nil {
		return err
	}
	for _, sc := range f.scenes {
		if sc.CampaignID == campaignID {
			sc.IsActive = sc.ID == id
		}
	}
	return nil
}

func (f *fakeRepo) SaveGrid(_ context.Context, id, grid string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.scenes[id].Grid.Data = grid
	return nil
}

func (f *fakeRepo) SaveObjects(_ context.Context, id string, objects []maps.Object) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.scenes[id].Objects = objects
	return nil
}

type event struct {
	typ     string
	toGM    bool
	toPlyr  bool
	toU1    bool
	payload any
}

type fakeNotifier struct {
	mu     sync.Mutex
	events []event
}

func (n *fakeNotifier) Broadcast(_ string, to realtime.Filter, typ string, payload any) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.events = append(n.events, event{typ, to(realtime.Sender{Role: "gm"}), to(realtime.Sender{Role: "player"}), to(realtime.Sender{UserID: "u1", ConnID: "u1", Role: "gm"}), payload})
}

func (n *fakeNotifier) last(typ string) (event, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for i := len(n.events) - 1; i >= 0; i-- {
		if n.events[i].typ == typ {
			return n.events[i], true
		}
	}
	return event{}, false
}

func newSvc() (*Service, *fakeRepo, *fakeNotifier) {
	r, n := newFakeRepo(), &fakeNotifier{}
	return NewService(r, n), r, n
}
