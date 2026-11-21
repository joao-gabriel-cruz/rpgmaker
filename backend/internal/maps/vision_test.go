package maps

import "testing"

func has(s Set, x, z int) bool { _, ok := s[Coord{x, z}]; return ok }

func TestVisibleOpenField(t *testing.T) {
	g := flat(t, 21, 21)
	s := Visible(g, nil, Coord{10, 10}, 1.5, 5)
	for z := 0; z < 21; z++ {
		for x := 0; x < 21; x++ {
			dx, dz := x-10, z-10
			if want := dx*dx+dz*dz <= 25; has(s, x, z) != want {
				t.Fatalf("(%d,%d) visible=%v want %v", x, z, !want, want)
			}
		}
	}
}

func TestWallHidesBehind(t *testing.T) {
	g := flat(t, 11, 5)
	s := Visible(g, []Object{{Prefab: "wall", X: 4, Z: 2}}, Coord{2, 2}, 1.5, 8)
	if !has(s, 4, 2) || has(s, 7, 2) {
		t.Fatal("wall cell visible, cell behind hidden")
	}
	if !has(s, 7, 0) {
		t.Fatal("cell off to the side stays visible")
	}
}

func TestLowWallSeenOverFromPeak(t *testing.T) {
	g := flat(t, 11, 5)
	set(g, 1, 2, "grass", 8)
	low := []Object{{Prefab: "crate", X: 3, Z: 2}}
	if !has(Visible(g, low, Coord{1, 2}, 1.5, 8), 6, 2) {
		t.Fatal("peak viewer sees over a low crate")
	}
	if has(Visible(g, low, Coord{2, 2}, 0.2, 8), 6, 2) {
		t.Fatal("low viewer must be blocked by the crate")
	}
}

func TestHillBlocksValley(t *testing.T) {
	g := flat(t, 11, 3)
	set(g, 0, 1, "grass", 0)
	set(g, 4, 1, "grass", 6)
	set(g, 8, 1, "grass", 0)
	s := Visible(g, nil, Coord{0, 1}, 1.5, 9)
	if !has(s, 4, 1) {
		t.Fatal("hill itself is visible")
	}
	if has(s, 8, 1) {
		t.Fatal("valley behind the hill must be hidden")
	}
}

func TestVolcanoBlocksMovementAndVisionUnlessViewerHigher(t *testing.T) {
	g := flat(t, 12, 7)
	volcano := []Object{{Prefab: "volcano", X: 4, Z: 2}}
	if ok, _ := CanEnter(g, volcano, Coord{3, 3}, Coord{5, 3}, 0); ok {
		t.Fatal("volcano blocks movement")
	}
	if has(Visible(g, volcano, Coord{2, 3}, 1.5, 9), 9, 3) {
		t.Fatal("volcano (top 4) blocks eye at 2.5")
	}
	set(g, 2, 3, "grass", 8)
	if !has(Visible(g, volcano, Coord{2, 3}, 1.5, 9), 9, 3) {
		t.Fatal("viewer above the volcano sees past it")
	}
}

func TestStoneTerrainBlocksVision(t *testing.T) {
	g := flat(t, 9, 3)
	set(g, 4, 1, "stone", 1)
	if has(Visible(g, nil, Coord{1, 1}, 1.5, 7), 7, 1) {
		t.Fatal("stone hides what is behind")
	}
}

func TestVisibleForManyUnion(t *testing.T) {
	g := flat(t, 30, 5)
	s := VisibleForMany(g, nil, []Coord{{2, 2}, {27, 2}}, 1.5, 3)
	if !has(s, 4, 2) || !has(s, 25, 2) || has(s, 15, 2) {
		t.Fatal("union of both viewers' areas expected")
	}
}

func BenchmarkVisible64Radius20(b *testing.B) {
	g, _ := NewGrid(64, 64)
	objs := make([]Object, 0, 200)
	for i := 0; i < 200; i++ {
		objs = append(objs, Object{Prefab: "tree", X: (i * 7) % 64, Z: (i * 13) % 64})
	}
	for b.Loop() {
		Visible(g, objs, Coord{32, 32}, 1.5, 20)
	}
}

// BenchmarkVisible128 measures one viewer (radius 20) on the largest scene with scattered blockers.
func BenchmarkVisible128(b *testing.B) {
	g, _ := NewGrid(128, 128)
	var objs []Object
	for i := 0; i < 1500; i++ {
		objs = append(objs, Object{ID: "o", Prefab: "tree", X: (i * 37) % 128, Z: (i * 61) % 128})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Visible(g, objs, Coord{64, 64}, 1.5, 20)
	}
}
