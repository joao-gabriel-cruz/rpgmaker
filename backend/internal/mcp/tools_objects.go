package mcp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

const maxPlacedObjects = 200

type objectIn struct {
	Prefab   string         `json:"prefab" jsonschema:"prefab id, see get_catalog"`
	X        int            `json:"x"`
	Z        int            `json:"z"`
	Rotation int            `json:"rotation,omitempty" jsonschema:"0, 90, 180 or 270 degrees"`
	Props    map[string]any `json:"props,omitempty"`
}

type objectLine struct {
	Prefab string `json:"prefab" jsonschema:"a 1x1 prefab id"`
	From   Point  `json:"from"`
	To     Point  `json:"to"`
}

type placeInput struct {
	Objects []objectIn  `json:"objects,omitempty"`
	Line    *objectLine `json:"line,omitempty" jsonschema:"one object per cell along the line (1x1 prefabs only)"`
}

type removeInput struct {
	IDs  []string `json:"ids,omitempty"`
	Area *Shape   `json:"area,omitempty" jsonschema:"remove every object whose footprint touches this area"`
}

type doorInput struct {
	ObjectID string `json:"objectId"`
	Open     bool   `json:"open"`
}

const liveNote = "Live for the players and undoable by the Mestre. " + coordsNote

func registerObjectTools(srv *sdk.Server, t Table, j *Journal, s Session) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "place_objects",
		Description: "Place objects (walls, trees, crates, doors...) on the active scene, individually and/or along a line " +
			"(max 200 per call). Blocking prefabs cannot be placed over a token. " + liveNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in placeInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(placeObjects(ctx, t, j, s, in))
	})
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "remove_objects",
		Description: "Remove objects of the active scene by id and/or by area. " + liveNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in removeInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(removeObjects(ctx, t, j, s, in))
	})
	sdk.AddTool(srv, &sdk.Tool{
		Name:        "set_door",
		Description: "Open or close a door object of the active scene. " + liveNote,
	}, func(ctx context.Context, _ *sdk.CallToolRequest, in doorInput) (*sdk.CallToolResult, any, error) {
		return summaryResult(setDoor(ctx, t, j, s, in))
	})
}

func summaryResult(text string, err error) (*sdk.CallToolResult, any, error) {
	if err != nil {
		return nil, nil, toolError(err)
	}
	return textResult(text), nil, nil
}

// editableScene loads the active scene and refuses when a draft is open.
func editableScene(ctx context.Context, t Table, s Session) (scenes.Scene, error) {
	sc, err := activeScene(ctx, t, s)
	if err != nil {
		return sc, err
	}
	return sc, requireNoDraft(sc)
}

func newObjectID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "mcp-" + hex.EncodeToString(b)
}

func lineObjects(l objectLine) ([]objectIn, error) {
	if r, ok := maps.PrefabRules()[l.Prefab]; ok && (r.W != 1 || r.D != 1) {
		return nil, newUserError(fmt.Sprintf("a linha só aceita prefabs 1x1; %q ocupa %dx%d", l.Prefab, r.W, r.D))
	}
	if err := checkPoint(&l.From, "from"); err != nil {
		return nil, err
	}
	if err := checkPoint(&l.To, "to"); err != nil {
		return nil, err
	}
	var out []objectIn
	for _, p := range bresenham(l.From, l.To) {
		out = append(out, objectIn{Prefab: l.Prefab, X: p.X, Z: p.Z})
	}
	return out, nil
}

func toMapObjects(in placeInput) ([]maps.Object, error) {
	list := append([]objectIn{}, in.Objects...)
	if in.Line != nil {
		cells, err := lineObjects(*in.Line)
		if err != nil {
			return nil, err
		}
		list = append(list, cells...)
	}
	if len(list) == 0 {
		return nil, newUserError("informe objects ou line")
	}
	if len(list) > maxPlacedObjects {
		return nil, newUserError("no máximo 200 objetos por chamada")
	}
	out := make([]maps.Object, 0, len(list))
	for _, o := range list {
		if !maps.KnownPrefab(o.Prefab) {
			return nil, newUserError(fmt.Sprintf("objeto desconhecido %q; veja get_catalog", o.Prefab))
		}
		if o.Rotation%90 != 0 || o.Rotation < 0 || o.Rotation > 270 {
			return nil, newUserError("a rotação deve ser 0, 90, 180 ou 270")
		}
		out = append(out, maps.Object{ID: newObjectID(), Prefab: o.Prefab, X: o.X, Z: o.Z, Rotation: o.Rotation, Props: o.Props})
	}
	return out, nil
}

