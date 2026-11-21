package maptemplates

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func TestFestival(t *testing.T) {
	checkTemplate(t, "festival")
	tpl, _ := Get("festival")
	for seed := int64(1); seed <= 4; seed++ {
		g, objs := tpl.Generate(seed)
		if prefabCount(objs, "market_stall") < 6 || stallColors(objs) < 4 {
			t.Errorf("seed %d: expected >=6 stalls in >=4 colors: %d/%d", seed, prefabCount(objs, "market_stall"), stallColors(objs))
		}
		if !hasObjectAt(objs, "campfire", 32, 30) || prefabCount(objs, "lamp_post") < 12 || prefabCount(objs, "bunting") < 10 || prefabCount(objs, "bench") < 4 {
			t.Errorf("seed %d: expected a central bonfire, lamps, bunting and benches", seed)
		}
		if c, _ := g.At(32, 15); c.Height < 3 {
			t.Errorf("seed %d: stage is not raised (height %d)", seed, c.Height)
		}
		if prefabCount(objs, "door") < 7 {
			t.Errorf("seed %d: expected houses around the plaza", seed)
		}
	}
}

func stallColors(objs []maps.Object) int {
	seen := map[any]bool{}
	for _, o := range objs {
		if o.Prefab == "market_stall" {
			seen[o.Props["color"]] = true
		}
	}
	return len(seen)
}

func hasObjectAt(objs []maps.Object, prefab string, x, z int) bool {
	for _, o := range objs {
		if o.Prefab == prefab && o.X == x && o.Z == z {
			return true
		}
	}
	return false
}
