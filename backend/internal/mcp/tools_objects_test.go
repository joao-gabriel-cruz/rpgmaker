package mcp

import (
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

func xz(x, z int) map[string]any { return map[string]any{"x": x, "z": z} }

func TestPlaceWallLine(t *testing.T) {
	r := liveRig(1)
	text, isErr := callTool(t, r.t, r.j, "place_objects", map[string]any{
		"line": map[string]any{"prefab": "wall", "from": xz(2, 2), "to": xz(2, 5)},
	})
	if isErr {
		t.Fatal(text)
	}
	add := r.sc.objects[0].Add
	if len(add) != 4 {
		t.Fatalf("added %d, want 4", len(add))
	}
	for _, o := range add {
		if !strings.HasPrefix(o.ID, "mcp-") || len(o.ID) != 16 || o.Prefab != "wall" {
			t.Fatalf("bad object %+v", o)
		}
	}
	if len(r.j.List("c1")) != 1 {
		t.Fatal("want one journal entry")
	}
}

func TestPlaceLineRejectsMultiCellPrefab(t *testing.T) {
	r := liveRig(1)
	text, isErr := callTool(t, r.t, r.j, "place_objects", map[string]any{
		"line": map[string]any{"prefab": "car", "from": xz(0, 0), "to": xz(3, 0)},
	})
	if !isErr || !strings.Contains(text, "1x1") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}

func TestPlaceBlockingOnTokenRefused(t *testing.T) {
	r := liveRig(1)
	r.tk.list = []tokens.Token{{ID: "t1", Name: "Junior", X: 3, Z: 3}}
	text, isErr := callTool(t, r.t, r.j, "place_objects", map[string]any{
		"objects": []map[string]any{{"prefab": "wall", "x": 3, "z": 3}},
	})
	if !isErr || !strings.Contains(text, "Junior") || len(r.sc.objects) != 0 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	// a 2x1 car at (2,3) covers (3,3) too
	_, isErr = callTool(t, r.t, r.j, "place_objects", map[string]any{
		"objects": []map[string]any{{"prefab": "car", "x": 2, "z": 3}},
	})
	if !isErr {
		t.Fatal("car footprint over token must be refused")
	}
	text, isErr = callTool(t, r.t, r.j, "place_objects", map[string]any{
		"objects": []map[string]any{{"prefab": "bush", "x": 3, "z": 3}},
	})
	if isErr {
		t.Fatalf("bush on token must be allowed: %s", text)
	}
}

func TestPlaceRejectsUnknownPrefabAndRefusesDraft(t *testing.T) {
	r := liveRig(1)
	text, isErr := callTool(t, r.t, r.j, "place_objects", map[string]any{
		"objects": []map[string]any{{"prefab": "ufo", "x": 1, "z": 1}},
	})
	if !isErr || !strings.Contains(text, "ufo") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	r.sc.scenes["s1"].Draft = &scenes.Draft{}
	_, isErr = callTool(t, r.t, r.j, "place_objects", map[string]any{
		"objects": []map[string]any{{"prefab": "bush", "x": 1, "z": 1}},
	})
	if !isErr {
		t.Fatal("draft must refuse")
	}
}

func TestRemoveByAreaAndUndo(t *testing.T) {
	r := liveRig(1)
	r.sc.scenes["s1"].Objects = []maps.Object{
		{ID: "a", Prefab: "wall", X: 1, Z: 1},
		{ID: "b", Prefab: "car", X: 3, Z: 1}, // covers (3,1),(4,1)
		{ID: "c", Prefab: "wall", X: 7, Z: 7},
	}
	text, isErr := callTool(t, r.t, r.j, "remove_objects", map[string]any{
		"area": map[string]any{"kind": "rect", "from": xz(4, 0), "to": xz(5, 2)},
	})
	if isErr || !strings.Contains(text, "1 objeto") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if rm := r.sc.objects[0].Remove; len(rm) != 1 || rm[0] != "b" {
		t.Fatalf("removed %v", rm)
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	if add := r.sc.objects[1].Add; len(add) != 1 || add[0].ID != "b" || add[0].X != 3 {
		t.Fatalf("restored %+v", add)
	}
}

func TestRemoveByIDsNothingFound(t *testing.T) {
	r := liveRig(1)
	text, isErr := callTool(t, r.t, r.j, "remove_objects", map[string]any{"ids": []string{"nope"}})
	if !isErr || !strings.Contains(text, "Nenhum objeto") || len(r.j.List("c1")) != 0 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}

func TestSetDoorOnlyDoors(t *testing.T) {
	r := liveRig(1)
	r.sc.scenes["s1"].Objects = []maps.Object{
		{ID: "d", Prefab: "door", X: 1, Z: 1},
		{ID: "w", Prefab: "wall", X: 2, Z: 2},
	}
	text, isErr := callTool(t, r.t, r.j, "set_door", map[string]any{"objectId": "w", "open": true})
	if !isErr || !strings.Contains(text, "porta") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	text, isErr = callTool(t, r.t, r.j, "set_door", map[string]any{"objectId": "d", "open": true})
	if isErr || !strings.Contains(text, "aberta") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if up := r.sc.objects[0].Update; len(up) != 1 || up[0].Props["open"] != true {
		t.Fatalf("update %+v", up)
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	if up := r.sc.objects[1].Update; len(up) != 1 || up[0].Props["open"] == true {
		t.Fatalf("inverse %+v", up)
	}
}
