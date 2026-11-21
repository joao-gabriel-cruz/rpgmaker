package mcp

import (
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/ambience"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

func tokenRig() *rig {
	r := liveRig(1)
	r.tk.list = []tokens.Token{
		{ID: "t1", Name: "Junior", X: 1, Z: 1},
		{ID: "t2", Name: "Marta", X: 2, Z: 2},
		{ID: "t3", Name: "Marta", X: 3, Z: 3},
	}
	return r
}

func TestMoveTokenByNameCaseInsensitive(t *testing.T) {
	r := tokenRig()
	text, isErr := callTool(t, r.t, r.j, "move_token", map[string]any{"token": "junior", "x": 5, "z": 7})
	if isErr || !strings.Contains(text, "Junior moveu para (5,7)") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if len(r.tk.moves) != 1 || r.tk.moves[0] != (TokenPos{"t1", 5, 7}) {
		t.Fatalf("moves %+v", r.tk.moves)
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	if r.tk.moves[1] != (TokenPos{"t1", 1, 1}) {
		t.Fatalf("inverse %+v", r.tk.moves)
	}
}

func TestMoveTokenAmbiguousName(t *testing.T) {
	r := tokenRig()
	text, isErr := callTool(t, r.t, r.j, "move_token", map[string]any{"token": "marta", "x": 5, "z": 5})
	if !isErr || !strings.Contains(text, "ambíguo") || !strings.Contains(text, "t2") || !strings.Contains(text, "t3") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if len(r.tk.moves) != 0 || len(r.j.List("c1")) != 0 {
		t.Fatal("nothing may change")
	}
	if _, isErr = callTool(t, r.t, r.j, "move_token", map[string]any{"token": "t3", "x": 5, "z": 5}); isErr {
		t.Fatal("exact id must disambiguate")
	}
	text, isErr = callTool(t, r.t, r.j, "move_token", map[string]any{"token": "Zé", "x": 5, "z": 5})
	if !isErr || !strings.Contains(text, `não encontrei "Zé"`) {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}

func TestMoveTokenIntoBlockedCellRefused(t *testing.T) {
	r := tokenRig()
	r.sc.scenes["s1"].Objects = []maps.Object{{ID: "w", Prefab: "wall", X: 4, Z: 4}}
	text, isErr := callTool(t, r.t, r.j, "move_token", map[string]any{"token": "Junior", "x": 4, "z": 4})
	if !isErr || !strings.Contains(text, "bloqueando") || len(r.tk.moves) != 0 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	_, isErr = callTool(t, r.t, r.j, "move_token", map[string]any{"token": "Junior", "x": 99, "z": 4})
	if !isErr || len(r.j.List("c1")) != 0 {
		t.Fatal("out of map must be refused")
	}
}

func TestSetAmbienceMergesAndUndo(t *testing.T) {
	r := tokenRig()
	r.amb.cur = ambience.Ambience{SceneID: "s1", TimeOfDay: "day", Weather: "clear"}
	text, isErr := callTool(t, r.t, r.j, "set_ambience", map[string]any{"timeOfDay": "night", "weather": "storm"})
	if isErr || !strings.Contains(text, "ambiente: noite, tempestade") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	_, isErr = callTool(t, r.t, r.j, "set_ambience", map[string]any{"weather": "fog"})
	if isErr || r.amb.set[1].TimeOfDay != "day" || r.amb.set[1].Weather != "fog" {
		t.Fatalf("merge failed %+v", r.amb.set)
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	if got := r.amb.set[2]; got.Weather != "clear" || got.TimeOfDay != "day" {
		t.Fatalf("inverse %+v", got)
	}
	for _, bad := range []map[string]any{{}, {"weather": "meteor"}, {"timeOfDay": "noon"}} {
		if _, isErr = callTool(t, r.t, r.j, "set_ambience", bad); !isErr {
			t.Fatalf("%v must be refused", bad)
		}
	}
}

func charView(id, name, sheet string) characters.View {
	return characters.View{Character: characters.Character{ID: id, Name: name, Sheet: []byte(sheet)}}
}

func TestSetConditionNoopNotRecorded(t *testing.T) {
	r := tokenRig()
	r.chr.list = []characters.View{charView("c9", "Junior", `{"conditions":["scared"]}`)}
	text, isErr := callTool(t, r.t, r.j, "set_condition", map[string]any{"character": "junior", "condition": "scared", "active": true})
	if isErr || text != "já estava assim" || len(r.amb.conditions) != 0 || len(r.j.List("c1")) != 0 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	text, isErr = callTool(t, r.t, r.j, "set_condition", map[string]any{"character": "Junior", "condition": "scared", "active": false})
	if isErr || !strings.Contains(text, "Junior: condição scared removida") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	if got := r.amb.conditions; len(got) != 2 || got[1] != (ConditionState{"c9", "scared", true}) {
		t.Fatalf("conditions %+v", got)
	}
	if _, isErr = callTool(t, r.t, r.j, "set_condition", map[string]any{"character": "Junior", "condition": "sleepy", "active": true}); !isErr {
		t.Fatal("unknown condition must be refused")
	}
}

func TestUndoLastEmptyJournal(t *testing.T) {
	r := tokenRig()
	text, isErr := callTool(t, r.t, r.j, "undo_last", map[string]any{})
	if !isErr || !strings.Contains(text, "não existe mais") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	callTool(t, r.t, r.j, "move_token", map[string]any{"token": "Junior", "x": 5, "z": 7})
	text, isErr = callTool(t, r.t, r.j, "undo_last", map[string]any{})
	if isErr || text != "desfeito: Junior moveu para (5,7)" {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}

func TestTokenToolsRefuseOpenDraft(t *testing.T) {
	r := tokenRig()
	r.sc.scenes["s1"].Draft = &scenes.Draft{}
	_, isErr := callTool(t, r.t, r.j, "move_token", map[string]any{"token": "Junior", "x": 5, "z": 7})
	_, isErr2 := callTool(t, r.t, r.j, "set_ambience", map[string]any{"weather": "rain"})
	if !isErr || !isErr2 {
		t.Fatal("draft must refuse")
	}
}
