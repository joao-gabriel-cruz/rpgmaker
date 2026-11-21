package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/ambience"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

type moveInput struct {
	Token string `json:"token" jsonschema:"token id or character name on the active scene"`
	X     int    `json:"x"`
	Z     int    `json:"z"`
}

type ambienceInput struct {
	TimeOfDay string `json:"timeOfDay,omitempty" jsonschema:"dawn, day, dusk or night"`
	Weather   string `json:"weather,omitempty" jsonschema:"clear, rain, storm, fog, snow or ash"`
}

type conditionInput struct {
	Character string `json:"character" jsonschema:"character id or name"`
	Condition string `json:"condition" jsonschema:"wounded, debilitated, susceptible, scared, hidden, infected or unconscious"`
	Active    bool   `json:"active"`
}

var ptWords = map[string]string{
	"dawn": "amanhecer", "day": "dia", "dusk": "entardecer", "night": "noite",
	"clear": "limpo", "rain": "chuva", "storm": "tempestade", "fog": "neblina", "snow": "neve", "ash": "cinzas",
}

// ptConditions mirrors the labels in frontend/src/features/ambience/conditions.ts.
var ptConditions = map[string]string{
	"wounded": "Ferido", "debilitated": "Debilitado", "susceptible": "Suscetível", "scared": "Assustado",
	"hidden": "Escondido", "infected": "Infectado", "unconscious": "Inconsciente",
}

type undoInput struct {
	ChangeID string `json:"changeId,omitempty" jsonschema:"id of a change from get_scene/recent changes; omit to undo the newest"`
}

func registerTokenTools(srv *sdk.Server, t Table, j *Journal, s Session) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "move_token",
		Description: "Move a character's token on the active scene, ignoring turn and distance limits " +
			"(the target cell must be enterable). " + liveNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in moveInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(moveToken(ctx, t, j, s, in))
	})
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "set_ambience",
		Description: "Set the active scene's time of day and/or weather; omitted fields keep their value. " + liveNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in ambienceInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(setAmbience(ctx, t, j, s, in))
	})
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "set_condition",
		Description: "Turn a condition on or off for a character (shown on the sheet and token). " + liveNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in conditionInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(setCondition(ctx, t, j, s, in))
	})
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "undo_last",
		Description: "Undo the most recent change made through these tools.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in undoInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(undoLast(ctx, t, j, s, in.ChangeID))
	})
}

func undoLast(ctx context.Context, t Table, j *Journal, s Session, changeID string) (string, error) {
	summary, err := j.Undo(ctx, t, s, changeID)
	if err != nil {
		return "", err
	}
	return "desfeito: " + summary, nil
}

func moveToken(ctx context.Context, t Table, j *Journal, s Session, in moveInput) (string, error) {
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
	g, err := maps.Deserialize(sc.Grid)
	if err != nil {
		return "", fmt.Errorf("decode grid: %w", err)
	}
	to := maps.Coord{X: in.X, Z: in.Z}
	if ok, reason := maps.CanEnter(g, sc.Objects, to, to, 0); !ok {
		return "", newUserError(reason)
	}
	if err := t.Tokens.Move(ctx, s.Sender, s.CampaignID, tk.ID, in.X, in.Z, true); err != nil {
		return "", err
	}
	summary := fmt.Sprintf("%s moveu para (%d,%d)", tk.Name, in.X, in.Z)
	j.Record(s.CampaignID, sc.ID, summary, Inverse{Token: &TokenPos{TokenID: tk.ID, X: tk.X, Z: tk.Z}})
	return summary, nil
}

func validateAmbience(in ambienceInput) error {
	if in.TimeOfDay == "" && in.Weather == "" {
		return newUserError("informe timeOfDay e/ou weather")
	}
	if in.TimeOfDay != "" && !slices.Contains(ambience.TimesOfDay(), in.TimeOfDay) {
		return newUserError(fmt.Sprintf("hora do dia inválida %q; use %v", in.TimeOfDay, ambience.TimesOfDay()))
	}
	if in.Weather != "" && !slices.Contains(ambience.Weathers(), in.Weather) {
		return newUserError(fmt.Sprintf("clima inválido %q; use %v", in.Weather, ambience.Weathers()))
	}
	return nil
}

func setAmbience(ctx context.Context, t Table, j *Journal, s Session, in ambienceInput) (string, error) {
	if err := validateAmbience(in); err != nil {
		return "", err
	}
	sc, err := editableScene(ctx, t, s)
	if err != nil {
		return "", err
	}
	prev, err := t.Ambience.Get(ctx, s.CampaignID, sc.ID)
	if err != nil {
		return "", fmt.Errorf("get ambience: %w", err)
	}
	next := prev
	next.SceneID = sc.ID
	if in.TimeOfDay != "" {
		next.TimeOfDay = in.TimeOfDay
	}
	if in.Weather != "" {
		next.Weather = in.Weather
	}
	if err := t.Ambience.Set(ctx, s.CampaignID, next); err != nil {
		return "", err
	}
	summary := fmt.Sprintf("ambiente: %s, %s", ptWords[next.TimeOfDay], ptWords[next.Weather])
	prev.SceneID = sc.ID
	j.Record(s.CampaignID, sc.ID, summary, Inverse{Ambience: &prev})
	return summary, nil
}

func hasCondition(c characters.View, cond string) bool {
	var sheet struct {
		Conditions []string `json:"conditions"`
	}
	_ = json.Unmarshal(c.Sheet, &sheet)
	return slices.Contains(sheet.Conditions, cond)
}

func setCondition(ctx context.Context, t Table, j *Journal, s Session, in conditionInput) (string, error) {
	if !slices.Contains(assimilacao.Conditions, in.Condition) {
		return "", newUserError(fmt.Sprintf("condição inválida %q; use %v", in.Condition, assimilacao.Conditions))
	}
	sc, err := editableScene(ctx, t, s)
	if err != nil {
		return "", err
	}
	m := campaigns.Member{CampaignID: s.CampaignID, UserID: s.Sender.UserID, Role: campaigns.RoleGM}
	list, err := t.Characters.List(ctx, m)
	if err != nil {
		return "", fmt.Errorf("list characters: %w", err)
	}
	ch, err := findCharacter(list, in.Character)
	if err != nil {
		return "", err
	}
	was := hasCondition(ch, in.Condition)
	if was == in.Active {
		return "já estava assim", nil
	}
	if err := t.Ambience.SetCondition(ctx, s.Sender, s.CampaignID, ch.ID, in.Condition, in.Active); err != nil {
		return "", err
	}
	state := map[bool]string{true: "ativada", false: "removida"}[in.Active]
	summary := fmt.Sprintf("%s: condição %s %s", ch.Name, ptConditions[in.Condition], state)
	j.Record(s.CampaignID, sc.ID, summary, Inverse{Condition: &ConditionState{CharacterID: ch.ID, Condition: in.Condition, Active: was}})
	return summary, nil
}
