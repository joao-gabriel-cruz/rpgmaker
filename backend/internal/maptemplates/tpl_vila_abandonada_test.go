package maptemplates

import "testing"

func TestVilaAbandonada(t *testing.T) {
	checkTemplate(t, "vila-abandonada")
	tpl, _ := Get("vila-abandonada")
	g, objs := tpl.Generate(3)
	if prefabCount(objs, "door") < 5 || prefabCount(objs, "wall") < 60 || prefabCount(objs, "car") < 3 {
		t.Errorf("expected houses and cars: doors=%d walls=%d cars=%d", prefabCount(objs, "door"), prefabCount(objs, "wall"), prefabCount(objs, "car"))
	}
	if prefabCount(objs, "watchtower") != 1 || terrainCount(g, "water") < 100 {
		t.Errorf("expected a watchtower and a river: water=%d", terrainCount(g, "water"))
	}
}
