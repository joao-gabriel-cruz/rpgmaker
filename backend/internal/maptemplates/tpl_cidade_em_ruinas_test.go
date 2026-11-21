package maptemplates

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func tallestCell(cells []maps.Cell) int {
	top := 0
	for _, c := range cells {
		top = max(top, c.Height)
	}
	return top
}

func TestCidadeEmRuinas(t *testing.T) {
	checkTemplate(t, "cidade-em-ruinas")
	tpl, _ := Get("cidade-em-ruinas")
	for seed := int64(1); seed <= 4; seed++ {
		g, objs := tpl.Generate(seed)
		if prefabCount(objs, "door") < 8 || prefabCount(objs, "car") < 8 || prefabCount(objs, "sandbags") < 6 {
			t.Errorf("seed %d: buildings/cars/sandbags: %d/%d/%d", seed, prefabCount(objs, "door"), prefabCount(objs, "car"), prefabCount(objs, "sandbags"))
		}
		if prefabCount(objs, "well") != 1 || prefabCount(objs, "bench") < 3 || prefabCount(objs, "lamp_post") < 4 {
			t.Errorf("seed %d: plaza is missing its well, benches or lamps", seed)
		}
		if prefabCount(objs, "watchtower") != 1 || tallestCell(g.Cells) < 4 {
			t.Errorf("seed %d: expected an elevated observation tower", seed)
		}
		if terrainCount(g, "floor") < 1500 || prefabCount(objs, "wall")+prefabCount(objs, "ruined_wall") < 200 {
			t.Errorf("seed %d: expected streets and walls", seed)
		}
	}
}
