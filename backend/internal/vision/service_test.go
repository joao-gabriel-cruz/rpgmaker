package vision

import (
	"context"
	"sync"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

const (
	camp    = "00000000-0000-0000-0000-00000000000c"
	sceneID = "00000000-0000-0000-0000-000000000001"
)

var (
	gm   = realtime.Sender{UserID: "gm", Name: "Mestre", Role: campaigns.RoleGM}
	ana  = realtime.Sender{UserID: "ana", Name: "Ana", Role: campaigns.RolePlayer}
	beto = realtime.Sender{UserID: "beto", Name: "Beto", Role: campaigns.RolePlayer}
)

type sent struct {
	to      realtime.Filter
	typ     string
	payload any
}

type fakeHub struct {
	mu     sync.Mutex
	events []sent
	online []realtime.Sender
}

func (h *fakeHub) Broadcast(_ string, to realtime.Filter, typ string, payload any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, sent{to, typ, payload})
}

func (h *fakeHub) Online(string) []realtime.Sender { return h.online }

// got returns the events of a type that the sender would receive.
func (h *fakeHub) got(s realtime.Sender, typ string) []any {
	var out []any
	for _, e := range h.events {
		if e.typ == typ && e.to(s) {
			out = append(out, e.payload)
		}
	}
	return out
}

func (h *fakeHub) reset() { h.events = nil }

type fakeRepo struct {
	fog Fog
	pos []Pos
}

func (f *fakeRepo) Fog(context.Context, string) (Fog, error) { return f.fog, nil }
func (f *fakeRepo) SaveConfig(_ context.Context, _, mode string, r int) error {
	f.fog.Mode, f.fog.Radius = mode, r
	return nil
}
func (f *fakeRepo) SaveRevealed(_ context.Context, _, rev string) error {
	f.fog.Revealed = rev
	return nil
}
func (f *fakeRepo) Positions(context.Context, string) ([]Pos, error) { return f.pos, nil }

type fakeScenes struct{ sc scenes.Scene }

func (f *fakeScenes) Get(_ context.Context, _, id, _ string) (scenes.Scene, error) {
	if id != f.sc.ID {
		return scenes.Scene{}, scenes.ErrNotFound
	}
	return f.sc, nil
}

type rig struct {
	svc  *Service
	repo *fakeRepo
	hub  *fakeHub
	sc   *fakeScenes
}

// newRig builds a 32x32 scene with a full-height wall along x=10; Ana's token is at (5,16) and Beto's at (20,16).
func newRig(mode string) *rig {
	g, _ := maps.NewGrid(32, 32)
	var walls []maps.Object
	for z := 0; z < 32; z++ {
		walls = append(walls, maps.Object{ID: "w", Prefab: "wall", X: 10, Z: z})
	}
	sc := &fakeScenes{sc: scenes.Scene{ID: sceneID, CampaignID: camp, Width: 32, Depth: 32, Grid: g.Serialize(), Objects: walls, IsActive: true}}
	a, b := "ana", "beto"
	repo := &fakeRepo{fog: Fog{Mode: mode, Radius: 20}, pos: []Pos{{ID: "ta", X: 5, Z: 16, OwnerID: &a}, {ID: "tb", X: 20, Z: 16, OwnerID: &b}, {ID: "npc", X: 25, Z: 16}}}
	hub := &fakeHub{online: []realtime.Sender{gm, ana, beto}}
	return &rig{NewService(repo, sc, hub), repo, hub, sc}
}

func member(s realtime.Sender) campaigns.Member {
	return campaigns.Member{CampaignID: camp, UserID: s.UserID, Role: s.Role}
}

func (r *rig) visible(t *testing.T, s realtime.Sender) Bits {
	t.Helper()
	v, err := r.svc.Get(context.Background(), member(s), sceneID)
	if err != nil {
		t.Fatal(err)
	}
	return DecodeBits(v.Visible, 32, 32)
}

func TestDynamicVisionBlockedByWall(t *testing.T) {
	r := newRig(ModeDynamic)
	a := r.visible(t, ana)
	if !a.Has(8, 16) || a.Has(14, 16) || a.Has(25, 16) {
		t.Fatalf("ana should see her side only: near=%v behind=%v far=%v", a.Has(8, 16), a.Has(14, 16), a.Has(25, 16))
	}
	if b := r.visible(t, beto); !b.Has(25, 16) || b.Has(5, 16) {
		t.Fatal("beto sees only his side")
	}
}

