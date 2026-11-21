package rolls

import (
	"context"
	"testing"
)

func TestRequestEntryAppliesTrustedBonusKeptAndHookFiresOnKeep(t *testing.T) {
	s, _, _ := newSvc(0, 0, 5, 4)
	var hooked []string
	s.OnFinalized(func(_ context.Context, _, rollID string) { hooked = append(hooked, rollID) })
	in := req(false)
	in.BonusKept = 1
	e, err := s.RequestEntry(ctx, camp, ana, in)
	if err != nil || e.RollID == "" || e.KeptCount != 2 {
		t.Fatalf("entry = %+v %v", e, err)
	}
	if err := s.Keep(ctx, camp, ana, e.RollID, e.Suggested); err != nil {
		t.Fatal(err)
	}
	if len(hooked) != 1 || hooked[0] != e.RollID {
		t.Fatalf("hooked = %v", hooked)
	}
}

func TestPoolIsMestreRollWithLimits(t *testing.T) {
	s, repo, _ := newSvc(0, 3)
	if err := s.Pool(ctx, camp, master, ThreatPool{D6: 2, D12: 1}); err != nil || len(repo.recs) != 1 || repo.recs[0].Kind != KindThreat {
		t.Fatalf("pool: %v %+v", err, repo.recs)
	}
	if s.Pool(ctx, camp, master, ThreatPool{}) == nil || s.Pool(ctx, camp, master, ThreatPool{D6: 21}) == nil {
		t.Fatal("empty or oversized pools must fail")
	}
}
