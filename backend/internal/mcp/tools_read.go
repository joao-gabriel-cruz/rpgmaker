package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const coordsNote = "Coordinates: x grows east (right), z grows south (down). " +
	"Changes made through other tools are live for the players, and every write can be undone by the Mestre."

type noArgs struct{}

func registerReadTools(srv *sdk.Server, t Table, _ *Journal, s Session) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "get_table",
		Description: "Show the active scene as text: terrain grid, heights, tokens (with conditions), objects, " +
			"time of day and weather. " + coordsNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ noArgs) (*sdk.CallToolResult, any, error) {
		return getTable(ctx, t, s)
	})
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "list_scenes",
		Description: "List the campaign's scenes (id · name · WxD · active). " + coordsNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ noArgs) (*sdk.CallToolResult, any, error) {
		return listScenes(ctx, t, s)
	})
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "get_catalog",
		Description: "List every terrain, prefab, time of day, weather and condition id the other tools accept.",
	}, func(context.Context, *sdk.CallToolRequest, noArgs) (*sdk.CallToolResult, any, error) {
		return textResult(RenderCatalog()), nil, nil
	})
}

func getTable(ctx context.Context, t Table, s Session) (*sdk.CallToolResult, any, error) {
	sc, err := activeScene(ctx, t, s)
	if errors.Is(err, errNoActiveScene) {
		return textResult("Nenhuma cena está ativa. Use list_scenes para ver as cenas disponíveis."), nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	v, err := tableView(ctx, t, s, sc)
	if err != nil {
		return nil, nil, err
	}
	text, err := RenderTable(v)
	if err != nil {
		return nil, nil, err
	}
	return textResult(text), nil, nil
}

func listScenes(ctx context.Context, t Table, s Session) (*sdk.CallToolResult, any, error) {
	list, err := t.Scenes.List(ctx, s.CampaignID)
	if err != nil {
		return nil, nil, err
	}
	if len(list) == 0 {
		return textResult("Esta campanha ainda não tem cenas."), nil, nil
	}
	var b strings.Builder
	for _, sc := range list {
		fmt.Fprintf(&b, "%s · %s · %dx%d", sc.ID, sc.Name, sc.Width, sc.Depth)
		if sc.IsActive {
			b.WriteString(" · ativa")
		}
		b.WriteByte('\n')
	}
	return textResult(b.String()), nil, nil
}