func TestOffModeSendsNothingToPlayers(t *testing.T) {
	r := newRig(ModeOff)
	v, _ := r.svc.Get(context.Background(), member(ana), sceneID)
	if v.Mode != ModeOff || v.Visible != "" {
		t.Fatalf("off mode must not hide anything: %+v", v)
	}
}

func TestManualRevealUnionsAndHides(t *testing.T) {
	r := newRig(ModeDynamic)
	cells := []maps.Coord{{X: 14, Z: 16}, {X: 15, Z: 16}}
	if err := r.svc.Brush(context.Background(), gm, camp, sceneID, cells, true); err != nil {
		t.Fatal(err)
	}
	if a := r.visible(t, ana); !a.Has(14, 16) || !a.Has(8, 16) {
		t.Fatal("revealed cells join the dynamic vision")
	}
	if len(r.hub.got(ana, EventVision)) != 1 || len(r.hub.got(gm, EventFog)) != 1 || len(r.hub.got(ana, EventFog)) != 0 {
		t.Fatalf("ana gets vision.updated, only the Mestre gets fog.updated: %+v", r.hub.events)
	}
	if err := r.svc.Brush(context.Background(), gm, camp, sceneID, cells[:1], false); err != nil {
		t.Fatal(err)
	}
	if a := r.visible(t, ana); a.Has(14, 16) || !a.Has(15, 16) {
		t.Fatal("hide removes only the given cell")
	}
}

func TestManualModeSeesOnlyRevealed(t *testing.T) {
	r := newRig(ModeManual)
	if a := r.visible(t, ana); a.Has(5, 16) {
		t.Fatal("manual mode ignores tokens")
	}
	_ = r.svc.Brush(context.Background(), gm, camp, sceneID, []maps.Coord{{X: 1, Z: 1}}, true)
	if a := r.visible(t, ana); !a.Has(1, 1) {
		t.Fatal("revealed cell visible")
	}
}

func TestPermissionsAndValidation(t *testing.T) {
	r := newRig(ModeManual)
	ctx := context.Background()
	if err := r.svc.Brush(ctx, ana, camp, sceneID, nil, true); err == nil {
		t.Fatal("player must not reveal")
	}
	if err := r.svc.Configure(ctx, ana, camp, sceneID, ModeOff, 5); err == nil {
		t.Fatal("player must not configure")
	}
	if err := r.svc.Configure(ctx, gm, camp, sceneID, "weird", 5); err == nil {
		t.Fatal("invalid mode")
	}
	if err := r.svc.Configure(ctx, gm, camp, sceneID, ModeDynamic, 0); err == nil {
		t.Fatal("invalid radius")
	}
	if err := r.svc.Brush(ctx, gm, camp, sceneID, []maps.Coord{{X: 99, Z: 0}}, true); err == nil {
		t.Fatal("out of bounds")
	}
	v, _ := r.svc.Get(ctx, member(ana), sceneID)
	if v.Revealed != "" || v.VisionRadius != 0 {
		t.Fatal("players do not get the manual reveals or config")
	}
}

func TestConfigurePushesModeToPlayers(t *testing.T) {
	r := newRig(ModeOff)
	if err := r.svc.Configure(context.Background(), gm, camp, sceneID, ModeDynamic, 8); err != nil {
		t.Fatal(err)
	}
	if len(r.hub.got(beto, EventVision)) != 1 || len(r.hub.got(gm, EventFog)) != 1 {
		t.Fatal("mode change reaches players and the Mestre")
	}
	if r.repo.fog.Radius != 8 {
		t.Fatal("radius saved")
	}
}

func movedPayload(token string, x, z, fx, fz int) map[string]any {
	return map[string]any{"tokenId": token, "sceneId": sceneID, "x": x, "z": z, "fromX": fx, "fromZ": fz}
}

