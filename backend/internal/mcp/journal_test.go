package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/ambience"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

var (
	_ Scenes     = (*scenes.Service)(nil)
	_ Tokens     = (*tokens.Service)(nil)
	_ Ambience   = (*ambience.Service)(nil)
	_ Characters = (*characters.Service)(nil)
)

func cellInverse(terrain string) Inverse {
	return Inverse{Cells: []maps.CellPatch{{X: 1, Z: 2, Terrain: terrain, Height: 3}}}
}

func userErr(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("err = %v, want containing %q", err, want)
	}
}

func TestRecordCapsAtTwenty(t *testing.T) {
	r := newRig("s1")
	for i := 0; i < 21; i++ {
		r.j.Record("c1", "s1", fmt.Sprintf("change %d", i), Inverse{})
	}
	list := r.j.List("c1")
	if len(list) != 20 || list[0].Summary != "change 20" || list[19].Summary != "change 1" {
		t.Fatalf("list = %+v", list)
	}
}

func TestRecordBroadcastsLatestToGM(t *testing.T) {
	r := newRig("s1")
	r.j.now = func() time.Time { return time.Unix(100, 0) }
	v := r.j.Record("c1", "s1", "x", Inverse{})
	got := r.n.last()
	p := got.Payload.(map[string]any)
	if got.Type != "mcp.changed" || p["latest"] != v || len(p["changes"].([]ChangeView)) != 1 {
		t.Fatalf("sent = %+v", got)
	}
	if !got.To(realtime.Sender{Role: "gm"}) || got.To(realtime.Sender{Role: "player"}) {
		t.Fatal("filter is not ToGM")
	}
	if !v.At.Equal(time.Unix(100, 0)) || v.ID == "" {
		t.Fatalf("view = %+v", v)
	}
}

func TestUndoRestoresCellsAndObjects(t *testing.T) {
	r := newRig("s1")
	obj := &scenes.ObjectPatch{Remove: []string{"o1"}}
	v := r.j.Record("c1", "s1", "paint", Inverse{Cells: cellInverse("grass").Cells, Objects: obj})
	sum, err := r.j.Undo(context.Background(), r.t, r.s, v.ID)
	if err != nil || sum != "paint" {
		t.Fatalf("sum=%q err=%v", sum, err)
	}
	if len(r.sc.painted) != 1 || r.sc.painted[0].Cells[0].Terrain != "grass" || len(r.sc.objects) != 1 {
		t.Fatalf("painted=%+v objects=%+v", r.sc.painted, r.sc.objects)
	}
	for _, o := range r.sc.origins {
		if o != "" {
			t.Fatalf("origin = %q", o)
		}
	}
	if len(r.j.List("c1")) != 0 {
		t.Fatal("entry not removed")
	}
	p := r.n.last().Payload.(map[string]any)
	if _, has := p["latest"]; has {
		t.Fatal("undo broadcast must not carry latest")
	}
}

func TestUndoTokenAmbienceAndCondition(t *testing.T) {
	r := newRig("s1")
	r.j.Record("c1", "s1", "mix", Inverse{
		Token:     &TokenPos{"tk", 4, 5},
		Ambience:  &ambience.Ambience{SceneID: "s1"},
		Condition: &ConditionState{"ch", "ferido", true},
	})
	if _, err := r.j.Undo(context.Background(), r.t, r.s, ""); err != nil {
		t.Fatal(err)
	}
	if len(r.tk.moves) != 1 || len(r.amb.set) != 1 || len(r.amb.conditions) != 1 {
		t.Fatalf("moves=%v set=%v cond=%v", r.tk.moves, r.amb.set, r.amb.conditions)
	}
	if r.amb.senders[0].UserID != "gm" || r.amb.senders[0].Name != "Mestre" {
		t.Fatalf("sender = %+v", r.amb.senders[0])
	}
}

func TestUndoAnyOrder(t *testing.T) {
	r := newRig("s1")
	a := r.j.Record("c1", "s1", "a", cellInverse("a"))
	b := r.j.Record("c1", "s1", "b", cellInverse("b"))
	c := r.j.Record("c1", "s1", "c", cellInverse("c"))
	if _, err := r.j.Undo(context.Background(), r.t, r.s, b.ID); err != nil {
		t.Fatal(err)
	}
	list := r.j.List("c1")
	if len(list) != 2 || list[0].ID != c.ID || list[1].ID != a.ID {
		t.Fatalf("list = %+v", list)
	}
}

func TestUndoTwiceFails(t *testing.T) {
	r := newRig("s1")
	v := r.j.Record("c1", "s1", "a", cellInverse("a"))
	if _, err := r.j.Undo(context.Background(), r.t, r.s, v.ID); err != nil {
		t.Fatal(err)
	}
	_, err := r.j.Undo(context.Background(), r.t, r.s, v.ID)
	userErr(t, err, "Essa mudança não existe mais (talvez já tenha sido desfeita)")
	if len(r.sc.painted) != 1 {
		t.Fatal("second undo must not apply anything")
	}
}

