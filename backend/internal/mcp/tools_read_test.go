package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
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

type failingScenes struct{ *fakeScenes }

func (failingScenes) List(context.Context, string) ([]scenes.Summary, error) {
	return nil, errors.New("pgx: connection refused to 10.0.0.1")
}

func TestToolErrorHidesInternals(t *testing.T) {
	r := newRig()
	r.t.Scenes = failingScenes{r.sc}
	for _, tool := range []string{"get_table", "list_scenes"} {
		text, isErr := callTool(t, r.t, r.j, tool, map[string]any{})
		if !isErr || text != "Erro interno. Tente novamente" {
			t.Fatalf("%s: want generic PT-BR error, got err=%v %q", tool, isErr, text)
		}
	}
}

func TestToolErrorMapping(t *testing.T) {
	cases := map[string]error{
		"Cena não encontrada":            fmt.Errorf("get: %w", scenes.ErrNotFound),
		"Mudança inválida no mapa: fora": fmt.Errorf("%w: fora", scenes.ErrInvalidPatch),
		"Sem permissão":                  realtime.UserError("Sem permissão"),
		"nenhuma cena ativa":             errNoActiveScene,
		"a forma fica toda fora do mapa": newUserError("a forma fica toda fora do mapa"),
		"Erro interno. Tente novamente":  errors.New("boom"),
	}
	for want, in := range cases {
		if got := toolError(in).Error(); got != want {
			t.Errorf("toolError(%v) = %q, want %q", in, got, want)
		}
	}
}
