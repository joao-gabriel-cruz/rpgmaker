package rolls

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

func chars(s *Service) *fakeChars { return s.chars.(*fakeChars) }

func detOf(t *testing.T, raw json.RawMessage) int {
	t.Helper()
	var doc struct {
		Tug struct {
			Points int `json:"determinationPoints"`
		} `json:"tugOfWar"`
		Instincts map[string]int `json:"instincts"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil || doc.Instincts["reaction"] != 2 {
		t.Fatalf("saved sheet = %s %v", raw, err)
	}
	return doc.Tug.Points
}

func TestEffortCostsDeterminationAndKeepsExtraDie(t *testing.T) {
	s, _, ev := newSvc(0, 0, 5, 4)
	in := req(false)
	in.Effort = true
	if err := s.Request(ctx, camp, ana, in); err != nil {
		t.Fatal(err)
	}
	if got := detOf(t, chars(s).saved); got != 4 {
		t.Errorf("determination = %d, want 4", got)
	}
	if e := ev.got[0].payload.(Entry); e.KeptCount != 2 || !e.Request.Effort {
		t.Errorf("entry = %+v", e)
	}
}

func TestEffortWithoutDeterminationFails(t *testing.T) {
	s, repo, _ := newSvc(0, 1)
	chars(s).det = 0
	in := req(false)
	in.Effort = true
	err := s.Request(ctx, camp, ana, in)
	if err == nil || !isRule(err) || len(repo.recs) != 0 || chars(s).saved != nil {
		t.Fatalf("err = %v recs=%d", err, len(repo.recs))
	}
}

func TestOccupationOncePerSessionAndResetsOnNewSession(t *testing.T) {
	s, _, _ := newSvc(0, 1)
	in := req(false)
	in.UseOccupation = true
	if err := s.Request(ctx, camp, ana, in); err != nil {
		t.Fatal(err)
	}
	if err := s.Request(ctx, camp, ana, in); !isUserError(err) {
		t.Fatalf("second use = %v", err)
	}
	st, _ := s.State(ctx, member(camp, ana), charID)
	if !st.OccupationUsed || st.MarkingEventUsed || !st.SessionActive {
		t.Errorf("state = %+v", st)
	}
	s.sess.(*fakeSessions).id = "sess-2"
	if err := s.Request(ctx, camp, ana, in); err != nil {
		t.Errorf("new session = %v", err)
	}
}

func TestMarkingEventIndependentAndNeedsSession(t *testing.T) {
	s, _, _ := newSvc(0, 1)
	in := req(false)
	in.UseMarkingEvent = true
	if err := s.Request(ctx, camp, ana, in); err != nil {
		t.Fatal(err)
	}
	in.UseOccupation = true
	if err := s.Request(ctx, camp, ana, in); !isUserError(err) {
		t.Errorf("marking twice = %v", err)
	}
	s.sess.(*fakeSessions).id = ""
	if err := s.Request(ctx, camp, ana, in); !isUserError(err) {
		t.Errorf("no session = %v", err)
	}
}

func TestAssimilatedKeepsAtLeastTwo(t *testing.T) {
	s, repo, ev := newSvc(0, 0, 5, 4)
	in := req(false)
	in.SecondInstinct = "potency"
	if err := s.Request(ctx, camp, ana, in); err != nil {
		t.Fatal(err)
	}
	e := ev.got[0].payload.(Entry)
	if e.KeptCount != 2 || e.Dice[0].Kind != "d12" || repo.recs[0].PlaySessionID == nil || *repo.recs[0].PlaySessionID != "sess-1" {
		t.Fatalf("entry = %+v", e)
	}
}

func TestDifficultyConsumedOnce(t *testing.T) {
	s, _, ev := newSvc(0, 1)
	if err := s.SetDifficulty(ctx, camp, master, charID, 2); err != nil {
		t.Fatal(err)
	}
	if got := ev.got[0]; got.typ != EventDifficulty || !got.toGM || got.toPlayer {
		t.Errorf("event = %+v", got)
	}
	_ = s.Request(ctx, camp, ana, req(false))
	_ = s.Request(ctx, camp, ana, req(false))
	first, second := ev.got[1].payload.(Entry), ev.got[2].payload.(Entry)
	if n := countDifficulty(first); n != 2 {
		t.Errorf("first difficulty dice = %d", n)
	}
	if n := countDifficulty(second); n != 0 {
		t.Errorf("second difficulty dice = %d", n)
	}
}

func countDifficulty(e Entry) int {
	n := 0
	for _, d := range e.Dice {
		if d.Difficulty && d.Kind == "d12" {
			n++
		}
	}
	return n
}

func TestDifficultyPermissionsAndLimits(t *testing.T) {
	s, _, ev := newSvc(0, 1)
	in := req(false)
	in.DifficultyD12 = 1
	if err := s.Request(ctx, camp, ana, in); !isUserError(err) {
		t.Errorf("player difficulty = %v", err)
	}
	if err := s.Request(ctx, camp, master, in); err != nil || countDifficulty(ev.got[0].payload.(Entry)) != 1 {
		t.Errorf("gm explicit = %v", err)
	}
	if err := s.SetDifficulty(ctx, camp, master, charID, 4); !isUserError(err) {
		t.Errorf("limit = %v", err)
	}
	if err := s.SetDifficulty(ctx, camp, master, "nope", 1); !errors.Is(err, errNotFoundChar()) {
		t.Errorf("unknown = %v", err)
	}
}

func TestStateHidesDifficultyFromOthers(t *testing.T) {
	s, _, _ := newSvc(0, 1)
	_ = s.SetDifficulty(ctx, camp, master, charID, 3)
	other := campaigns.Member{CampaignID: camp, UserID: beaID, Role: campaigns.RolePlayer}
	if st, _ := s.State(ctx, other, charID); st.DifficultyD12 != 0 {
		t.Errorf("other = %+v", st)
	}
	if st, _ := s.State(ctx, member(camp, ana), charID); st.DifficultyD12 != 3 {
		t.Errorf("owner = %+v", st)
	}
}

func isRule(err error) bool { return errors.As(err, new(assimilacao.RuleError)) }
