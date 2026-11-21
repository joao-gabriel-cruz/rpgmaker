package mcp

import (
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

// liveRig returns a rig with an active 8x8 scene "s1" of floor cells at the given height.
func liveRig(height int) *rig {
	r := newRig()
	g, _ := maps.NewGrid(8, 8)
	for i := range g.Cells {
		g.Cells[i].Height = height
	}
	r.sc.scenes["s1"] = &scenes.Scene{ID: "s1", Name: "Praça", Width: 8, Depth: 8, Grid: g.Serialize(), IsActive: true}
	return r
}

func TestPaintLavaCircleIsLiveAndUndoable(t *testing.T) {
	r := liveRig(1)
	text, isErr := callTool(t, r.t, r.j, "paint_terrain", map[string]any{
		"shape":   map[string]any{"kind": "circle", "from": map[string]any{"x": 4, "z": 4}, "radius": 1},
		"terrain": "lava",
	})
	if isErr || !strings.Contains(text, "5 células → Lava") || !strings.Contains(text, "(4,4)") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
	if len(r.sc.painted) != 1 || len(r.sc.painted[0].Cells) != 5 || r.sc.origins[0] != "" || r.sc.hit[0] != "s1" {
		t.Fatalf("painted=%+v origins=%v hit=%v", r.sc.painted, r.sc.origins, r.sc.hit)
	}
	if got := r.j.List("c1"); len(got) != 1 || got[0].SceneID != "s1" {
		t.Fatalf("journal = %+v", got)
	}
	if _, err := r.j.Undo(t.Context(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	back := r.sc.painted[1].Cells
	if len(back) != 5 || back[0].Terrain != "floor" || back[0].Height != 1 {
		t.Fatalf("inverse = %+v", back)
	}
}

func TestPaintHeightDeltaClamps(t *testing.T) {
	r := liveRig(9)
	_, isErr := callTool(t, r.t, r.j, "paint_terrain", map[string]any{
		"shape":       map[string]any{"kind": "cells", "cells": []map[string]int{{"x": 1, "z": 1}}},
		"heightDelta": 3,
	})
	if isErr {
		t.Fatal("unexpected error")
	}
	c := r.sc.painted[0].Cells[0]
	if c.Height != 10 || c.Terrain != "floor" {
		t.Fatalf("cell = %+v", c)
	}
}

func TestPaintRequiresSomething(t *testing.T) {
	r := liveRig(1)
	text, isErr := callTool(t, r.t, r.j, "paint_terrain", map[string]any{
		"shape": map[string]any{"kind": "cells", "cells": []map[string]int{{"x": 1, "z": 1}}},
	})
	if !isErr || !strings.Contains(text, "terrain") || len(r.sc.painted) != 0 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}

func TestPaintRejectsUnknownTerrain(t *testing.T) {
	r := liveRig(1)
	text, isErr := callTool(t, r.t, r.j, "paint_terrain", map[string]any{
		"shape":   map[string]any{"kind": "cells", "cells": []map[string]int{{"x": 1, "z": 1}}},
		"terrain": "mud",
	})
	if !isErr || !strings.Contains(text, "terreno") {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}

func TestPaintRefusesDraft(t *testing.T) {
	r := liveRig(1)
	r.sc.scenes["s1"].Draft = &scenes.Draft{}
	text, isErr := callTool(t, r.t, r.j, "paint_terrain", map[string]any{
		"shape":   map[string]any{"kind": "cells", "cells": []map[string]int{{"x": 1, "z": 1}}},
		"terrain": "lava",
	})
	if !isErr || !strings.Contains(text, "rascunho aberto") || len(r.j.List("c1")) != 0 {
		t.Fatalf("err=%v text=%q", isErr, text)
	}
}
