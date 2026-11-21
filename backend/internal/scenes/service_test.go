package scenes

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

var ctx = context.Background()

func TestCreateValidatesNameAndSize(t *testing.T) {
	svc, _, _ := newSvc()
	if _, err := svc.Create(ctx, camp, " ", 32, 32); !errors.Is(err, ErrInvalidName) {
		t.Errorf("empty name: %v", err)
	}
	for _, sz := range [][2]int{{15, 32}, {32, 129}, {0, 0}} {
		if _, err := svc.Create(ctx, camp, "x", sz[0], sz[1]); !errors.Is(err, ErrInvalidSize) {
			t.Errorf("size %v: %v", sz, err)
		}
	}
	sc, err := svc.Create(ctx, camp, "  Floresta ", 128, 16)
	if err != nil || sc.Name != "Floresta" || len(sc.Grid.Data) != 128*16*2 {
		t.Fatalf("create: %+v %v", sc, err)
	}
}

func TestActivateKeepsOnlyOneActive(t *testing.T) {
	svc, _, n := newSvc()
	a, _ := svc.Create(ctx, camp, "A", 16, 16)
	b, _ := svc.Create(ctx, camp, "B", 16, 16)
	for _, id := range []string{a.ID, b.ID} {
		if err := svc.Activate(ctx, camp, id); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := svc.List(ctx, camp)
	if list[0].IsActive || !list[1].IsActive {
		t.Errorf("list: %+v", list)
	}
	if ev, ok := n.last(EventActivated); !ok || !ev.toPlyr {
		t.Errorf("activated event: %+v", ev)
	}
	if err := svc.Activate(ctx, camp, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("bad id: %v", err)
	}
}

func TestPlayersOnlyLoadActiveScene(t *testing.T) {
	svc, _, _ := newSvc()
	a, _ := svc.Create(ctx, camp, "A", 16, 16)
	if _, err := svc.Get(ctx, camp, a.ID, "player"); !errors.Is(err, ErrNotFound) {
		t.Errorf("inactive for player: %v", err)
	}
	if _, err := svc.Get(ctx, camp, a.ID, "gm"); err != nil {
		t.Errorf("inactive for gm: %v", err)
	}
	_ = svc.Activate(ctx, camp, a.ID)
	if _, err := svc.Get(ctx, camp, a.ID, "player"); err != nil {
		t.Errorf("active for player: %v", err)
	}
	if _, err := svc.Get(ctx, "00000000-0000-0000-0000-0000000000ff", a.ID, "gm"); !errors.Is(err, ErrNotFound) {
		t.Errorf("other campaign: %v", err)
	}
}

func TestPaintPersistsAndRejectsInvalid(t *testing.T) {
	svc, _, n := newSvc()
	sc, _ := svc.Create(ctx, camp, "A", 32, 32)
	good := maps.Patch{Cells: []maps.CellPatch{{X: 3, Z: 4, Terrain: "stone", Height: 5}}}
	if err := svc.Paint(ctx, camp, sc.ID, "", good); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.Get(ctx, camp, sc.ID, "gm")
	g, _ := maps.Deserialize(got.Grid)
	if c, _ := g.At(3, 4); c.Terrain != "stone" || c.Height != 5 {
		t.Errorf("cell: %+v", c)
	}
	if ev, _ := n.last(EventPatched); !ev.toGM || ev.toPlyr {
		t.Errorf("inactive scene patch must go to gm only: %+v", ev)
	}
	for _, bad := range []maps.CellPatch{{X: 99, Z: 0, Terrain: "floor"}, {X: 0, Z: 0, Terrain: "lava2"}, {X: 0, Z: 0, Terrain: "floor", Height: 11}} {
		if err := svc.Paint(ctx, camp, sc.ID, "", maps.Patch{Cells: []maps.CellPatch{bad}}); !errors.Is(err, ErrInvalidPatch) {
			t.Errorf("bad %+v: %v", bad, err)
		}
	}
	if err := svc.Paint(ctx, camp, sc.ID, "", maps.Patch{}); !errors.Is(err, ErrInvalidPatch) {
		t.Errorf("empty: %v", err)
	}
	_ = svc.Activate(ctx, camp, sc.ID)
	_ = svc.Paint(ctx, camp, sc.ID, "", good)
	if ev, _ := n.last(EventPatched); !ev.toPlyr {
		t.Errorf("active scene patch must reach players")
	}
}

func TestConcurrentPaintsAreNotLost(t *testing.T) {
	svc, _, _ := newSvc()
	sc, _ := svc.Create(ctx, camp, "A", 32, 32)
	var wg sync.WaitGroup
	for x := 0; x < 20; x++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = svc.Paint(ctx, camp, sc.ID, "", maps.Patch{Cells: []maps.CellPatch{{X: x, Z: 0, Terrain: "lava", Height: 2}}})
		}()
	}
	wg.Wait()
	got, _ := svc.Get(ctx, camp, sc.ID, "gm")
	g, _ := maps.Deserialize(got.Grid)
	for x := 0; x < 20; x++ {
		if c, _ := g.At(x, 0); c.Terrain != "lava" {
			t.Fatalf("lost patch at %d", x)
		}
	}
}