// checkTokenCells keeps blocking objects off cells where a token stands.
func checkTokenCells(ctx context.Context, t Table, s Session, sceneID string, objs []maps.Object) error {
	tks, err := sceneTokens(ctx, t, s, sceneID)
	if err != nil {
		return err
	}
	at := map[maps.Coord]string{}
	for _, tk := range tks {
		at[maps.Coord{X: tk.X, Z: tk.Z}] = tk.Name
	}
	for _, o := range objs {
		if !maps.RuleFor(o).BlocksMovement {
			continue
		}
		for _, c := range maps.Footprint(o) {
			if name, ok := at[c]; ok {
				return newUserError(fmt.Sprintf("%s está em (%d,%d); %s bloquearia a célula do token", name, c.X, c.Z, o.Prefab))
			}
		}
	}
	return nil
}

func placeObjects(ctx context.Context, t Table, j *Journal, s Session, in placeInput) (string, error) {
	objs, err := toMapObjects(in)
	if err != nil {
		return "", err
	}
	sc, err := editableScene(ctx, t, s)
	if err != nil {
		return "", err
	}
	if err := checkTokenCells(ctx, t, s, sc.ID, objs); err != nil {
		return "", err
	}
	if err := t.Scenes.EditObjects(ctx, s.CampaignID, sc.ID, "", scenes.ObjectPatch{Add: objs}); err != nil {
		return "", err
	}
	ids := make([]string, len(objs))
	for i, o := range objs {
		ids[i] = o.ID
	}
	summary := fmt.Sprintf("%d objeto(s) colocado(s) perto de (%d,%d)", len(objs), objs[0].X, objs[0].Z)
	j.Record(s.CampaignID, sc.ID, summary, Inverse{Objects: &scenes.ObjectPatch{Remove: ids}})
	return summary, nil
}

func removeObjects(ctx context.Context, t Table, j *Journal, s Session, in removeInput) (string, error) {
	sc, err := editableScene(ctx, t, s)
	if err != nil {
		return "", err
	}
	found, err := matchObjects(sc, in)
	if err != nil {
		return "", err
	}
	if len(found) == 0 {
		return "", newUserError("Nenhum objeto encontrado para remover")
	}
	ids := make([]string, len(found))
	for i, o := range found {
		ids[i] = o.ID
	}
	if err := t.Scenes.EditObjects(ctx, s.CampaignID, sc.ID, "", scenes.ObjectPatch{Remove: ids}); err != nil {
		return "", err
	}
	summary := fmt.Sprintf("%d objeto(s) removido(s)", len(found))
	j.Record(s.CampaignID, sc.ID, summary, Inverse{Objects: &scenes.ObjectPatch{Add: found}})
	return summary, nil
}

func matchObjects(sc scenes.Scene, in removeInput) ([]maps.Object, error) {
	want := map[string]bool{}
	for _, id := range in.IDs {
		want[id] = true
	}
	area := map[maps.Coord]bool{}
	if in.Area != nil {
		cells, err := Rasterize(*in.Area, sc.Width, sc.Depth)
		if err != nil {
			return nil, err
		}
		for _, c := range cells {
			area[c] = true
		}
	} else if len(want) == 0 {
		return nil, newUserError("informe ids ou area")
	}
	var found []maps.Object
	for _, o := range sc.Objects {
		if want[o.ID] || touches(o, area) {
			found = append(found, o)
		}
	}
	return found, nil
}

func touches(o maps.Object, area map[maps.Coord]bool) bool {
	for _, c := range maps.Footprint(o) {
		if area[c] {
			return true
		}
	}
	return false
}

func setDoor(ctx context.Context, t Table, j *Journal, s Session, in doorInput) (string, error) {
	sc, err := editableScene(ctx, t, s)
	if err != nil {
		return "", err
	}
	o, err := findDoor(sc, in.ObjectID)
	if err != nil {
		return "", err
	}
	next := o
	next.Props = map[string]any{}
	for k, v := range o.Props {
		next.Props[k] = v
	}
	next.Props["open"] = in.Open
	if err := t.Scenes.EditObjects(ctx, s.CampaignID, sc.ID, "", scenes.ObjectPatch{Update: []maps.Object{next}}); err != nil {
		return "", err
	}
	state := map[bool]string{true: "aberta", false: "fechada"}[in.Open]
	summary := fmt.Sprintf("porta em (%d,%d) %s", o.X, o.Z, state)
	j.Record(s.CampaignID, sc.ID, summary, Inverse{Objects: &scenes.ObjectPatch{Update: []maps.Object{o}}})
	return summary, nil
}

func findDoor(sc scenes.Scene, id string) (maps.Object, error) {
	for _, o := range sc.Objects {
		if o.ID != id {
			continue
		}
		if o.Prefab != "door" {
			return o, newUserError("esse objeto não é uma porta")
		}
		return o, nil
	}
	return maps.Object{}, newUserError("objeto não encontrado")
}
