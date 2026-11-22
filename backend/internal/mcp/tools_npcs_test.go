package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

func at(x, z int) map[string]any { return map[string]any{"x": x, "z": z} }

func TestCreateNPCWithSheetFields(t *testing.T) {
	r := liveRig(1)
	text, isErr := callTool(t, r.t, r.j, "create_npc", map[string]any{
		"name": "  Brutamontes ", "instincts": map[string]any{"potency": 4},
		"threats": []string{"Golpe esmagador"}, "boxesPerLevel": 5,
	})
	if isErr || !strings.Contains(text, "1 NPC(s) criado(s): Brutamontes") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	var sh assimilacao.NPCSheet
	if err := json.Unmarshal(r.chr.saved["npc1"], &sh); err != nil {
		t.Fatalf("saved sheet: %v", err)
	}
	if sh.Instincts.Potency != 4 || sh.Instincts.Reaction != 1 || sh.Health.BoxesPerLevel != 5 ||
		len(sh.Threats) != 1 || sh.Identity.Name != "Brutamontes" {
		t.Fatalf("merged sheet = %+v", sh)
	}
}

func TestCreateNPCCountNamesAndPlacement(t *testing.T) {
	r := liveRig(1)
	text, isErr := callTool(t, r.t, r.j, "create_npc", map[string]any{"name": "Infectado", "count": 3, "at": at(4, 4)})
	if isErr || !strings.Contains(text, "3 NPC(s) criado(s): Infectado 1, Infectado 2, Infectado 3") ||
		!strings.Contains(text, "colocados perto de (4,4)") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if len(r.chr.saved) != 0 {
		t.Fatal("no sheet field given, Save must not run")
	}
	seen := map[maps.Coord]bool{}
	for _, p := range r.tk.placed {
		c := maps.Coord{X: p.X, Z: p.Z}
		if seen[c] || max(abs(p.X-4), abs(p.Z-4)) > 1 {
			t.Fatalf("bad placement %+v", r.tk.placed)
		}
		seen[c] = true
	}
	if len(r.tk.placed) != 3 || len(r.j.List("c1")) != 1 {
		t.Fatalf("placed=%d journal=%d", len(r.tk.placed), len(r.j.List("c1")))
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	if len(r.tk.removed) != 3 || len(r.tk.list) != 0 || len(r.chr.list) != 3 {
		t.Fatalf("removed=%v tokens=%d chars=%d", r.tk.removed, len(r.tk.list), len(r.chr.list))
	}
}

func TestCreateNPCWithoutAtNoJournal(t *testing.T) {
	r := newRig() // no scene at all: creation is not tied to one
	text, isErr := callTool(t, r.t, r.j, "create_npc", map[string]any{"name": "Rato"})
	if isErr || strings.Contains(text, "colocados") || len(r.tk.placed) != 0 || len(r.j.List("c1")) != 0 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}

func TestCreateNPCValidation(t *testing.T) {
	r := liveRig(1)
	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"empty name", map[string]any{"name": "  "}, "nome"},
		{"count 11", map[string]any{"name": "X", "count": 11}, "entre 1 e 10"},
		{"outside grid", map[string]any{"name": "X", "at": at(50, 50)}, "fora da cena"},
		{"bad instinct", map[string]any{"name": "X", "instincts": map[string]any{"luck": 1}}, "luck"},
		{"accented long name", map[string]any{"name": strings.Repeat("é", 60)}, "entre 1 e 100"},
		{"suffix overflow", map[string]any{"name": strings.Repeat("a", 98), "count": 10}, "entre 1 e 100"},
		{"aptitude 100", map[string]any{"name": "X", "practices": map[string]any{"weapons": 100}}, "0 e 99"},
	}
	for _, c := range cases {
		text, isErr := callTool(t, r.t, r.j, "create_npc", c.args)
		if !isErr || !strings.Contains(text, c.want) {
			t.Fatalf("%s: err=%v text=%q", c.name, isErr, text)
		}
	}
	if len(r.chr.list) != 0 {
		t.Fatal("nothing may be created")
	}
}

func charRig() *rig {
	r := liveRig(1)
	r.chr.list = []characters.View{
		{Character: characters.Character{ID: "c-orc", Name: "Orc", Kind: "npc"}},
		{Character: characters.Character{ID: "c-ana", Name: "Ana"}},
	}
	return r
}

