package mcp

import (
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/ambience"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

func viewFor(t *testing.T, w, d int, patch ...maps.CellPatch) TableView {
	t.Helper()
	g, err := maps.NewGrid(w, d)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range patch {
		g.Cells[p.Z*w+p.X] = maps.Cell{Terrain: p.Terrain, Height: p.Height}
	}
	return tableOf(w, d, g)
}

func tableOf(w, d int, g *maps.Grid) TableView {
	return TableView{
		Scene:    scenes.Scene{Name: "Rua", Width: w, Depth: d, Grid: g.Serialize()},
		Ambience: ambience.Default("s"),
	}
}

func TestRenderTableFlatGrid(t *testing.T) {
	out, err := RenderTable(viewFor(t, 3, 2))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "z00 ...\n") || !strings.Contains(out, "z01 ...") {
		t.Fatalf("grid lines missing:\n%s", out)
	}
	if strings.Contains(out, "Alturas") || !strings.Contains(out, "altura uniforme: 1") {
		t.Fatalf("expected uniform height line only:\n%s", out)
	}
	if !strings.Contains(out, "x cresce para a direita (leste); z cresce para baixo (sul); norte = z menor") {
		t.Fatalf("axis line missing:\n%s", out)
	}
}

func TestRenderTableLavaAndHeights(t *testing.T) {
	g, _ := maps.NewGrid(3, 1)
	for i := range g.Cells {
		g.Cells[i].Height = 0
	}
	g.Cells[1] = maps.Cell{Terrain: "lava", Height: 2}
	out, _ := RenderTable(tableOf(3, 1, g))
	if !strings.Contains(out, "z00 .!.") || !strings.Contains(out, "z00 020") {
		t.Fatalf("lava/heights missing:\n%s", out)
	}
}

func TestRenderTableTokensAndObjects(t *testing.T) {
	v := viewFor(t, 4, 4)
	v.Tokens = []TokenInfo{
		{ID: "t1", Name: "Junior", X: 1, Z: 2, Player: true, Conditions: []string{"wounded"}},
		{ID: "t2", Name: "Zumbi", X: 3, Z: 0},
	}
	v.Scene.Objects = []maps.Object{{ID: "o1", Prefab: "door", X: 2, Z: 1, Rotation: 90, Props: map[string]any{"open": true}}}
	out, _ := RenderTable(v)
	for _, want := range []string{
		"- Junior (jogador) em (1,2) [wounded]", "- Zumbi (NPC) em (3,0)", "- o1 door em (2,1) rot 90 aberta",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestRenderCatalogListsEverything(t *testing.T) {
	out := RenderCatalog()
	var ids []string
	for _, tr := range maps.TerrainOrder {
		ids = append(ids, tr.ID)
	}
	for id := range maps.PrefabRules() {
		ids = append(ids, id)
	}
	ids = append(ids, assimilacao.Conditions...)
	ids = append(ids, ambience.TimesOfDay()...)
	ids = append(ids, ambience.Weathers()...)
	for _, id := range ids {
		if !strings.Contains(out, id) {
			t.Fatalf("catalog missing %q:\n%s", id, out)
		}
	}
}

func TestRenderTableHeightGridOnlyWhenVarying(t *testing.T) {
	out, _ := RenderTable(viewFor(t, 3, 1, maps.CellPatch{X: 1, Z: 0, Terrain: "floor", Height: 2}))
	if !strings.Contains(out, "Alturas") || !strings.Contains(out, "z00 121") || strings.Contains(out, "altura uniforme") {
		t.Fatalf("expected heights grid:\n%s", out)
	}
}