func TestUndoDeletedScene(t *testing.T) {
	r := newRig("s1")
	v := r.j.Record("c1", "s1", "a", cellInverse("a"))
	delete(r.sc.scenes, "s1")
	_, err := r.j.Undo(context.Background(), r.t, r.s, v.ID)
	userErr(t, err, "A cena dessa mudança não existe mais")
	if len(r.j.List("c1")) != 0 || len(r.sc.painted) != 0 {
		t.Fatal("entry must be dropped and nothing applied")
	}
}

func TestUndoInactiveSceneStillApplies(t *testing.T) {
	r := newRig("s1", "s2")
	r.sc.scenes["s2"].IsActive = true
	v := r.j.Record("c1", "s1", "a", cellInverse("a"))
	if _, err := r.j.Undo(context.Background(), r.t, r.s, v.ID); err != nil || len(r.sc.painted) != 1 || r.sc.hit[0] != "s1" {
		t.Fatalf("err=%v painted=%d", err, len(r.sc.painted))
	}
}

func TestUndoRefusesOpenDraft(t *testing.T) {
	r := newRig("s1")
	a := r.j.Record("c1", "s1", "a", cellInverse("a"))
	b := r.j.Record("c1", "s1", "b", cellInverse("b"))
	r.sc.scenes["s1"].Draft = &scenes.Draft{}
	_, err := r.j.Undo(context.Background(), r.t, r.s, a.ID)
	userErr(t, err, "há um rascunho aberto nesta cena; aplique ou descarte no editor antes")
	list := r.j.List("c1")
	if len(list) != 2 || list[0].ID != b.ID || list[1].ID != a.ID || len(r.sc.painted) != 0 {
		t.Fatalf("entry not kept in place: %+v", list)
	}
}

func TestUndoGenericErrorKeepsEntry(t *testing.T) {
	r := newRig("s1")
	v := r.j.Record("c1", "s1", "a", cellInverse("a"))
	r.sc.paintErr = errors.New("boom")
	if _, err := r.j.Undo(context.Background(), r.t, r.s, v.ID); err == nil {
		t.Fatal("want error")
	}
	if l := r.j.List("c1"); len(l) != 1 || l[0].ID != v.ID {
		t.Fatalf("list = %+v", l)
	}
}

func TestUndoNotFoundFromServiceDropsEntry(t *testing.T) {
	r := newRig("s1")
	v := r.j.Record("c1", "s1", "a", cellInverse("a"))
	r.sc.paintErr = scenes.ErrNotFound
	_, err := r.j.Undo(context.Background(), r.t, r.s, v.ID)
	userErr(t, err, "A cena dessa mudança não existe mais")
	if len(r.j.List("c1")) != 0 {
		t.Fatal("entry must be dropped")
	}
	if _, has := r.n.last().Payload.(map[string]any)["latest"]; has {
		t.Fatal("drop broadcast must not carry latest")
	}
}

func TestPutBackKeepsTimeOrder(t *testing.T) {
	r := newRig("s1")
	tick := 0
	r.j.now = func() time.Time { tick++; return time.Unix(int64(tick), 0) }
	a := r.j.Record("c1", "s1", "a", cellInverse("a"))
	b := r.j.Record("c1", "s1", "b", cellInverse("b"))
	c := r.j.Record("c1", "s1", "c", cellInverse("c"))
	taken, _ := r.j.take("c1", b.ID)
	if _, err := r.j.Undo(context.Background(), r.t, r.s, c.ID); err != nil {
		t.Fatal(err)
	}
	r.j.putBack("c1", taken)
	l := r.j.List("c1")
	if len(l) != 2 || l[0].ID != b.ID || l[1].ID != a.ID {
		t.Fatalf("list = %+v", l)
	}
}

func TestUndoPermanentUserErrorDropsEntry(t *testing.T) {
	r := newRig("s1")
	v := r.j.Record("c1", "s1", "move", Inverse{Token: &TokenPos{"tk", 1, 1}})
	r.tk.moveErr = realtime.UserError("Token não encontrado")
	_, err := r.j.Undo(context.Background(), r.t, r.s, v.ID)
	userErr(t, err, "não dá para desfazer: Token não encontrado (removida da lista)")
	if len(r.j.List("c1")) != 0 {
		t.Fatal("entry must be dropped")
	}
	if _, has := r.n.last().Payload.(map[string]any)["latest"]; has {
		t.Fatal("drop broadcast must not carry latest")
	}
}

func TestUndoInvalidPatchDropsEntry(t *testing.T) {
	r := newRig("s1")
	v := r.j.Record("c1", "s1", "a", cellInverse("a"))
	r.sc.paintErr = fmt.Errorf("%w: cell out of range", scenes.ErrInvalidPatch)
	_, err := r.j.Undo(context.Background(), r.t, r.s, v.ID)
	userErr(t, err, "(removida da lista)")
	if len(r.j.List("c1")) != 0 {
		t.Fatal("entry must be dropped")
	}
}