func TestPlaceTokenAndUndo(t *testing.T) {
	r := charRig()
	text, isErr := callTool(t, r.t, r.j, "place_token", map[string]any{"character": "orc", "x": 3, "z": 2})
	if isErr || !strings.Contains(text, "Orc colocado em (3,2)") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if len(r.tk.placed) != 1 || r.tk.placed[0] != (placement{"s1", "c-orc", 3, 2}) {
		t.Fatalf("placed %+v", r.tk.placed)
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	if len(r.tk.removed) != 1 || len(r.tk.list) != 0 {
		t.Fatalf("removed=%v", r.tk.removed)
	}
}

func TestPlaceTokenRefusesOccupiedOrBlocked(t *testing.T) {
	r := charRig()
	r.tk.list = []tokens.Token{{ID: "t1", Name: "Ana", CharacterID: "c-ana", X: 1, Z: 1}}
	text, isErr := callTool(t, r.t, r.j, "place_token", map[string]any{"character": "c-orc", "x": 1, "z": 1})
	if !isErr || !strings.Contains(text, "ocupada") {
		t.Fatalf("occupied: err=%v text=%q", isErr, text)
	}
	sc := r.sc.scenes["s1"]
	sc.Objects = []maps.Object{{ID: "w", Prefab: "wall", X: 5, Z: 5}}
	text, isErr = callTool(t, r.t, r.j, "place_token", map[string]any{"character": "c-orc", "x": 5, "z": 5})
	if !isErr || !strings.Contains(text, "bloqueando") {
		t.Fatalf("blocked: err=%v text=%q", isErr, text)
	}
	text, isErr = callTool(t, r.t, r.j, "place_token", map[string]any{"character": "c-orc", "x": 99, "z": 0})
	if !isErr || !strings.Contains(text, "fora da cena") {
		t.Fatalf("outside: err=%v text=%q", isErr, text)
	}
	if len(r.tk.placed) != 0 || len(r.j.List("c1")) != 0 {
		t.Fatal("nothing may change")
	}
}

func TestPlaceTokenAlreadyOnScene(t *testing.T) {
	r := charRig()
	r.tk.list = []tokens.Token{{ID: "t1", Name: "Ana", CharacterID: "c-ana", X: 1, Z: 1}}
	text, isErr := callTool(t, r.t, r.j, "place_token", map[string]any{"character": "Ana", "x": 4, "z": 4})
	if !isErr || !strings.Contains(text, "Ana já está na cena") || len(r.tk.placed) != 0 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}

func TestRemoveTokenAndUndo(t *testing.T) {
	r := charRig()
	r.tk.list = []tokens.Token{{ID: "t1", Name: "Orc", CharacterID: "c-orc", X: 2, Z: 3}}
	text, isErr := callTool(t, r.t, r.j, "remove_token", map[string]any{"token": "orc"})
	if isErr || !strings.Contains(text, "Orc removido da cena") || len(r.tk.removed) != 1 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	if len(r.tk.placed) != 1 || r.tk.placed[0] != (placement{"s1", "c-orc", 2, 3}) {
		t.Fatalf("placed %+v", r.tk.placed)
	}
}

func TestNPCToolsRefuseDraft(t *testing.T) {
	r := charRig()
	r.sc.scenes["s1"].Draft = &scenes.Draft{}
	r.tk.list = []tokens.Token{{ID: "t1", Name: "Ana", CharacterID: "c-ana", X: 1, Z: 1}}
	calls := map[string]map[string]any{
		"place_token":  {"character": "c-orc", "x": 3, "z": 3},
		"remove_token": {"token": "t1"},
		"create_npc":   {"name": "X", "at": at(3, 3)},
	}
	for name, args := range calls {
		text, isErr := callTool(t, r.t, r.j, name, args)
		if !isErr || !strings.Contains(text, "rascunho") {
			t.Fatalf("%s: err=%v text=%q", name, isErr, text)
		}
	}
	if len(r.tk.placed) != 0 || len(r.tk.removed) != 0 || len(r.chr.list) != 2 {
		t.Fatal("nothing may change")
	}
}

func TestUndoCreateNPCSurvivesManuallyRemovedToken(t *testing.T) {
	r := liveRig(1)
	if _, isErr := callTool(t, r.t, r.j, "create_npc", map[string]any{"name": "Inf", "count": 3, "at": at(4, 4)}); isErr {
		t.Fatal("create failed")
	}
	if err := r.tk.Remove(t.Context(), r.s.Sender, "c1", r.tk.list[1].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if len(r.tk.list) != 0 || len(r.tk.removed) != 3 {
		t.Fatalf("tokens=%d removed=%v", len(r.tk.list), r.tk.removed)
	}
}

func TestUndoInReverseRestoresEmptyMap(t *testing.T) {
	r := liveRig(1)
	if _, isErr := callTool(t, r.t, r.j, "create_npc", map[string]any{"name": "Inf", "count": 3, "at": at(4, 4)}); isErr {
		t.Fatal("create failed")
	}
	if _, isErr := callTool(t, r.t, r.j, "remove_token", map[string]any{"token": "npc2"}); isErr {
		t.Fatal("remove failed")
	}
	for range 2 {
		if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
			t.Fatalf("undo: %v", err)
		}
	}
	if len(r.tk.list) != 0 || len(r.chr.list) != 3 || len(r.j.List("c1")) != 0 {
		t.Fatalf("tokens=%+v chars=%d journal=%d", r.tk.list, len(r.chr.list), len(r.j.List("c1")))
	}
}
