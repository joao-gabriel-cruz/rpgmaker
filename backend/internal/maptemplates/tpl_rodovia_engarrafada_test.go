package maptemplates

import "testing"

func TestRodoviaEngarrafada(t *testing.T) {
	checkTemplate(t, "rodovia-engarrafada")
	tpl, _ := Get("rodovia-engarrafada")
	_, objs := tpl.Generate(2)
	if prefabCount(objs, "car") < 30 || prefabCount(objs, "barricade") < 20 || prefabCount(objs, "tree") < 100 {
		t.Errorf("cars=%d barricades=%d trees=%d", prefabCount(objs, "car"), prefabCount(objs, "barricade"), prefabCount(objs, "tree"))
	}
	if tpl.Width != 96 || tpl.Depth != 64 {
		t.Errorf("size %dx%d", tpl.Width, tpl.Depth)
	}
}
