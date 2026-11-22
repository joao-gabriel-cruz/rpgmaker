package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

type placeTokenInput struct {
	Character string `json:"character" jsonschema:"character id or name (NPC or player character)"`
	X         int    `json:"x"`
	Z         int    `json:"z"`
}

type removeTokenInput struct {
	Token string `json:"token" jsonschema:"token id or character name on the active scene"`
}

func registerNPCTokenTools(srv *sdk.Server, t Table, j *Journal, s Session) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "place_token",
		Description: "Put an existing character (NPC or player) on the active scene at a free, enterable cell; " +
			"a character already on the scene is refused. " + liveNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in placeTokenInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(placeToken(ctx, t, j, s, in))
	})
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "remove_token",
		Description: "Take a token off the active scene (the character is kept; undo puts the token back on the same cell). " + liveNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in removeTokenInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(removeToken(ctx, t, j, s, in))
	})
}

func placeToken(ctx context.Context, t Table, j *Journal, s Session, in placeTokenInput) (string, error) {
	list, err := t.Characters.List(ctx, gmMember(s))
	if err != nil {
		return "", fmt.Errorf("list characters: %w", err)
	}
	ch, err := findCharacter(list, in.Character)
	if err != nil {
		return "", err
	}
	sc, err := editableScene(ctx, t, s)
	if err != nil {
		return "", err
	}
	present, err := sceneTokens(ctx, t, s, sc.ID)
	if err != nil {
		return "", err
	}
	if err := checkFreeCell(sc, present, ch.ID, ch.Name, in); err != nil {
		return "", err
	}
	if err := t.Tokens.Place(ctx, s.Sender, s.CampaignID, sc.ID, ch.ID, in.X, in.Z); err != nil {
		return "", err
	}
	summary := fmt.Sprintf("%s colocado em (%d,%d)", ch.Name, in.X, in.Z)
	j.Record(s.CampaignID, sc.ID, summary, Inverse{RemoveCharacters: []string{ch.ID}})
	return summary, nil
}

func checkFreeCell(sc scenes.Scene, present []tokens.Token, charID, name string, in placeTokenInput) error {
	for _, tk := range present {
		if tk.CharacterID == charID {
			return newUserError(name + " já está na cena")
		}
	}
	if in.X < 0 || in.Z < 0 || in.X >= sc.Width || in.Z >= sc.Depth {
		return newUserError(fmt.Sprintf("a posição (%d,%d) fica fora da cena (%dx%d)", in.X, in.Z, sc.Width, sc.Depth))
	}
	for _, tk := range present {
		if tk.X == in.X && tk.Z == in.Z {
			return newUserError(fmt.Sprintf("a célula (%d,%d) está ocupada por %s", in.X, in.Z, tk.Name))
		}
	}
	g, err := maps.Deserialize(sc.Grid)
	if err != nil {
		return fmt.Errorf("decode grid: %w", err)
	}
	to := maps.Coord{X: in.X, Z: in.Z}
	if ok, reason := maps.CanEnter(g, sc.Objects, to, to, 0); !ok {
		return newUserError(reason)
	}
	return nil
}

func removeToken(ctx context.Context, t Table, j *Journal, s Session, in removeTokenInput) (string, error) {
	sc, err := editableScene(ctx, t, s)
	if err != nil {
		return "", err
	}
	list, err := sceneTokens(ctx, t, s, sc.ID)
	if err != nil {
		return "", err
	}
	tk, err := findToken(list, in.Token)
	if err != nil {
		return "", err
	}
	if err := t.Tokens.Remove(ctx, s.Sender, s.CampaignID, tk.ID); err != nil {
		return "", err
	}
	summary := tk.Name + " removido da cena"
	spot := TokenSpot{CharacterID: tk.CharacterID, X: tk.X, Z: tk.Z}
	j.Record(s.CampaignID, sc.ID, summary, Inverse{PlaceTokens: []TokenSpot{spot}})
	return summary, nil
}
