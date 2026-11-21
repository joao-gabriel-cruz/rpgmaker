package maptemplates

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func TestRefugioEscola(t *testing.T) {
	checkTemplate(t, "refugio-escola")
	tpl, _ := Get("refugio-escola")
	for seed := int64(1); seed <= 4; seed++ {
		_, objs := tpl.Generate(seed)
		if gaps := perimeterGaps(objs, refugioYard); gaps != 0 {
			t.Errorf("seed %d: compound wall has %d open cells", seed, gaps)
		}
		gate := 0
		for _, o := range objs {
			if o.Prefab == "door" && o.Z == refugioYard.Z+refugioYard.D-1 {
				gate++
			}
		}
		if gate != 2 || prefabCount(objs, "watchtower") < 4 {
			t.Errorf("seed %d: expected one gate and corner towers: gate=%d towers=%d", seed, gate, prefabCount(objs, "watchtower"))
		}
		if prefabCount(objs, "sandbags") < 8 || prefabCount(objs, "tent") < 4 || prefabCount(objs, "lamp_post") < 8 || prefabCount(objs, "campfire") < 2 || prefabCount(objs, "door") < 7 {
			t.Errorf("seed %d: expected sandbags, tents, lamps, fires and school rooms: %d/%d/%d/%d/%d", seed, prefabCount(objs, "sandbags"), prefabCount(objs, "tent"), prefabCount(objs, "lamp_post"), prefabCount(objs, "campfire"), prefabCount(objs, "door"))
		}
	}
}

// perimeterGaps counts compound perimeter cells with no wall, tower or door.
func perimeterGaps(objs []maps.Object, r Rect) int {
	at := map[Pt]bool{}
	for _, o := range objs {
		if o.Prefab == "wall" || o.Prefab == "watchtower" || o.Prefab == "door" {
			at[Pt{o.X, o.Z}] = true
		}
	}
	gaps := 0
	for z := r.Z; z < r.Z+r.D; z++ {
		for x := r.X; x < r.X+r.W; x++ {
			if (x == r.X || z == r.Z || x == r.X+r.W-1 || z == r.Z+r.D-1) && !at[Pt{x, z}] {
				gaps++
			}
		}
	}
	return gaps
}
