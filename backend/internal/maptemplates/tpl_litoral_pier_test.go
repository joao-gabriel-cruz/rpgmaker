package maptemplates

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func TestLitoralPier(t *testing.T) {
	checkTemplate(t, "litoral-pier")
	tpl, _ := Get("litoral-pier")
	if tpl.Width != 64 || tpl.Depth != 96 {
		t.Fatalf("size %dx%d", tpl.Width, tpl.Depth)
	}
	for seed := int64(1); seed <= 6; seed++ {
		g, objs := tpl.Generate(seed)
		if n := pierPlanks(g, objs); n < 8 {
			t.Fatalf("seed %d: only %d bridge planks over water", seed, n)
		}
		if water := terrainCount(g, "water"); water < 64*96/6 || terrainCount(g, "sand") < 400 {
			t.Fatalf("seed %d: water=%d sand=%d", seed, water, terrainCount(g, "sand"))
		}
		if prefabCount(objs, "door") < 3 || prefabCount(objs, "watchtower") != 1 || prefabCount(objs, "tree") < 15 {
			t.Fatalf("seed %d: shacks=%d tower=%d palms=%d", seed, prefabCount(objs, "door"), prefabCount(objs, "watchtower"), prefabCount(objs, "tree"))
		}
	}
}

func pierPlanks(g maps.Grid, objs []maps.Object) int {
	n := 0
	for _, o := range objs {
		if o.Prefab == "bridge" && g.Cells[o.Z*g.Width+o.X].Terrain == "water" {
			n++
		}
	}
	return n
}
