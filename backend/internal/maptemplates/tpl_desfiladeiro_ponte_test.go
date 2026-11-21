package maptemplates

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func TestDesfiladeiroPonte(t *testing.T) {
	checkTemplate(t, "desfiladeiro-ponte")
	tpl, _ := Get("desfiladeiro-ponte")
	for seed := int64(1); seed <= 6; seed++ {
		g, objs := tpl.Generate(seed)
		if n := bridgeSpans2(objs); n != 1 {
			t.Fatalf("seed %d: %d bridges, want exactly one", seed, n)
		}
		if terrainCount(g, "stone") < 300 || terrainCount(g, "water") < 150 || maxHeight(g) < 7 {
			t.Fatalf("seed %d: stone=%d water=%d", seed, terrainCount(g, "stone"), terrainCount(g, "water"))
		}
		cz := canyonZ(NewBuilder(64, 96, seed, "grass", 1), canyonBridge)
		north, river := Pt{32, cz - 18}, Pt{36, canyonZ(NewBuilder(64, 96, seed, "grass", 1), 36)}
		if !ReachSet(g, objs)[north] {
			t.Fatalf("seed %d: the far rim must be reachable by the bridge", seed)
		}
		dry := reachableWithout(g, objs, "bridge")
		if dry[north] || dry[Pt{32, 3}] {
			t.Fatalf("seed %d: the far rim must only be reachable by the bridge", seed)
		}
		if !dry[river] {
			t.Fatalf("seed %d: the river bed must be reachable by the steep trail", seed)
		}
		if prefabCount(objs, "tent") < 2 || prefabCount(objs, "rock") < 30 || prefabCount(objs, "bush") < 30 {
			t.Fatalf("seed %d: camp or ambush cover missing", seed)
		}
	}
}

// bridgeSpans2 counts groups of 4-adjacent bridge planks.
func bridgeSpans2(objs []maps.Object) int {
	cells := map[Pt]bool{}
	for _, o := range objs {
		if o.Prefab == "bridge" {
			cells[Pt{o.X, o.Z}] = true
		}
	}
	groups := 0
	for p := range cells {
		if !cells[p] {
			continue
		}
		groups++
		stack := []Pt{p}
		delete(cells, p)
		for len(stack) > 0 {
			cur := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, n := range []Pt{{cur.X + 1, cur.Z}, {cur.X - 1, cur.Z}, {cur.X, cur.Z + 1}, {cur.X, cur.Z - 1}} {
				if cells[n] {
					delete(cells, n)
					stack = append(stack, n)
				}
			}
		}
	}
	return groups
}
