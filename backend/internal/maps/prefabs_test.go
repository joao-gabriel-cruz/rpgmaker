package maps

import "testing"

func TestNewPrefabRules(t *testing.T) {
	cases := []struct {
		id         string
		w, d       int
		move, view bool
	}{
		{"tent", 2, 2, true, false}, {"barrel", 1, 1, false, false}, {"bench", 2, 1, true, false},
		{"market_stall", 2, 2, false, false}, {"lamp_post", 1, 1, false, false}, {"sandbags", 1, 1, true, false},
		{"ruined_wall", 1, 1, true, true}, {"well", 1, 1, true, false}, {"sign", 1, 1, false, false},
		{"bridge", 1, 1, false, false}, {"bunting", 2, 1, false, false},
	}
	for _, c := range cases {
		if !KnownPrefab(c.id) {
			t.Fatalf("%s unknown", c.id)
		}
		r := RuleFor(Object{Prefab: c.id})
		if r.W != c.w || r.D != c.d || r.BlocksMovement != c.move || r.BlocksVision != c.view {
			t.Errorf("%s: got %+v", c.id, r)
		}
	}
	if n := len(Footprint(Object{Prefab: "tent", X: 1, Z: 1, Rotation: 90})); n != 4 {
		t.Errorf("tent footprint %d", n)
	}
	if f := Footprint(Object{Prefab: "bench", X: 0, Z: 0, Rotation: 90}); f[1] != (Coord{0, 1}) {
		t.Errorf("rotated bench %v", f)
	}
}

func TestRuinedWallHeightProp(t *testing.T) {
	for in, want := range map[any]int{nil: 1, float64(1): 1, float64(2): 2, 2: 2, float64(9): 2} {
		o := Object{Prefab: "ruined_wall"}
		if in != nil {
			o.Props = map[string]any{"height": in}
		}
		if got := RuleFor(o).Height; got != want {
			t.Errorf("height %v: got %d want %d", in, got, want)
		}
	}
}
