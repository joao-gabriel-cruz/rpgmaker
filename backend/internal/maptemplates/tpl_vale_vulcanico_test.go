package maptemplates

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func TestValeVulcanico(t *testing.T) {
	checkTemplate(t, "vale-vulcanico")
	tpl, _ := Get("vale-vulcanico")
	for seed := int64(1); seed <= 6; seed++ {
		g, objs := tpl.Generate(seed)
		if prefabCount(objs, "volcano") != 1 || terrainCount(g, "lava") < 30 || prefabCount(objs, "watchtower") != 1 {
			t.Fatalf("seed %d: volcano=%d lava=%d tower=%d", seed, prefabCount(objs, "volcano"), terrainCount(g, "lava"), prefabCount(objs, "watchtower"))
		}
		if prefabCount(objs, "wall") < 30 || prefabCount(objs, "crate") < 4 || maxHeight(g) < 7 {
			t.Fatalf("seed %d: station walls=%d crates=%d top=%d", seed, prefabCount(objs, "wall"), prefabCount(objs, "crate"), maxHeight(g))
		}
		if !walkable(g, objs, Pt{48, 60}) || !walkable(g, objs, Pt{82, 16}) {
			t.Fatalf("seed %d: the safe path must reach the station and the lookout", seed)
		}
	}
}

func maxHeight(g maps.Grid) int {
	m := 0
	for _, c := range g.Cells {
		m = max(m, c.Height)
	}
	return m
}

// walkable reports whether a cell is reachable from the entry area.
func walkable(g maps.Grid, objs []maps.Object, to Pt) bool { return ReachSet(g, objs)[to] }
