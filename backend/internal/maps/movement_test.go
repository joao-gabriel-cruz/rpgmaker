package maps

import "testing"

func flat(t *testing.T, w, d int) *Grid {
	t.Helper()
	g, err := NewGrid(w, d)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func set(g *Grid, x, z int, terrain string, h int) {
	g.Cells[z*g.Width+x] = Cell{terrain, h}
}

type enterCase struct {
	name     string
	from, to Coord
	step     int
	want     bool
}

var enterCases = []enterCase{
	{"free", Coord{0, 0}, Coord{0, 2}, 0, true},
	{"terrain blocks", Coord{0, 0}, Coord{1, 0}, 0, false},
	{"object blocks", Coord{0, 2}, Coord{0, 1}, 0, false},
	{"closed door", Coord{1, 2}, Coord{1, 1}, 0, false},
	{"open door", Coord{2, 2}, Coord{2, 1}, 0, true},
	{"rotated car covers z+1", Coord{4, 2}, Coord{3, 2}, 0, false},
	{"rotated car not wide", Coord{3, 2}, Coord{4, 2}, 0, true},
	{"high step", Coord{0, 0}, Coord{2, 0}, 0, false},
	{"step allowed by param", Coord{0, 0}, Coord{2, 0}, 3, true},
	{"step 2 ok", Coord{2, 0}, Coord{3, 0}, 0, true},
	{"out of grid", Coord{0, 0}, Coord{-1, 0}, 0, false},
	{"unknown prefab", Coord{4, 1}, Coord{5, 1}, 0, true},
}

func TestCanEnter(t *testing.T) {
	g := flat(t, 6, 3)
	set(g, 1, 0, "stone", 1)
	set(g, 2, 0, "grass", 4)
	set(g, 3, 0, "grass", 3)
	objs := []Object{
		{ID: "a", Prefab: "wall", X: 0, Z: 1},
		{ID: "d1", Prefab: "door", X: 1, Z: 1},
		{ID: "d2", Prefab: "door", X: 2, Z: 1, Props: map[string]any{"open": true}},
		{ID: "c", Prefab: "car", X: 3, Z: 1, Rotation: 90},
		{ID: "u", Prefab: "mystery", X: 5, Z: 1},
	}
	for _, c := range enterCases {
		t.Run(c.name, func(t *testing.T) {
			ok, reason := CanEnter(g, objs, c.from, c.to, c.step)
			if ok != c.want || (!ok && reason == "") {
				t.Fatalf("got %v %q, want %v", ok, reason, c.want)
			}
		})
	}
}

func TestPathBlocked(t *testing.T) {
	g := flat(t, 7, 3)
	wall := []Object{{Prefab: "wall", X: 3, Z: 1}}
	if PathBlocked(g, nil, Coord{0, 1}, Coord{6, 1}) {
		t.Error("open path must not be blocked")
	}
	if !PathBlocked(g, wall, Coord{0, 1}, Coord{6, 1}) {
		t.Error("wall must block the path")
	}
	if PathBlocked(g, wall, Coord{0, 0}, Coord{6, 0}) {
		t.Error("path beside the wall is free")
	}
}

func bridgeGrid(t *testing.T) *Grid {
	g := flat(t, 5, 3)
	for z := 0; z < 3; z++ {
		set(g, 0, z, "grass", 4)
		set(g, 4, z, "grass", 4)
		for x := 1; x <= 3; x++ {
			set(g, x, z, "water", 0)
		}
	}
	set(g, 2, 2, "lava", 0)
	return g
}

func TestBridgeCrossesGap(t *testing.T) {
	g := bridgeGrid(t)
	bridge := []Object{
		{ID: "b1", Prefab: "bridge", X: 1, Z: 1},
		{ID: "b2", Prefab: "bridge", X: 2, Z: 1},
		{ID: "b3", Prefab: "bridge", X: 3, Z: 1},
	}
	if ok, _ := CanEnter(g, nil, Coord{0, 1}, Coord{1, 1}, 0); ok {
		t.Fatal("drop of 4 must block without a bridge")
	}
	if PathBlocked(g, bridge, Coord{0, 1}, Coord{4, 1}) {
		t.Fatal("bridge must allow crossing the gap")
	}
	if !PathBlocked(g, bridge, Coord{0, 0}, Coord{4, 0}) {
		t.Fatal("cells beside the bridge keep blocking")
	}
	if ok, _ := CanEnter(g, bridge, Coord{1, 1}, Coord{1, 0}, 0); ok {
		t.Fatal("stepping off the bridge into the gap stays blocked")
	}
}

func TestBridgeOverBlockingTerrainAndRules(t *testing.T) {
	g := flat(t, 3, 1)
	set(g, 1, 0, "stone", 1)
	br := []Object{{ID: "b", Prefab: "bridge", X: 1, Z: 0}}
	if ok, _ := CanEnter(g, nil, Coord{0, 0}, Coord{1, 0}, 0); ok {
		t.Fatal("stone blocks without bridge")
	}
	if ok, _ := CanEnter(g, br, Coord{0, 0}, Coord{1, 0}, 0); !ok {
		t.Fatal("bridge makes the cell walkable")
	}
	if r := RuleFor(br[0]); r.BlocksMovement || r.BlocksVision {
		t.Fatal("bridge never blocks")
	}
}

func TestBridgeDeckProp(t *testing.T) {
	g := bridgeGrid(t)
	b := []Object{{ID: "b", Prefab: "bridge", X: 1, Z: 1, Props: map[string]any{"deck": float64(1)}}}
	if ok, _ := CanEnter(g, b, Coord{0, 1}, Coord{1, 1}, 0); ok {
		t.Fatal("deck 1 is 3 below the bank: too steep")
	}
	b[0].Props["deck"] = float64(3)
	if ok, _ := CanEnter(g, b, Coord{0, 1}, Coord{1, 1}, 0); !ok {
		t.Fatal("deck 3 is one below the bank")
	}
}
