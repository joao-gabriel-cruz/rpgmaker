package maptemplates

import "testing"

func TestFazenda(t *testing.T) {
	checkTemplate(t, "fazenda")
	tpl, _ := Get("fazenda")
	for seed := int64(1); seed <= 4; seed++ {
		g, objs := tpl.Generate(seed)
		if prefabCount(objs, "fence") < 150 || prefabCount(objs, "well") != 1 || prefabCount(objs, "door") < 2 {
			t.Errorf("seed %d: expected fences, a well and two buildings: fence=%d door=%d", seed, prefabCount(objs, "fence"), prefabCount(objs, "door"))
		}
		if terrainCount(g, "dirt") < 700 || terrainCount(g, "grass") < 1500 || prefabCount(objs, "bush") < 80 {
			t.Errorf("seed %d: expected crop rows of dirt between grass", seed)
		}
		if prefabCount(objs, "tree") < 20 || prefabCount(objs, "car") < 2 || prefabCount(objs, "sign") < 3 {
			t.Errorf("seed %d: expected orchard, wrecks and signs", seed)
		}
	}
}