func obj(id, prefab string, x, z, rot int) maps.Object {
	return maps.Object{ID: id, Prefab: prefab, X: x, Z: z, Rotation: rot}
}

func TestObjectPatches(t *testing.T) {
	svc, _, n := newSvc()
	sc, _ := svc.Create(ctx, camp, "A", 16, 16)
	add := ObjectPatch{Add: []maps.Object{obj("w1", "wall", 2, 2, 0), obj("c1", "car", 4, 4, 90)}}
	if err := svc.EditObjects(ctx, camp, sc.ID, "", add); err != nil {
		t.Fatal(err)
	}
	if _, ok := n.last(EventObjectsPatched); !ok {
		t.Error("no event")
	}
	upd := ObjectPatch{Update: []maps.Object{obj("w1", "wall", 5, 5, 0)}, Remove: []string{"c1"}}
	if err := svc.EditObjects(ctx, camp, sc.ID, "", upd); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.Get(ctx, camp, sc.ID, "gm")
	if len(got.Objects) != 1 || got.Objects[0].X != 5 {
		t.Errorf("objects: %+v", got.Objects)
	}
}

func TestObjectPatchRejections(t *testing.T) {
	svc, _, _ := newSvc()
	sc, _ := svc.Create(ctx, camp, "A", 16, 16)
	_ = svc.EditObjects(ctx, camp, sc.ID, "", ObjectPatch{Add: []maps.Object{obj("w1", "wall", 2, 2, 0)}})
	bad := map[string]ObjectPatch{
		"unknown prefab": {Add: []maps.Object{obj("x", "dragon", 1, 1, 0)}},
		"out of bounds":  {Add: []maps.Object{obj("x", "volcano", 14, 14, 0)}},
		"rotated oob":    {Add: []maps.Object{obj("x", "car", 15, 15, 90)}},
		"bad rotation":   {Add: []maps.Object{obj("x", "wall", 1, 1, 45)}},
		"duplicate id":   {Add: []maps.Object{obj("w1", "wall", 3, 3, 0)}},
		"update missing": {Update: []maps.Object{obj("zz", "wall", 3, 3, 0)}},
		"empty id":       {Add: []maps.Object{obj("", "wall", 3, 3, 0)}},
	}
	for name, p := range bad {
		if err := svc.EditObjects(ctx, camp, sc.ID, "", p); !errors.Is(err, ErrInvalidPatch) {
			t.Errorf("%s: %v", name, err)
		}
	}
	got, _ := svc.Get(ctx, camp, sc.ID, "gm")
	if len(got.Objects) != 1 {
		t.Errorf("rejected patches must not persist: %+v", got.Objects)
	}
}

func TestRenameAndDelete(t *testing.T) {
	svc, _, _ := newSvc()
	sc, _ := svc.Create(ctx, camp, "A", 16, 16)
	if err := svc.Rename(ctx, camp, sc.ID, ""); !errors.Is(err, ErrInvalidName) {
		t.Errorf("rename empty: %v", err)
	}
	if err := svc.Rename(ctx, camp, sc.ID, "B"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete(ctx, camp, sc.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete(ctx, camp, sc.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("delete twice: %v", err)
	}
}

func TestPatchSkipsOrigin(t *testing.T) {
	svc, _, n := newSvc()
	sc, _ := svc.Create(ctx, camp, "A", 16, 16)
	_ = svc.Activate(ctx, camp, sc.ID)
	_ = svc.Paint(ctx, camp, sc.ID, "u1", maps.Patch{Cells: []maps.CellPatch{{X: 1, Z: 1, Terrain: "sand", Height: 1}}})
	ev, _ := n.last(EventPatched)
	if ev.toU1 || !ev.toPlyr || !ev.toGM {
		t.Errorf("origin must be skipped and others served: %+v", ev)
	}
}
