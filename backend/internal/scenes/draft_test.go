package scenes

import (
	"errors"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

var volcano = ObjectPatch{Add: []maps.Object{{ID: "v", Prefab: "wall", X: 2, Z: 2}}}

func activeScene(t *testing.T, svc *Service) Scene {
	t.Helper()
	sc, _ := svc.Create(ctx, camp, "A", 16, 16)
	if err := svc.Activate(ctx, camp, sc.ID); err != nil {
		t.Fatal(err)
	}
	return sc
}

func TestDraftEditsStayHiddenFromPlayers(t *testing.T) {
	svc, _, n := newSvc()
	sc := activeScene(t, svc)
	if err := svc.StartDraft(ctx, camp, sc.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.StartDraft(ctx, camp, sc.ID); !errors.Is(err, ErrDraftState) {
		t.Errorf("double start: %v", err)
	}
	n.events = nil
	if err := svc.EditObjects(ctx, camp, sc.ID, "u1", volcano); err != nil {
		t.Fatal(err)
	}
	if err := svc.Paint(ctx, camp, sc.ID, "u1", oneCell(5)); err != nil {
		t.Fatal(err)
	}
	for _, e := range n.events {
		if e.toPlyr || e.toU1 || !e.toGM {
			t.Errorf("draft event leaked or misrouted: %+v", e)
		}
	}
	if _, ok := n.last(EventPatched); ok {
		t.Error("real event emitted while drafting")
	}
	checkViews(t, svc, sc)
}

func checkViews(t *testing.T, svc *Service, sc Scene) {
	t.Helper()
	pl, _ := svc.Get(ctx, camp, sc.ID, "player")
	if len(pl.Objects) != 0 || pl.Draft != nil || pl.Grid.Data != sc.Grid.Data {
		t.Errorf("player sees draft: %+v", pl)
	}
	gm, _ := svc.Get(ctx, camp, sc.ID, "gm")
	if gm.Draft == nil || len(gm.Draft.Objects) != 1 || len(gm.Objects) != 0 {
		t.Errorf("gm view: %+v", gm)
	}
}

func TestDraftApplyReplacesForEveryone(t *testing.T) {
	svc, _, n := newSvc()
	sc := activeScene(t, svc)
	_ = svc.StartDraft(ctx, camp, sc.ID)
	_ = svc.EditObjects(ctx, camp, sc.ID, "u1", volcano)
	if err := svc.ApplyDraft(ctx, camp, sc.ID); err != nil {
		t.Fatal(err)
	}
	ev, ok := n.last(EventReplaced)
	if !ok || !ev.toPlyr || !ev.toGM || !ev.toU1 {
		t.Fatalf("replaced event: %+v", ev)
	}
	if got := ev.payload.(map[string]any)["objects"].([]maps.Object); len(got) != 1 {
		t.Errorf("payload: %+v", got)
	}
	pl, _ := svc.Get(ctx, camp, sc.ID, "player")
	if len(pl.Objects) != 1 {
		t.Error("player does not see applied volcano")
	}
	gm, _ := svc.Get(ctx, camp, sc.ID, "gm")
	if gm.Draft != nil {
		t.Error("draft survived apply")
	}
	if err := svc.ApplyDraft(ctx, camp, sc.ID); !errors.Is(err, ErrDraftState) {
		t.Errorf("apply without draft: %v", err)
	}
}

func TestDraftDiscardRestoresLiveEditing(t *testing.T) {
	svc, _, n := newSvc()
	sc := activeScene(t, svc)
	_ = svc.StartDraft(ctx, camp, sc.ID)
	_ = svc.EditObjects(ctx, camp, sc.ID, "u1", volcano)
	if err := svc.DiscardDraft(ctx, camp, sc.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.DiscardDraft(ctx, camp, sc.ID); !errors.Is(err, ErrDraftState) {
		t.Errorf("discard twice: %v", err)
	}
	gm, _ := svc.Get(ctx, camp, sc.ID, "gm")
	if gm.Draft != nil || len(gm.Objects) != 0 {
		t.Errorf("after discard: %+v", gm)
	}
	_ = svc.EditObjects(ctx, camp, sc.ID, "u1", volcano)
	if ev, ok := n.last(EventObjectsPatched); !ok || !ev.toPlyr {
		t.Error("live edit not broadcast after discard")
	}
}

func TestSnapshotsSaveApplyDelete(t *testing.T) {
	svc, _, n := newSvc()
	sc := activeScene(t, svc)
	_ = svc.EditObjects(ctx, camp, sc.ID, "u1", volcano)
	sn, err := svc.SaveSnapshot(ctx, camp, sc.ID, " erupção ")
	if err != nil || sn.Name != "erupção" {
		t.Fatalf("save: %+v %v", sn, err)
	}
	if _, err := svc.SaveSnapshot(ctx, camp, sc.ID, " "); !errors.Is(err, ErrInvalidName) {
		t.Errorf("empty name: %v", err)
	}
	_ = svc.EditObjects(ctx, camp, sc.ID, "u1", ObjectPatch{Remove: []string{"v"}})
	if err := svc.ApplySnapshot(ctx, camp, sc.ID, sn.ID); err != nil {
		t.Fatal(err)
	}
	ev, ok := n.last(EventReplaced)
	if !ok || !ev.toPlyr || len(ev.payload.(map[string]any)["objects"].([]maps.Object)) != 1 {
		t.Errorf("replaced: %+v", ev)
	}
	if err := svc.ApplySnapshot(ctx, camp, sc.ID, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("bad snapshot: %v", err)
	}
	if err := svc.DeleteSnapshot(ctx, camp, sc.ID, sn.ID); err != nil {
		t.Fatal(err)
	}
	if l, _ := svc.ListSnapshots(ctx, camp, sc.ID); len(l) != 0 {
		t.Errorf("list after delete: %+v", l)
	}
}

func TestDraftAndSnapshotCommandsAreGMOnly(t *testing.T) {
	svc, _, _ := newSvc()
	reg := fakeRegistrar{}
	RegisterCommands(reg, svc)
	sc := activeScene(t, svc)
	p := map[string]any{"sceneId": sc.ID, "snapshotId": sc.ID}
	for _, typ := range []string{"map.draft.start", "map.draft.apply", "map.draft.discard", "map.snapshot.apply"} {
		if err := reg.dispatch("player", typ, p); err == nil || err.Error() != "Você não tem permissão para fazer isso" {
			t.Errorf("%s as player: %v", typ, err)
		}
	}
	if err := reg.dispatch("gm", "map.draft.start", p); err != nil {
		t.Errorf("gm start: %v", err)
	}
	if err := reg.dispatch("gm", "map.draft.start", p); err == nil {
		t.Error("double start accepted")
	}
	if err := reg.dispatch("gm", "map.draft.apply", p); err != nil {
		t.Errorf("gm apply: %v", err)
	}
}

func oneCell(h int) maps.Patch {
	return maps.Patch{Cells: []maps.CellPatch{{X: 3, Z: 4, Terrain: "stone", Height: h}}}
}
