package mcp

import (
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

func addScene(r *rig, id, name string, active bool) {
	g, _ := maps.NewGrid(4, 3)
	r.sc.scenes[id] = &scenes.Scene{ID: id, Name: name, Width: 4, Depth: 3, Grid: g.Serialize(), IsActive: active}
}

func TestGetTableRendersActiveScene(t *testing.T) {
	r := newRig()
	addScene(r, "s1", "Praça", true)
	text, isErr := callTool(t, r.t, r.j, "get_table", map[string]any{})
	if isErr {
		t.Fatalf("unexpected tool error: %s", text)
	}
	if !strings.Contains(text, "Praça") || !strings.Contains(text, "z00 ") {
		t.Fatalf("table text missing scene or grid:\n%s", text)
	}
}

func TestGetTableWithoutActiveScene(t *testing.T) {
	r := newRig()
	addScene(r, "s1", "Praça", false)
	text, isErr := callTool(t, r.t, r.j, "get_table", map[string]any{})
	if isErr || !strings.Contains(text, "list_scenes") {
		t.Fatalf("want non-error text mentioning list_scenes, got err=%v %q", isErr, text)
	}
}

func TestListScenesOneLineEach(t *testing.T) {
	r := newRig()
	addScene(r, "s1", "Praça", true)
	addScene(r, "s2", "Cripta", false)
	text, isErr := callTool(t, r.t, r.j, "list_scenes", map[string]any{})
	if isErr {
		t.Fatal(text)
	}
	if !strings.Contains(text, "s1 · Praça · 4x3 · ativa") || !strings.Contains(text, "s2 · Cripta · 4x3") {
		t.Fatalf("unexpected listing:\n%s", text)
	}
}

func TestGetCatalogListsTerrains(t *testing.T) {
	text, isErr := callTool(t, newRig().t, NewJournal(&recNotifier{}), "get_catalog", map[string]any{})
	if isErr || !strings.Contains(text, "grass") {
		t.Fatalf("catalog missing terrains: err=%v %q", isErr, text)
	}
}
