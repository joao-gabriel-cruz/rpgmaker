package maptemplates

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func TestPassoNevado(t *testing.T) {
	checkTemplate(t, "passo-nevado")
	tpl, _ := Get("passo-nevado")
	for seed := int64(1); seed <= 6; seed++ {
		g, objs := tpl.Generate(seed)
		if maxHeight(g) < 7 || terrainCount(g, "stone") < 300 || prefabCount(objs, "tree") < 40 {
			t.Fatalf("seed %d: peak=%d stone=%d pines=%d", seed, maxHeight(g), terrainCount(g, "stone"), prefabCount(objs, "tree"))
		}
		reach := ReachSet(g, objs)
		if !reach[Pt{40, 40}] || !reach[Pt{31, 3}] || !reach[Pt{25, 42}] || !reach[Pt{57, 66}] {
			t.Fatalf("seed %d: the pass, its far end and both ledges must be reachable", seed)
		}
		if prefabCount(objs, "bridge") < 8 || prefabCount(objs, "campfire") != 1 || prefabCount(objs, "door") != 1 {
			t.Fatalf("seed %d: bridge=%d campfire=%d door=%d", seed, prefabCount(objs, "bridge"), prefabCount(objs, "campfire"), prefabCount(objs, "door"))
		}
		if reachableWithout(g, objs, "bridge")[Pt{31, 3}] {
			t.Fatalf("seed %d: the fissure must only be crossed by the bridge", seed)
		}
	}
}

// reachableWithout is ReachSet with every object of a prefab removed.
func reachableWithout(g maps.Grid, objs []maps.Object, prefab string) map[Pt]bool {
	var kept []maps.Object
	for _, o := range objs {
		if o.Prefab != prefab {
			kept = append(kept, o)
		}
	}
	return ReachSet(g, kept)
}
