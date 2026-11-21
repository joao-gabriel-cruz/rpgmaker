package mcp

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

const (
	maxSceneNameLen = 80
	minSceneSide    = 16
	maxSceneSide    = 64
)

type createSceneInput struct {
	Name        string `json:"name" jsonschema:"scene name, up to 80 characters"`
	Width       int    `json:"width" jsonschema:"cells along x, 16-64"`
	Depth       int    `json:"depth" jsonschema:"cells along z, 16-64"`
	BaseTerrain string `json:"baseTerrain,omitempty" jsonschema:"terrain id for every cell, default floor"`
}

type activateSceneInput struct {
	Scene   string `json:"scene" jsonschema:"scene id or name"`
	Arrival *Point `json:"arrival,omitempty" jsonschema:"where the party arrives; default is the scene center"`
}

func registerSceneTools(srv *sdk.Server, t Table, _ *Journal, s Session) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "create_scene",
		Description: "Create a new scene. It stays inactive and invisible to the players until activate_scene. " +
			"Not undoable. " + coordsNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in createSceneInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(createScene(ctx, t, s, in))
	})
	sdk.AddTool(srv, &sdk.Tool{
		Name: "activate_scene",
		Description: "Make a scene the active one (players see it at once) and bring every player character that " +
			"has no token there to the free cells nearest to the arrival point. Not undoable. " + coordsNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in activateSceneInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(activateScene(ctx, t, s, in))
	})
}

func validateCreateScene(in createSceneInput) error {
	if n := utf8.RuneCountInString(strings.TrimSpace(in.Name)); n == 0 || n > maxSceneNameLen {
		return newUserError("o nome da cena deve ter entre 1 e 80 caracteres")
	}
	for _, side := range []int{in.Width, in.Depth} {
		if side < minSceneSide || side > maxSceneSide {
			return newUserError("largura e profundidade devem ficar entre 16 e 64")
		}
	}
	if in.BaseTerrain != "" {
		if _, ok := maps.TerrainByID(in.BaseTerrain); !ok {
			return newUserError(fmt.Sprintf("terreno desconhecido %q; veja get_catalog", in.BaseTerrain))
		}
	}
	return nil
}

func createScene(ctx context.Context, t Table, s Session, in createSceneInput) (string, error) {
	if err := validateCreateScene(in); err != nil {
		return "", err
	}
	sc, err := t.Scenes.Create(ctx, s.CampaignID, strings.TrimSpace(in.Name), in.Width, in.Depth)
	if err != nil {
		return "", err
	}
	if in.BaseTerrain != "" && in.BaseTerrain != "floor" {
		if err := paintBase(ctx, t, s, sc, in.BaseTerrain); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("cena %q criada (id %s, %dx%d); continua inativa, os jogadores não a veem",
		sc.Name, sc.ID, sc.Width, sc.Depth), nil
}

// paintBase keeps each cell's current height and only swaps the terrain.
func paintBase(ctx context.Context, t Table, s Session, sc scenes.Scene, terrain string) error {
	g, err := maps.Deserialize(sc.Grid)
	if err != nil {
		return fmt.Errorf("decode grid: %w", err)
	}
	cells := make([]maps.CellPatch, 0, len(g.Cells))
	for i, c := range g.Cells {
		cells = append(cells, maps.CellPatch{X: i % g.Width, Z: i / g.Width, Terrain: terrain, Height: c.Height})
	}
	return t.Scenes.Paint(ctx, s.CampaignID, sc.ID, "", maps.Patch{Cells: cells})
}

func activateScene(ctx context.Context, t Table, s Session, in activateSceneInput) (string, error) {
	list, err := t.Scenes.List(ctx, s.CampaignID)
	if err != nil {
		return "", fmt.Errorf("list scenes: %w", err)
	}
	sum, err := pick(list, in.Scene, func(x scenes.Summary) string { return x.ID }, func(x scenes.Summary) string { return x.Name })
	if err != nil {
		return "", err
	}
	sc, err := t.Scenes.Get(ctx, s.CampaignID, sum.ID, campaigns.RoleGM)
	if err != nil {
		return "", err
	}
	arrival := maps.Coord{X: sc.Width / 2, Z: sc.Depth / 2}
	if in.Arrival != nil {
		arrival = maps.Coord{X: in.Arrival.X, Z: in.Arrival.Z}
	}
	if arrival.X < 0 || arrival.Z < 0 || arrival.X >= sc.Width || arrival.Z >= sc.Depth {
		return "", newUserError(fmt.Sprintf("a chegada (%d,%d) fica fora da cena (%dx%d)", arrival.X, arrival.Z, sc.Width, sc.Depth))
	}
	if err := t.Scenes.Activate(ctx, s.CampaignID, sc.ID); err != nil {
		return "", err
	}
	brought, failed, err := bringParty(ctx, t, s, sc, arrival)
	if err != nil {
		return "", err
	}
	return activateSummary(sc.Name, brought, failed), nil
}

func activateSummary(name string, brought, failed []string) string {
	out := fmt.Sprintf("cena %q ativada", name)
	if len(brought) > 0 {
		out += "; trouxe: " + strings.Join(brought, ", ")
	}
	if len(failed) > 0 {
		out += "; não colocado (sem espaço livre ou erro): " + strings.Join(failed, ", ")
	}
	return out
}

// bringParty places each player character lacking a token on the free cell nearest to arrival.
func bringParty(ctx context.Context, t Table, s Session, sc scenes.Scene, arrival maps.Coord) (brought, failed []string, err error) {
	m := campaigns.Member{CampaignID: s.CampaignID, UserID: s.Sender.UserID, Role: campaigns.RoleGM}
	chars, err := t.Characters.List(ctx, m)
	if err != nil {
		return nil, nil, fmt.Errorf("list characters: %w", err)
	}
	present, err := sceneTokens(ctx, t, s, sc.ID)
	if err != nil {
		return nil, nil, err
	}
	g, err := maps.Deserialize(sc.Grid)
	if err != nil {
		return nil, nil, fmt.Errorf("decode grid: %w", err)
	}
	taken := map[maps.Coord]bool{}
	has := map[string]bool{}
	for _, tk := range present {
		taken[maps.Coord{X: tk.X, Z: tk.Z}] = true
		has[tk.CharacterID] = true
	}
	for _, ch := range chars {
		if ch.OwnerID == nil || has[ch.ID] {
			continue
		}
		at, ok := nearestFree(g, sc.Objects, taken, arrival)
		if !ok || t.Tokens.Place(ctx, s.Sender, s.CampaignID, sc.ID, ch.ID, at.X, at.Z) != nil {
			failed = append(failed, ch.Name)
			continue
		}
		taken[at] = true
		brought = append(brought, fmt.Sprintf("%s em (%d,%d)", ch.Name, at.X, at.Z))
	}
	return brought, failed, nil
}

// nearestFree walks Chebyshev rings outward from arrival, row-major inside each ring.
func nearestFree(g *maps.Grid, objs []maps.Object, taken map[maps.Coord]bool, arrival maps.Coord) (maps.Coord, bool) {
	for r := 0; r <= max(g.Width, g.Depth); r++ {
		for z := arrival.Z - r; z <= arrival.Z+r; z++ {
			for x := arrival.X - r; x <= arrival.X+r; x++ {
				c := maps.Coord{X: x, Z: z}
				if max(abs(x-arrival.X), abs(z-arrival.Z)) != r || !g.InBounds(x, z) || taken[c] {
					continue
				}
				if ok, _ := maps.CanEnter(g, objs, c, c, 0); ok {
					return c, true
				}
			}
		}
	}
	return maps.Coord{}, false
}
