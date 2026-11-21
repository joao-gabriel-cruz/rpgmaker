package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

var terrainNamesPT = map[string]string{
	"floor": "Chão", "grass": "Grama", "dirt": "Terra", "sand": "Areia", "water": "Água",
	"stone": "Pedra", "snow": "Neve", "lava": "Lava", "ruin": "Ruína",
}

type paintInput struct {
	Shape       Shape  `json:"shape" jsonschema:"area to paint: rect, circle, line or cells"`
	Terrain     string `json:"terrain,omitempty" jsonschema:"terrain id, see get_catalog"`
	Height      *int   `json:"height,omitempty" jsonschema:"absolute height 0-10; do not combine with heightDelta"`
	HeightDelta *int   `json:"heightDelta,omitempty" jsonschema:"raise (+) or lower (-) the current height, result clamped to 0-10; do not combine with height"`
}

func registerTerrainTools(srv *sdk.Server, t Table, j *Journal, s Session) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "paint_terrain",
		Description: "Paint terrain and/or heights over an area of the active scene. Give at least one of terrain, " +
			"height or heightDelta (heights are clamped to 0-10). Live for the players and undoable. " + coordsNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in paintInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(paintTerrain(ctx, t, j, s, in))
	})
}

func validatePaint(in paintInput) error {
	if in.Terrain == "" && in.Height == nil && in.HeightDelta == nil {
		return newUserError("informe terrain, height ou heightDelta")
	}
	if in.Height != nil && in.HeightDelta != nil {
		return newUserError("use height ou heightDelta, não os dois")
	}
	if in.Terrain != "" {
		if _, ok := maps.TerrainByID(in.Terrain); !ok {
			return newUserError(fmt.Sprintf("terreno desconhecido %q; veja get_catalog", in.Terrain))
		}
	}
	if in.Height != nil && (*in.Height < maps.MinHeight || *in.Height > maps.MaxHeight) {
		return newUserError("a altura deve ficar entre 0 e 10")
	}
	return nil
}

func paintTerrain(ctx context.Context, t Table, j *Journal, s Session, in paintInput) (string, error) {
	if err := validatePaint(in); err != nil {
		return "", err
	}
	sc, err := editableScene(ctx, t, s)
	if err != nil {
		return "", err
	}
	g, err := maps.Deserialize(sc.Grid)
	if err != nil {
		return "", err
	}
	coords, err := Rasterize(in.Shape, g.Width, g.Depth)
	if err != nil {
		return "", err
	}
	next, prev := paintPatches(g, coords, in)
	if err := t.Scenes.Paint(ctx, s.CampaignID, sc.ID, "", maps.Patch{Cells: next}); err != nil {
		return "", err
	}
	summary := paintSummary(coords, in)
	j.Record(s.CampaignID, sc.ID, summary, Inverse{Cells: prev})
	return summary, nil
}

func paintPatches(g *maps.Grid, coords []maps.Coord, in paintInput) (next, prev []maps.CellPatch) {
	for _, c := range coords {
		cur, _ := g.At(c.X, c.Z)
		n := maps.CellPatch{X: c.X, Z: c.Z, Terrain: cur.Terrain, Height: cur.Height}
		if in.Terrain != "" {
			n.Terrain = in.Terrain
		}
		if in.Height != nil {
			n.Height = *in.Height
		}
		if in.HeightDelta != nil {
			n.Height = cur.Height + *in.HeightDelta
		}
		n.Height = min(maps.MaxHeight, max(maps.MinHeight, n.Height))
		next = append(next, n)
		prev = append(prev, maps.CellPatch{X: c.X, Z: c.Z, Terrain: cur.Terrain, Height: cur.Height})
	}
	return next, prev
}

func paintSummary(coords []maps.Coord, in paintInput) string {
	what := ""
	if in.Terrain != "" {
		what = terrainNamesPT[in.Terrain]
	}
	switch {
	case in.Height != nil:
		what += fmt.Sprintf("%saltura %d", sep(what), *in.Height)
	case in.HeightDelta != nil:
		what += fmt.Sprintf("%saltura %+d", sep(what), *in.HeightDelta)
	}
	x, z := centroid(coords)
	return fmt.Sprintf("%d células → %s perto de (%d,%d)", len(coords), what, x, z)
}

func sep(s string) string {
	if s == "" {
		return ""
	}
	return ", "
}

func centroid(coords []maps.Coord) (int, int) {
	var sx, sz int
	for _, c := range coords {
		sx += c.X
		sz += c.Z
	}
	return sx / len(coords), sz / len(coords)
}
