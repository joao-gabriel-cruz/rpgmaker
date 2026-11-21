package maptemplates

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func TestFlorestaClareira(t *testing.T) {
	checkTemplate(t, "floresta-clareira")
	tpl, _ := Get("floresta-clareira")
	for seed := int64(1); seed <= 4; seed++ {
		g, objs := tpl.Generate(seed)
		if prefabCount(objs, "tree") < 400 || terrainCount(g, "water") < 60 {
			t.Errorf("seed %d: expected dense woods and a stream: trees=%d water=%d", seed, prefabCount(objs, "tree"), terrainCount(g, "water"))
		}
		if bridgesOverWater(g, objs) < 2 {
			t.Errorf("seed %d: expected a plank bridge over the stream", seed)
		}
		if prefabCount(objs, "campfire") != 1 || prefabCount(objs, "tent") < 3 || prefabCount(objs, "barrel") < 3 || prefabCount(objs, "crate") < 3 {
			t.Errorf("seed %d: the clearing camp is incomplete", seed)
		}
	}
}

func bridgesOverWater(g maps.Grid, objs []maps.Object) int {
	n := 0
	for _, o := range objs {
		if c, _ := g.At(o.X, o.Z); o.Prefab == "bridge" && c.Terrain == "water" {
			n++
		}
	}
	return n
}
