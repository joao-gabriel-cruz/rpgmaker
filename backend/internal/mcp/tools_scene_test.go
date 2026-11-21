package mcp

import (
	"errors"
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

func owned(id, name string) characters.View {
	o := "u-" + id
	return characters.View{Character: characters.Character{ID: id, Name: name, OwnerID: &o}}
}

// partyRig has an active scene "s1", an inactive 8x8 scene "s2" and two players plus one NPC.
func partyRig() *rig {
	r := liveRig(1)
	g, _ := maps.NewGrid(8, 8)
	r.sc.scenes["s2"] = &scenes.Scene{ID: "s2", Name: "Caverna", Width: 8, Depth: 8, Grid: g.Serialize()}
	r.chr.list = []characters.View{owned("c1", "Ana"), owned("c2", "Beto"), {Character: characters.Character{ID: "n1", Name: "Zumbi"}}}
	return r
}

func TestCreateSceneWithBaseTerrain(t *testing.T) {
	r := newRig()
	text, isErr := callTool(t, r.t, r.j, "create_scene", map[string]any{"name": "Pântano", "width": 16, "depth": 20, "baseTerrain": "water"})
	if isErr || !strings.Contains(text, "new1") || !strings.Contains(text, "inativa") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if len(r.sc.painted) != 1 || len(r.sc.painted[0].Cells) != 16*20 || r.sc.origins[0] != "" || r.sc.hit[0] != "new1" {
		t.Fatalf("painted=%d origins=%v", len(r.sc.painted), r.sc.origins)
	}
	if c := r.sc.painted[0].Cells[0]; c.Terrain != "water" || c.Height != 1 {
		t.Fatalf("cell = %+v", c)
	}
	if r.sc.scenes["new1"].IsActive || len(r.j.List("c1")) != 0 {
		t.Fatal("scene must stay inactive and unjournaled")
	}
}

func TestCreateSceneFloorDoesNotPaintAndValidates(t *testing.T) {
	r := newRig()
	if _, isErr := callTool(t, r.t, r.j, "create_scene", map[string]any{"name": "Sala", "width": 16, "depth": 16}); isErr || len(r.sc.painted) != 0 {
		t.Fatalf("err=%v painted=%d", isErr, len(r.sc.painted))
	}
	for _, in := range []map[string]any{
		{"name": " ", "width": 16, "depth": 16},
		{"name": "x", "width": 8, "depth": 16},
		{"name": "x", "width": 16, "depth": 129},
		{"name": "x", "width": 16, "depth": 16, "baseTerrain": "mud"},
	} {
		if text, isErr := callTool(t, r.t, r.j, "create_scene", in); !isErr || text == "" {
			t.Fatalf("input %v: err=%v text=%q", in, isErr, text)
		}
	}
	if len(r.sc.scenes) != 1 {
		t.Fatalf("invalid input created scenes: %d", len(r.sc.scenes))
	}
}

func TestActivateBringsPlayersNearArrival(t *testing.T) {
	r := partyRig()
	text, isErr := callTool(t, r.t, r.j, "activate_scene", map[string]any{"scene": "caverna", "arrival": map[string]int{"x": 3, "z": 3}})
	if isErr || !strings.Contains(text, "Ana") || !strings.Contains(text, "Beto") || strings.Contains(text, "Zumbi") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if len(r.sc.activated) != 1 || r.sc.activated[0] != "s2" || len(r.tk.placed) != 2 {
		t.Fatalf("activated=%v placed=%+v", r.sc.activated, r.tk.placed)
	}
	a, b := r.tk.placed[0], r.tk.placed[1]
	if a.Scene != "s2" || (a.X != 3 || a.Z != 3) || (a.X == b.X && a.Z == b.Z) {
		t.Fatalf("placed %+v %+v", a, b)
	}
	if max(abs(b.X-3), abs(b.Z-3)) != 1 || len(r.j.List("c1")) != 0 {
		t.Fatalf("second must be on ring 1: %+v", b)
	}
}

func TestActivateDefaultsToCenter(t *testing.T) {
	r := partyRig()
	if _, isErr := callTool(t, r.t, r.j, "activate_scene", map[string]any{"scene": "s2"}); isErr {
		t.Fatal("unexpected error")
	}
	if p := r.tk.placed[0]; p.X != 4 || p.Z != 4 {
		t.Fatalf("placed %+v", p)
	}
	if text, isErr := callTool(t, r.t, r.j, "activate_scene", map[string]any{"scene": "s2", "arrival": map[string]int{"x": 9, "z": 0}}); !isErr || !strings.Contains(text, "fora") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}

func TestActivateKeepsExistingTokens(t *testing.T) {
	r := partyRig()
	r.tk.list = []tokens.Token{{SceneID: "s2", CharacterID: "c1", Name: "Ana", X: 0, Z: 0}}
	text, isErr := callTool(t, r.t, r.j, "activate_scene", map[string]any{"scene": "s2"})
	if isErr || len(r.tk.placed) != 1 || r.tk.placed[0].Char != "c2" || strings.Contains(text, "Ana") {
		t.Fatalf("err=%v text=%q placed=%+v", isErr, text, r.tk.placed)
	}
}

func TestActivateSkipsBlockedCells(t *testing.T) {
	r := partyRig()
	g, _ := maps.NewGrid(8, 8)
	for _, c := range []maps.Coord{{X: 4, Z: 4}, {X: 3, Z: 3}} {
		g.Cells[c.Z*8+c.X].Terrain = "stone"
	}
	r.sc.scenes["s2"].Grid = g.Serialize()
	r.tk.list = []tokens.Token{{SceneID: "s2", CharacterID: "n1", X: 5, Z: 4}}
	if _, isErr := callTool(t, r.t, r.j, "activate_scene", map[string]any{"scene": "s2"}); isErr {
		t.Fatal("unexpected error")
	}
	for _, p := range r.tk.placed {
		if (p.X == 4 && p.Z == 4) || (p.X == 3 && p.Z == 3) || (p.X == 5 && p.Z == 4) {
			t.Fatalf("placed on blocked/occupied cell: %+v", p)
		}
	}
	if len(r.tk.placed) != 2 {
		t.Fatalf("placed %+v", r.tk.placed)
	}
}

func TestActivateReportsUnplaced(t *testing.T) {
	r := partyRig()
	r.tk.placeErr = map[string]error{"c1": errors.New("boom")}
	text, isErr := callTool(t, r.t, r.j, "activate_scene", map[string]any{"scene": "s2"})
	if isErr || !strings.Contains(text, "não colocado") || !strings.Contains(text, "Ana") || len(r.tk.placed) != 1 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}
