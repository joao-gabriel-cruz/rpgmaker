package characters

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

const cid = "11111111-1111-1111-1111-111111111111"

var (
	gm   = campaigns.Member{CampaignID: cid, UserID: "gm", Role: campaigns.RoleGM}
	ana  = campaigns.Member{CampaignID: cid, UserID: "ana", Role: campaigns.RolePlayer}
	bob  = campaigns.Member{CampaignID: cid, UserID: "bob", Role: campaigns.RolePlayer}
	ctx  = context.Background()
	over = `{"identity":{"name":"X"},"instincts":{"influence":3,"perception":3,"potency":3,"reaction":3,"resolution":3,"sagacity":3}}`
)

func newSvc(settings string) (*Service, *fakeRepo, *fakeEvents) {
	repo, ev := newFakeRepo(), &fakeEvents{}
	return NewService(repo, &fakeCampaigns{settings: settings}, systems.Default(), ev), repo, ev
}

func TestCreateOwnAndForOthers(t *testing.T) {
	s, _, ev := newSvc(`{}`)
	v, err := s.Create(ctx, ana, CreateInput{Name: "Rita"})
	if err != nil || *v.OwnerID != "ana" || v.Name != "Rita" || !strings.Contains(string(v.Sheet), `"name":"Rita"`) {
		t.Fatalf("own = %+v %v", v, err)
	}
	if len(ev.got) != 1 || ev.got[0].typ != "character.created" || ev.got[0].gmOnly {
		t.Errorf("events = %+v", ev.got)
	}
	if _, err := s.Create(ctx, ana, CreateInput{OwnerID: "bob"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("player for other = %v", err)
	}
	if v, err := s.Create(ctx, gm, CreateInput{OwnerID: "bob"}); err != nil || *v.OwnerID != "bob" || v.Name != defaultName {
		t.Errorf("gm for member = %+v %v", v, err)
	}
	if _, err := s.Create(ctx, gm, CreateInput{OwnerID: "outsider"}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("gm for outsider = %v", err)
	}
}

func TestDerivedFromPotencyAndResolution(t *testing.T) {
	s, _, _ := newSvc(`{}`)
	v, _ := s.Create(ctx, ana, CreateInput{})
	saved, err := s.Save(ctx, ana, v.ID, json.RawMessage(`{"identity":{"name":"Rita"},"instincts":{"potency":2,"resolution":1}}`))
	if err != nil || !strings.Contains(string(saved.Derived), `"boxesPerLevel":4`) || saved.Name != "Rita" {
		t.Fatalf("saved = %+v %v", saved, err)
	}
}

func TestWarnModeSavesWithIssues(t *testing.T) {
	s, _, ev := newSvc(`{"creationLimits":"warn"}`)
	v, _ := s.Create(ctx, ana, CreateInput{})
	saved, err := s.Save(ctx, ana, v.ID, json.RawMessage(over))
	if err != nil || len(saved.Issues) == 0 {
		t.Fatalf("warn = %+v %v", saved, err)
	}
	if last := ev.got[len(ev.got)-1]; last.typ != "character.updated" {
		t.Errorf("event = %+v", last)
	}
}

func TestBlockModeRejectsLimits(t *testing.T) {
	s, repo, ev := newSvc(`{"creationLimits":"block"}`)
	v, _ := s.Create(ctx, ana, CreateInput{})
	_, err := s.Save(ctx, ana, v.ID, json.RawMessage(over))
	var rej *RejectedError
	if !errors.As(err, &rej) || len(rej.Issues) == 0 {
		t.Fatalf("block = %v", err)
	}
	if strings.Contains(string(repo.byID[v.ID].Sheet), `"sagacity":3`) || len(ev.got) != 1 {
		t.Error("rejected sheet was stored or broadcast")
	}
}

func TestBlockingIssuesAlwaysReject(t *testing.T) {
	s, _, _ := newSvc(`{"creationLimits":"warn"}`)
	v, _ := s.Create(ctx, ana, CreateInput{})
	for _, sheet := range []string{`{"instincts":"x"}`, `{"health":{"damage":-1}}`} {
		var rej *RejectedError
		if _, err := s.Save(ctx, ana, v.ID, json.RawMessage(sheet)); !errors.As(err, &rej) {
			t.Errorf("%s = %v", sheet, err)
		}
	}
}

func TestPermissions(t *testing.T) {
	s, _, ev := newSvc(`{}`)
	v, _ := s.Create(ctx, ana, CreateInput{})
	ok := json.RawMessage(`{"identity":{"name":"Z"}}`)
	if _, err := s.Save(ctx, bob, v.ID, ok); !errors.Is(err, ErrForbidden) {
		t.Errorf("other player save = %v", err)
	}
	if err := s.Delete(ctx, bob, v.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("other player delete = %v", err)
	}
	if g, err := s.Get(ctx, bob, v.ID); err != nil || g.ID != v.ID {
		t.Errorf("read other = %v", err)
	}
	if _, err := s.Save(ctx, gm, v.ID, ok); err != nil {
		t.Errorf("gm save = %v", err)
	}
	if err := s.Delete(ctx, ana, v.ID); err != nil {
		t.Errorf("owner delete = %v", err)
	}
	if last := ev.got[len(ev.got)-1]; last.typ != "character.deleted" {
		t.Errorf("event = %+v", last)
	}
	if _, err := s.Get(ctx, ana, v.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("after delete = %v", err)
	}
	if _, err := s.Get(ctx, ana, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("bad id = %v", err)
	}
}

func TestPlayersOnlySeePCs(t *testing.T) {
	s, repo, _ := newSvc(`{}`)
	if _, err := s.Create(ctx, ana, CreateInput{}); err != nil {
		t.Fatal(err)
	}
	npc, _ := repo.Create(ctx, Character{CampaignID: cid, Kind: KindNPC, Name: "N", Sheet: json.RawMessage(`{}`)})
	if l, _ := s.List(ctx, bob); len(l) != 1 || l[0].Kind != KindPC {
		t.Errorf("player list = %+v", l)
	}
	if l, _ := s.List(ctx, gm); len(l) != 2 {
		t.Errorf("gm list = %d", len(l))
	}
	if _, err := s.Get(ctx, bob, npc.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("npc for player = %v", err)
	}
	if _, err := s.Get(ctx, gm, npc.ID); err != nil {
		t.Errorf("npc for gm = %v", err)
	}
}

func TestInvalidAppearanceRejected(t *testing.T) {
	s, repo, ev := newSvc(`{}`)
	v, _ := s.Create(ctx, ana, CreateInput{})
	bad := json.RawMessage(`{"identity":{"name":"X"},"appearance":{"hair":{"style":"afro","color":"red"}}}`)
	_, err := s.Save(ctx, ana, v.ID, bad)
	var rej *RejectedError
	if !errors.As(err, &rej) || len(rej.Issues) != 2 {
		t.Fatalf("bad appearance = %v", err)
	}
	if strings.Contains(string(repo.byID[v.ID].Sheet), "afro") || len(ev.got) != 1 {
		t.Error("rejected appearance was stored or broadcast")
	}
	good := json.RawMessage(`{"identity":{"name":"X"},"appearance":{"hair":{"style":"long","color":"#112233"}}}`)
	if _, err := s.Save(ctx, ana, v.ID, good); err != nil {
		t.Errorf("good appearance = %v", err)
	}
	if _, err := s.Save(ctx, ana, v.ID, json.RawMessage(`{"identity":{"name":"X"}}`)); err != nil {
		t.Errorf("sheet without appearance = %v", err)
	}
}

func TestEditSheetKeepsFieldsAndAnnounces(t *testing.T) {
	s, _, ev := newSvc(`{}`)
	v, _ := s.Create(ctx, ana, CreateInput{Name: "Rita"})
	before := len(ev.got)
	out, err := s.EditSheet(ctx, cid, v.ID, func(sh json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(strings.Replace(string(sh), `"notes":""`, `"notes":"x"`, 1)), nil
	})
	if err != nil || !strings.Contains(string(out.Sheet), `"notes":"x"`) || out.Name != "Rita" {
		t.Fatalf("edit = %+v %v", out, err)
	}
	if len(ev.got) != before+1 || ev.got[before].typ != "character.updated" {
		t.Errorf("events = %+v", ev.got)
	}
	if _, err := s.EditSheet(ctx, cid, "nope", nil); !errors.Is(err, ErrNotFound) {
		t.Errorf("bad id = %v", err)
	}
	boom := errors.New("boom")
	if _, err := s.EditSheet(ctx, cid, v.ID, func(json.RawMessage) (json.RawMessage, error) { return nil, boom }); !errors.Is(err, boom) {
		t.Errorf("fn error = %v", err)
	}
}
