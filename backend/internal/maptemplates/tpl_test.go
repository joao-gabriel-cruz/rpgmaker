package maptemplates

import (
	"reflect"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// terrainCount counts cells of a terrain.
func terrainCount(g maps.Grid, id string) int {
	n := 0
	for _, c := range g.Cells {
		if c.Terrain == id {
			n++
		}
	}
	return n
}

func prefabCount(objs []maps.Object, id string) int {
	n := 0
	for _, o := range objs {
		if o.Prefab == id {
			n++
		}
	}
	return n
}

// checkTemplate runs the contract tests every template must pass.
func checkTemplate(t *testing.T, id string) {
	t.Helper()
	tpl, ok := Get(id)
	if !ok {
		t.Fatalf("template %s is not registered", id)
	}
	if tpl.Width < 64 || tpl.Depth < 64 || tpl.Width > 128 || tpl.Depth > 128 || tpl.Name == "" || tpl.Description == "" || len(tpl.Tags) == 0 {
		t.Fatalf("bad metadata: %+v", tpl.Info())
	}
	for seed := int64(1); seed <= 12; seed++ {
		g, objs := tpl.Generate(seed)
		if g.Width != tpl.Width || g.Depth != tpl.Depth {
			t.Fatalf("seed %d: grid %dx%d != %dx%d", seed, g.Width, g.Depth, tpl.Width, tpl.Depth)
		}
		if err := Check(g, objs); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if f := Reachable(g, objs); f < 0.9 {
			t.Fatalf("seed %d: only %.0f%% of walkable cells reachable from the entry", seed, f*100)
		}
	}
	checkSeeds(t, tpl)
}

func checkSeeds(t *testing.T, tpl Template) {
	t.Helper()
	g1, o1 := tpl.Generate(7)
	g2, o2 := tpl.Generate(7)
	if !reflect.DeepEqual(g1, g2) || !reflect.DeepEqual(o1, o2) {
		t.Fatal("same seed must give the same map")
	}
	g3, o3 := tpl.Generate(8)
	if reflect.DeepEqual(g1, g3) && reflect.DeepEqual(o1, o3) {
		t.Fatal("different seeds must give different maps")
	}
}
