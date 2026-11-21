package maptemplates

import (
	"reflect"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func newB() *Builder { return NewBuilder(64, 64, 1, "grass", 1) }

func TestNoiseIsSeededAndBounded(t *testing.T) {
	a, b, c := NewNoise(1), NewNoise(1), NewNoise(2)
	same, diff := true, false
	for i := 0; i < 200; i++ {
		x, z := float64(i)*0.37, float64(i)*0.91
		v := a.FBM(x, z, 3)
		if v < 0 || v > 1 {
			t.Fatalf("out of range: %v", v)
		}
		same = same && v == b.FBM(x, z, 3)
		diff = diff || v != c.FBM(x, z, 3)
	}
	if !same || !diff {
		t.Errorf("same seed equal=%v, other seed differs=%v", same, diff)
	}
}

func TestShapes(t *testing.T) {
	b := newB()
	b.FillRect(Rect{2, 2, 4, 3}, Brush{Terrain: "sand", Height: 2})
	b.FillCircle(30, 30, 3, Brush{Terrain: "dirt", Height: Keep})
	b.Line(Pt{0, 10}, Pt{20, 10}, 3, Brush{Terrain: "stone", Height: 1})
	g, _ := b.Build()
	if c, _ := g.At(5, 4); c.Terrain != "sand" || c.Height != 2 {
		t.Errorf("rect: %+v", c)
	}
	if c, _ := g.At(6, 4); c.Terrain != "grass" {
		t.Errorf("rect leaks: %+v", c)
	}
	if c, _ := g.At(30, 33); c.Terrain != "dirt" || c.Height != 1 {
		t.Errorf("circle edge: %+v", c)
	}
	if c, _ := g.At(10, 9); c.Terrain != "stone" {
		t.Errorf("line width: %+v", c)
	}
	if c, _ := g.At(10, 12); c.Terrain != "grass" {
		t.Errorf("line too wide: %+v", c)
	}
}

func TestRoadFlattensAndReserves(t *testing.T) {
	b := newB()
	b.Hill(20, 20, 10, 4, "")
	b.Road([]Pt{{0, 20}, {40, 20}}, 3, "dirt", 1)
	b.Scatter(Rect{0, 0, 64, 64}, Scatter{Prefab: "bush", Density: 1})
	g, objs := b.Build()
	for x := 0; x < 40; x++ {
		if c, _ := g.At(x, 20); c.Terrain != "dirt" || c.Height != 1 {
			t.Fatalf("road cell %d: %+v", x, c)
		}
	}
	for _, o := range objs {
		if o.Z >= 19 && o.Z <= 21 && o.X < 40 {
			t.Fatalf("scatter on the road: %+v", o)
		}
	}
}

func TestRiverHasWaterAndBanks(t *testing.T) {
	b := newB()
	path := b.River(Pt{10, 0}, Pt{14, 63}, 3, 2, "sand")
	g, _ := b.Build()
	if path[0] != (Pt{10, 0}) || path[len(path)-1] != (Pt{14, 63}) {
		t.Errorf("endpoints: %v %v", path[0], path[len(path)-1])
	}
	if terrainCount(g, "water") < 150 || terrainCount(g, "sand") < 150 {
		t.Errorf("water=%d sand=%d", terrainCount(g, "water"), terrainCount(g, "sand"))
	}
	for i := 1; i < len(path); i++ {
		if absInt(path[i].X-path[i-1].X)+absInt(path[i].Z-path[i-1].Z) != 1 {
			t.Fatalf("path is not 4-connected at %d", i)
		}
	}
	if c := b.Cell(path[10].X, path[10].Z); c.Terrain != "water" || c.Height != 0 {
		t.Errorf("center: %+v", c)
	}
}

func TestHillIsSmooth(t *testing.T) {
	b := newB()
	b.Hill(32, 32, 12, 5, "grass")
	g, _ := b.Build()
	if c, _ := g.At(32, 32); c.Height < 5 {
		t.Errorf("peak: %d", c.Height)
	}
	for z := 1; z < 63; z++ {
		for x := 1; x < 63; x++ {
			a, _ := g.At(x, z)
			r, _ := g.At(x+1, z)
			if d := a.Height - r.Height; d > 2 || d < -2 {
				t.Fatalf("cliff at (%d,%d): %d", x, z, d)
			}
		}
	}
}

func TestBuildingHasWallsDoorAndFreeFront(t *testing.T) {
	b := newB()
	b.Scatter(Rect{0, 0, 64, 64}, Scatter{Prefab: "tree", Density: 1})
	info := b.Building(Rect{20, 20, 7, 6}, BuildingOpts{Door: East, Open: true})
	g, objs := b.Build()
	if err := checkObjectsOnly(g, objs); err != nil {
		t.Fatal(err)
	}
	if info.Door != (Pt{26, 23}) || info.Front != (Pt{27, 23}) {
		t.Errorf("door/front: %+v", info)
	}
	var door *maps.Object
	for i, o := range objs {
		if o.Prefab == "door" {
			door = &objs[i]
		}
		if o.X == info.Front.X && o.Z == info.Front.Z {
			t.Errorf("front blocked by %s", o.Prefab)
		}
	}
	if door == nil || door.Props["open"] != true {
		t.Fatalf("door: %+v", door)
	}
	if prefabCount(objs, "wall") != 2*(7+6)-4-1 {
		t.Errorf("walls: %d", prefabCount(objs, "wall"))
	}
	if c, _ := g.At(22, 22); c.Terrain != "floor" {
		t.Errorf("interior: %+v", c)
	}
}

func TestRuinedBuildingMissesWalls(t *testing.T) {
	b := newB()
	b.Building(Rect{20, 20, 8, 8}, BuildingOpts{Ruin: 0.5})
	_, objs := b.Build()
	if n := prefabCount(objs, "wall"); n >= 2*(8+8)-4-1 || n == 0 {
		t.Errorf("walls: %d", n)
	}
}

func TestScatterRespectsTerrainAndDensity(t *testing.T) {
	b := newB()
	b.FillRect(Rect{0, 0, 64, 32}, Brush{Terrain: "water", Height: 0})
	n := b.Scatter(Rect{0, 0, 64, 64}, Scatter{Prefab: "tree", Density: 0.2, Variants: []string{"pine", "round"}})
	_, objs := b.Build()
	if n != len(objs) || n < 200 || n > 600 {
		t.Fatalf("placed %d (objs %d)", n, len(objs))
	}
	for _, o := range objs {
		if o.Z < 32 || o.Props["variant"] == nil {
			t.Fatalf("bad scatter: %+v", o)
		}
	}
}

func TestFenceGapsAndClear(t *testing.T) {
	b := newB()
	b.Fence(Pt{10, 10}, Pt{19, 10}, 4, 5)
	_, objs := b.Build()
	if len(objs) != 8 {
		t.Fatalf("fence: %d", len(objs))
	}
	b.FenceRect(Rect{30, 30, 10, 8}, South)
	b.Clear(Rect{30, 30, 10, 8})
	_, objs = b.Build()
	if len(objs) != 8 {
		t.Errorf("clear left %d objects", len(objs))
	}
	if !b.Place("crate", 30, 30, 0, nil) {
		t.Error("cleared cell must be placeable again")
	}
}

func TestPlaceRejectsOverlapAndEdges(t *testing.T) {
	b := newB()
	if !b.Place("car", 10, 10, 0, nil) || b.Place("crate", 11, 10, 0, nil) {
		t.Error("car footprint is 2 cells")
	}
	if b.Place("car", 63, 5, 0, nil) || b.Place("crate", -1, 0, 0, nil) {
		t.Error("must stay inside the grid")
	}
}

func TestClearEntryFreesTheArea(t *testing.T) {
	b := newB()
	r := EntryRect(64, 64)
	b.FillRect(r, Brush{Terrain: "stone", Height: 1})
	b.Place("tree", r.X+1, r.Z+1, 0, nil)
	b.ClearEntry()
	g, objs := b.Build()
	if err := Check(g, objs); err != nil {
		t.Fatal(err)
	}
}

func checkObjectsOnly(g maps.Grid, objs []maps.Object) error {
	_, err := checkObjects(g, objs)
	return err
}

func TestCheckRejectsBadMaps(t *testing.T) {
	b := newB()
	g, _ := b.Build()
	bad := map[string][]maps.Object{
		"unknown prefab": {{ID: "a", Prefab: "ufo", X: 1, Z: 1}},
		"outside":        {{ID: "a", Prefab: "car", X: 63, Z: 1}},
		"overlap":        {{ID: "a", Prefab: "car", X: 5, Z: 5}, {ID: "b", Prefab: "tree", X: 6, Z: 5}},
		"duplicate id":   {{ID: "a", Prefab: "tree", X: 5, Z: 5}, {ID: "a", Prefab: "tree", X: 8, Z: 5}},
		"rotation":       {{ID: "a", Prefab: "tree", X: 5, Z: 5, Rotation: 45}},
		"entry blocked":  {{ID: "a", Prefab: "tree", X: 31, Z: 59}},
	}
	for name, objs := range bad {
		if Check(g, objs) == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	small, _ := maps.NewGrid(32, 32)
	if Check(*small, nil) == nil {
		t.Error("32x32 must be rejected")
	}
	if Check(g, nil) != nil {
		t.Error("an empty 64x64 map is valid")
	}
}

func TestBuilderIsDeterministic(t *testing.T) {
	run := func() (maps.Grid, []maps.Object) {
		b := NewBuilder(64, 64, 9, "grass", 1)
		b.River(Pt{5, 0}, Pt{9, 63}, 3, 3, "sand")
		b.Scatter(Rect{0, 0, 64, 64}, Scatter{Prefab: "tree", Density: 0.1, Clump: 6, Rotate: true})
		return b.Build()
	}
	g1, o1 := run()
	g2, o2 := run()
	if !reflect.DeepEqual(g1, g2) || !reflect.DeepEqual(o1, o2) {
		t.Fatal("builder output must be deterministic")
	}
}

func TestRegistry(t *testing.T) {
	all := All()
	if len(all) < 2 || all[0].ID != "vila-abandonada" || all[1].ID != "rodovia-engarrafada" {
		t.Fatalf("gallery order: %v", len(all))
	}
	if _, ok := Get("nope"); ok {
		t.Error("unknown id")
	}
	defer func() {
		if recover() == nil {
			t.Error("duplicate register must panic")
		}
	}()
	Register(all[0])
}
