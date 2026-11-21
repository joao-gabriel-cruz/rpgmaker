package maptemplates

import "testing"

func TestPantano(t *testing.T) {
	checkTemplate(t, "pantano")
	tpl, _ := Get("pantano")
	for seed := int64(1); seed <= 6; seed++ {
		g, objs := tpl.Generate(seed)
		if prefabCount(objs, "bridge") < 10 {
			t.Fatalf("seed %d: only %d bridge planks", seed, prefabCount(objs, "bridge"))
		}
		if frac := float64(terrainCount(g, "water")) / float64(len(g.Cells)); frac < 0.25 {
			t.Fatalf("seed %d: only %.0f%% water", seed, frac*100)
		}
		if terrainCount(g, "dirt") < 100 || prefabCount(objs, "bush") < 60 || prefabCount(objs, "door") != 1 || prefabCount(objs, "lamp_post") < 2 {
			t.Fatalf("seed %d: mud=%d bushes=%d doors=%d lamps=%d", seed, terrainCount(g, "dirt"), prefabCount(objs, "bush"), prefabCount(objs, "door"), prefabCount(objs, "lamp_post"))
		}
		if !ReachSet(g, objs)[Pt{32, 20}] {
			t.Fatalf("seed %d: the hut islet must be reachable", seed)
		}
	}
}
