package characters

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestOnlyGMCreatesNPCs(t *testing.T) {
	s, _, ev := newSvc(`{}`)
	if _, err := s.Create(ctx, ana, CreateInput{Kind: KindNPC, Name: "Saqueador"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("player npc = %v", err)
	}
	if _, err := s.Create(ctx, gm, CreateInput{Kind: "boss"}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("bad kind = %v", err)
	}
	v, err := s.Create(ctx, gm, CreateInput{Kind: KindNPC, Name: "Saqueador"})
	if err != nil || v.Kind != KindNPC || v.OwnerID != nil || !strings.Contains(string(v.Sheet), `"threats":[]`) {
		t.Fatalf("gm npc = %+v %v", v, err)
	}
	if len(ev.got) != 1 || ev.got[0].typ != "character.created" || !ev.got[0].gmOnly {
		t.Errorf("npc event reached players: %+v", ev.got)
	}
	if !strings.Contains(string(v.Derived), `"boxesPerLevel":3`) {
		t.Errorf("derived = %s", v.Derived)
	}
}

func TestNPCHiddenAndEditableOnlyByGM(t *testing.T) {
	s, _, ev := newSvc(`{}`)
	v, _ := s.Create(ctx, gm, CreateInput{Kind: KindNPC, Name: "N"})
	sheet := json.RawMessage(`{"identity":{"name":"Saqueador"},"instincts":{"potency":9},"health":{"damage":0,"boxesPerLevel":5},"threats":["Emboscada"]}`)
	if _, err := s.Get(ctx, ana, v.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("player get = %v", err)
	}
	if _, err := s.Save(ctx, ana, v.ID, sheet); !errors.Is(err, ErrNotFound) {
		t.Errorf("player put = %v", err)
	}
	if err := s.Delete(ctx, ana, v.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("player delete = %v", err)
	}
	saved, err := s.Save(ctx, gm, v.ID, sheet)
	if err != nil || saved.Name != "Saqueador" || !strings.Contains(string(saved.Derived), `"boxesPerLevel":5`) {
		t.Fatalf("gm save = %+v %v", saved, err)
	}
	if last := ev.got[len(ev.got)-1]; last.typ != "character.updated" || !last.gmOnly {
		t.Errorf("update event = %+v", last)
	}
	var rej *RejectedError
	if _, err := s.Save(ctx, gm, v.ID, json.RawMessage(`{"health":{"boxesPerLevel":0}}`)); !errors.As(err, &rej) {
		t.Errorf("structural issue = %v", err)
	}
	if err := s.Delete(ctx, gm, v.ID); err != nil || !ev.got[len(ev.got)-1].gmOnly {
		t.Errorf("delete = %v %+v", err, ev.got)
	}
}

func TestNPCHTTPFlow(t *testing.T) {
	e := newDBEnv(t)
	if code, _ := e.do("ana", "POST", "/characters", `{"kind":"npc","name":"X"}`); code != 403 {
		t.Errorf("player create npc = %d", code)
	}
	code, npc := e.do("gm", "POST", "/characters", `{"kind":"npc","name":"Saqueador"}`)
	if code != 201 || npc["kind"] != "npc" {
		t.Fatalf("gm create = %d %v", code, npc)
	}
	id := npc["id"].(string)
	if code, _ := e.do("ana", "GET", "/characters/"+id, ""); code != 404 {
		t.Errorf("player get = %d", code)
	}
	if code, _ := e.do("ana", "PUT", "/characters/"+id, `{"sheet":{}}`); code != 404 {
		t.Errorf("player put = %d", code)
	}
	body := `{"sheet":{"identity":{"name":"Saqueador 2"},"health":{"damage":4,"boxesPerLevel":2},"threats":["x"]}}`
	if code, out := e.do("gm", "PUT", "/characters/"+id, body); code != 200 || out["name"] != "Saqueador 2" {
		t.Errorf("gm put = %d %v", code, out)
	}
	l, err := e.store.List(ctx, e.camp.ID, KindNPC)
	if err != nil || len(l) != 1 || l[0].OwnerID != nil {
		t.Errorf("store npc list = %+v %v", l, err)
	}
}
