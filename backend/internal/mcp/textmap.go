package mcp

import (
	"fmt"
	"strings"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/ambience"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// TokenInfo is a token reduced to what Claude needs to place things around it.
type TokenInfo struct {
	ID, Name   string
	X, Z       int
	Player     bool
	Conditions []string
}

// TableView is everything RenderTable needs about the live scene.
type TableView struct {
	Scene    scenes.Scene
	Ambience ambience.Ambience
	Tokens   []TokenInfo
}

const axisLine = "x cresce para a direita (leste); z cresce para baixo (sul); norte = z menor"

var terrainGlyph = map[string]byte{
	"floor": '.', "grass": '"', "dirt": ':', "sand": 's', "water": '~',
	"stone": '#', "snow": '*', "lava": '!', "ruin": '%',
}

const legendLine = `Legenda: . piso, " grama, : terra, s areia, ~ água, # pedra, * neve, ! lava, % ruína`

// RenderTable draws the scene as compact PT-BR text.
func RenderTable(v TableView) (string, error) {
	g, err := maps.Deserialize(v.Scene.Grid)
	if err != nil {
		return "", fmt.Errorf("decode scene grid: %w", err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Cena: %s (%d×%d) — período: %s, clima: %s\n",
		v.Scene.Name, g.Width, g.Depth, v.Ambience.TimeOfDay, v.Ambience.Weather)
	b.WriteString(axisLine + "\n" + legendLine + "\n")
	writeGrid(&b, g, func(c maps.Cell) byte { return glyphOf(c.Terrain) })
	if variesHeight(g) {
		b.WriteString("Alturas (0-9, X = 10):\n")
		writeGrid(&b, g, func(c maps.Cell) byte { return heightGlyph(c.Height) })
	} else {
		fmt.Fprintf(&b, "altura uniforme: %d\n", g.Cells[0].Height)
	}
	writeTokens(&b, v.Tokens)
	writeObjects(&b, v.Scene.Objects)
	return b.String(), nil
}

func glyphOf(terrain string) byte {
	if c, ok := terrainGlyph[terrain]; ok {
		return c
	}
	return '?'
}

func heightGlyph(h int) byte {
	if h >= 10 {
		return 'X'
	}
	return byte('0' + max(h, 0))
}

func variesHeight(g *maps.Grid) bool {
	for _, c := range g.Cells {
		if c.Height != g.Cells[0].Height {
			return true
		}
	}
	return false
}

func writeGrid(b *strings.Builder, g *maps.Grid, glyph func(maps.Cell) byte) {
	for z := 0; z < g.Depth; z++ {
		fmt.Fprintf(b, "z%02d ", z)
		for x := 0; x < g.Width; x++ {
			b.WriteByte(glyph(g.Cells[z*g.Width+x]))
		}
		b.WriteByte('\n')
	}
}

func writeTokens(b *strings.Builder, tokens []TokenInfo) {
	b.WriteString("Tokens:\n")
	if len(tokens) == 0 {
		b.WriteString("(nenhum)\n")
	}
	for _, t := range tokens {
		kind := "NPC"
		if t.Player {
			kind = "jogador"
		}
		fmt.Fprintf(b, "- %s (%s) em (%d,%d)", t.Name, kind, t.X, t.Z)
		if len(t.Conditions) > 0 {
			fmt.Fprintf(b, " [%s]", strings.Join(t.Conditions, ", "))
		}
		b.WriteByte('\n')
	}
}

func writeObjects(b *strings.Builder, objects []maps.Object) {
	b.WriteString("Objetos:\n")
	if len(objects) == 0 {
		b.WriteString("(nenhum)\n")
	}
	for _, o := range objects {
		fmt.Fprintf(b, "- %s %s em (%d,%d) rot %d", o.ID, o.Prefab, o.X, o.Z, o.Rotation)
		if open, _ := o.Props["open"].(bool); open && o.Prefab == "door" {
			b.WriteString(" aberta")
		}
		b.WriteByte('\n')
	}
}

// RenderCatalog lists every id the write tools accept.
func RenderCatalog() string {
	var b strings.Builder
	b.WriteString("Terrenos (id glifo; bloqueia movimento/visão):\n")
	for _, t := range maps.TerrainOrder {
		fmt.Fprintf(&b, "- %s %c; movimento=%s, visão=%s\n",
			t.ID, glyphOf(t.ID), yesNo(t.BlocksMovement), yesNo(t.BlocksVision))
	}
	b.WriteString("Prefabs (id; área LxP; altura; bloqueia):\n")
	rules := maps.PrefabRules()
	for _, id := range maps.PrefabIDs() {
		r := rules[id]
		fmt.Fprintf(&b, "- %s; %dx%d; altura %d; movimento=%s, visão=%s\n",
			id, r.W, r.D, r.Height, yesNo(r.BlocksMovement), yesNo(r.BlocksVision))
	}
	fmt.Fprintf(&b, "Períodos: %s\n", strings.Join(ambience.TimesOfDay(), ", "))
	fmt.Fprintf(&b, "Climas: %s\n", strings.Join(ambience.Weathers(), ", "))
	fmt.Fprintf(&b, "Condições: %s\n", strings.Join(assimilacao.Conditions, ", "))
	return b.String()
}

func yesNo(v bool) string {
	if v {
		return "sim"
	}
	return "não"
}
