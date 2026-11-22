package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

func registerNPCTools(srv *sdk.Server, t Table, j *Journal, s Session) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "create_npc",
		Description: "Create one or more villain NPCs ready for combat (sheet fields you omit keep the NPC defaults: " +
			"instincts 1, aptitudes 0, 3 boxes per level). With 'at', also place their tokens on the free cells " +
			"nearest to that cell of the active scene; that placement is live and undoable (undo removes the tokens " +
			"but keeps the NPCs in the character list). " + coordsNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in createNPCInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(createNPC(ctx, t, j, s, in))
	})
	registerNPCTokenTools(srv, t, j, s)
}

func gmMember(s Session) campaigns.Member {
	return campaigns.Member{CampaignID: s.CampaignID, UserID: s.Sender.UserID, Role: campaigns.RoleGM}
}

func npcNames(name string, count int) []string {
	name = strings.TrimSpace(name)
	if count <= 1 {
		return []string{name}
	}
	out := make([]string, count)
	for i := range out {
		out[i] = fmt.Sprintf("%s %d", name, i+1)
	}
	return out
}

// npcSpot is the scene context needed to place new NPCs.
type npcSpot struct {
	sc    scenes.Scene
	grid  *maps.Grid
	taken map[maps.Coord]bool
	at    maps.Coord
}

func loadNPCSpot(ctx context.Context, t Table, s Session, at *Point) (*npcSpot, error) {
	if at == nil {
		return nil, nil
	}
	sc, err := editableScene(ctx, t, s)
	if err != nil {
		return nil, err
	}
	if at.X < 0 || at.Z < 0 || at.X >= sc.Width || at.Z >= sc.Depth {
		return nil, newUserError(fmt.Sprintf("a posição (%d,%d) fica fora da cena (%dx%d)", at.X, at.Z, sc.Width, sc.Depth))
	}
	g, err := maps.Deserialize(sc.Grid)
	if err != nil {
		return nil, fmt.Errorf("decode grid: %w", err)
	}
	list, err := sceneTokens(ctx, t, s, sc.ID)
	if err != nil {
		return nil, err
	}
	taken := map[maps.Coord]bool{}
	for _, tk := range list {
		taken[maps.Coord{X: tk.X, Z: tk.Z}] = true
	}
	return &npcSpot{sc: sc, grid: g, taken: taken, at: maps.Coord{X: at.X, Z: at.Z}}, nil
}

func createNPC(ctx context.Context, t Table, j *Journal, s Session, in createNPCInput) (string, error) {
	if err := validateCreateNPC(in); err != nil {
		return "", err
	}
	spot, err := loadNPCSpot(ctx, t, s, in.At)
	if err != nil {
		return "", err
	}
	made, err := makeNPCs(ctx, t, s, in)
	if err != nil {
		return "", partialNPCError(made, err)
	}
	out := fmt.Sprintf("%d NPC(s) criado(s): %s", len(made), strings.Join(viewNames(made), ", "))
	if spot == nil {
		return out, nil
	}
	placed, failed := placeNPCs(ctx, t, j, s, spot, made)
	if len(placed) > 0 {
		out += fmt.Sprintf("; colocados perto de (%d,%d)", spot.at.X, spot.at.Z)
	}
	if len(failed) > 0 {
		out += "; sem espaço livre ou erro ao colocar: " + strings.Join(failed, ", ")
	}
	return out, nil
}

func viewNames(vs []characters.View) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.Name
	}
	return out
}

func partialNPCError(made []characters.View, err error) error {
	msg := toolError(err).Error()
	if len(made) == 0 {
		return newUserError(msg)
	}
	return newUserError(fmt.Sprintf("%s (já criados antes do erro: %s)", msg, strings.Join(viewNames(made), ", ")))
}

func makeNPCs(ctx context.Context, t Table, s Session, in createNPCInput) ([]characters.View, error) {
	m := gmMember(s)
	var made []characters.View
	for _, name := range npcNames(in.Name, in.Count) {
		v, err := t.Characters.Create(ctx, m, characters.CreateInput{Name: name, Kind: characters.KindNPC})
		if err != nil {
			return made, err
		}
		if hasSheetFields(in) {
			sheet, err := mergeNPCSheet(v.Sheet, in)
			if err != nil {
				return made, err
			}
			if v, err = saveNPC(ctx, t, m, v, sheet); err != nil {
				return made, err
			}
		}
		made = append(made, v)
	}
	return made, nil
}

func saveNPC(ctx context.Context, t Table, m campaigns.Member, v characters.View, sheet []byte) (characters.View, error) {
	saved, err := t.Characters.Save(ctx, m, v.ID, sheet)
	var rej *characters.RejectedError
	if errors.As(err, &rej) {
		msgs := make([]string, len(rej.Issues))
		for i, is := range rej.Issues {
			msgs[i] = is.Message
		}
		return v, newUserError("a ficha do NPC foi rejeitada: " + strings.Join(msgs, "; "))
	}
	if err != nil {
		return v, err
	}
	if saved.ID == "" {
		return v, nil
	}
	return saved, nil
}

// placeNPCs puts each NPC on the nearest free cell and journals the whole batch as one change.
func placeNPCs(ctx context.Context, t Table, j *Journal, s Session, sp *npcSpot, made []characters.View) (placed, failed []string) {
	for _, v := range made {
		c, ok := nearestFree(sp.grid, sp.sc.Objects, sp.taken, sp.at)
		if !ok || t.Tokens.Place(ctx, s.Sender, s.CampaignID, sp.sc.ID, v.ID, c.X, c.Z) != nil {
			failed = append(failed, v.Name)
			continue
		}
		sp.taken[c] = true
		placed = append(placed, v.ID)
	}
	if len(placed) == 0 {
		return nil, failed
	}
	ids := tokenIDsOf(ctx, t, s, sp.sc.ID, placed)
	summary := fmt.Sprintf("%d NPC(s) colocado(s) perto de (%d,%d)", len(placed), sp.at.X, sp.at.Z)
	j.Record(s.CampaignID, sp.sc.ID, summary, Inverse{RemoveTokens: ids})
	return placed, failed
}

// tokenIDsOf finds the tokens Place created, since Place does not return them.
func tokenIDsOf(ctx context.Context, t Table, s Session, sceneID string, charIDs []string) []string {
	list, err := sceneTokens(ctx, t, s, sceneID)
	if err != nil {
		return nil
	}
	var ids []string
	for _, tk := range list {
		for _, c := range charIDs {
			if tk.CharacterID == c {
				ids = append(ids, tk.ID)
			}
		}
	}
	return ids
}