func TestMovedFilteredByVision(t *testing.T) {
	r := newRig(ModeDynamic)
	tap := r.svc.Tap()
	// NPC moves behind the wall: only the Mestre hears of it; Ana saw neither cell.
	tap.Broadcast(camp, realtime.ToAll, tokens.EventMoved, movedPayload("npc", 26, 16, 25, 16))
	if len(r.hub.got(ana, tokens.EventMoved))+len(r.hub.got(ana, tokens.EventRemoved)) != 0 {
		t.Fatal("ana must not learn about the hidden npc")
	}
	if len(r.hub.got(gm, tokens.EventMoved)) != 1 || len(r.hub.got(beto, tokens.EventMoved)) != 1 {
		t.Fatal("the Mestre and the viewer on that side still get the move")
	}
	// A token walking in view of Ana is delivered; leaving her view tells her to drop it.
	r.hub.reset()
	tap.Broadcast(camp, realtime.ToAll, tokens.EventMoved, movedPayload("npc", 6, 16, 7, 16))
	if len(r.hub.got(ana, tokens.EventMoved)) != 1 {
		t.Fatal("visible move delivered")
	}
	r.hub.reset()
	tap.Broadcast(camp, realtime.ToAll, tokens.EventMoved, movedPayload("npc", 26, 16, 6, 16))
	if len(r.hub.got(ana, tokens.EventRemoved)) != 1 || len(r.hub.got(ana, tokens.EventMoved)) != 0 {
		t.Fatal("a token leaving the view is removed, its new position is not leaked")
	}
}

func TestOwnerMoveRecalculatesVision(t *testing.T) {
	r := newRig(ModeDynamic)
	_ = r.visible(t, ana)
	r.repo.pos[0].X, r.repo.pos[0].Z = 20, 16
	r.svc.Tap().Broadcast(camp, realtime.ToAll, tokens.EventMoved, movedPayload("ta", 20, 16, 5, 16))
	pushes := r.hub.got(ana, EventVision)
	if len(pushes) != 1 || len(r.hub.got(beto, EventVision)) != 0 {
		t.Fatalf("only the mover gets a new vision.updated: %v", r.hub.events)
	}
	if a := r.visible(t, ana); !a.Has(25, 16) || a.Has(5, 16) {
		t.Fatal("ana now sees the far side")
	}
	if len(r.hub.got(ana, tokens.EventMoved)) != 1 {
		t.Fatal("the owner always gets her own move")
	}
}

func TestTerrainChangeRecalculates(t *testing.T) {
	r := newRig(ModeDynamic)
	r.svc.Tap().Broadcast(camp, realtime.ToAll, tokens.EventMoved, movedPayload("npc", 26, 16, 25, 16))
	r.hub.reset()
	r.sc.sc.Objects = nil // wall removed
	r.svc.Tap().Broadcast(camp, realtime.ToAll, scenes.EventObjectsPatched, map[string]any{"sceneId": sceneID})
	if len(r.hub.got(ana, scenes.EventObjectsPatched)) != 1 || len(r.hub.got(ana, EventVision)) != 1 {
		t.Fatal("event forwarded and vision recomputed")
	}
	if a := r.visible(t, ana); !a.Has(14, 16) {
		t.Fatal("the cell behind the removed wall is visible")
	}
}

func TestFilterListsOnlyVisibleTokens(t *testing.T) {
	r := newRig(ModeDynamic)
	a, b := "ana", "beto"
	list := []tokens.Token{{ID: "ta", X: 5, Z: 16, OwnerID: &a}, {ID: "tb", X: 20, Z: 16, OwnerID: &b}, {ID: "npc", X: 6, Z: 16}, {ID: "npc2", X: 25, Z: 16}}
	got := r.svc.Filter(context.Background(), member(ana), sceneID, list)
	if len(got) != 2 || got[0].ID != "ta" || got[1].ID != "npc" {
		t.Fatalf("ana sees her token and the npc in view: %+v", got)
	}
	if len(r.svc.Filter(context.Background(), member(gm), sceneID, list)) != 4 {
		t.Fatal("the Mestre sees everything")
	}
	r.repo.fog.Mode = ModeOff
	r.svc.invalidate(sceneID, "")
	if len(r.svc.Filter(context.Background(), member(ana), sceneID, list)) != 4 {
		t.Fatal("mode off shows everything")
	}
}

func TestBitsRoundTrip(t *testing.T) {
	b := NewBits(5, 3)
	b.Put(4, 2, true)
	b.Put(0, 0, true)
	b.Put(9, 9, true)
	d := DecodeBits(b.Encode(), 5, 3)
	if !d.Has(4, 2) || !d.Has(0, 0) || d.Has(1, 1) || d.Has(9, 9) {
		t.Fatal("round trip")
	}
	if DecodeBits("%%%", 5, 3).Has(0, 0) {
		t.Fatal("malformed input is empty")
	}
}
