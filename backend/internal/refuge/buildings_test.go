package refuge

import (
	"errors"
	"testing"
)

func TestAllocationRules(t *testing.T) {
	r := weekRefuge(Building{ID: "a", Kind: KindQuarry, Level: 1, State: StatePlanned}, done(KindSawmill, 0))
	r.Buildings[1].ID = "b"
	if err := r.allocate(map[string]int{"a": 3, "b": 3}); !errors.Is(err, errAllocOver) {
		t.Errorf("over = %v", err)
	}
	if err := r.allocate(map[string]int{"a": -1}); !errors.Is(err, errAllocNeg) {
		t.Errorf("neg = %v", err)
	}
	if err := r.allocate(map[string]int{"zzz": 1}); !errors.Is(err, errNoBuild) {
		t.Errorf("unknown = %v", err)
	}
	if err := r.allocate(map[string]int{"a": 3, "b": 2}); err != nil || r.Buildings[0].State != StateBuilding || r.Buildings[1].Workers != 2 {
		t.Errorf("ok = %v %+v", err, r.Buildings)
	}
	if err := r.allocate(map[string]int{"a": 1}); err != nil || r.Buildings[1].Workers != 0 {
		t.Errorf("replace = %v %+v", err, r.Buildings)
	}
}

func TestProposalsFlow(t *testing.T) {
	s, repo, _, _ := setup(sheet)
	repo.r = weekRefuge()
	if _, err := s.ProposeBuilding(ctx, ana, "Ana", KindSawmill); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ProposeBuilding(ctx, bob, "Bob", KindSawmill); !errors.Is(err, errHasKind) {
		t.Errorf("duplicate = %v", err)
	}
	if _, err := s.ProposeBuilding(ctx, ana, "Ana", "castle"); !errors.Is(err, errBadKind) {
		t.Errorf("bad kind = %v", err)
	}
	id := repo.r.Proposals[0].ID
	if _, err := s.ResolveProposal(ctx, ana, id, true); err == nil {
		t.Error("player approved")
	}
	r, err := s.ResolveProposal(ctx, gm, id, true)
	if err != nil || len(r.Proposals) != 0 || len(r.Buildings) != 1 || r.Buildings[0].State != StatePlanned {
		t.Fatalf("approve = %+v %v", r, err)
	}
	if _, err := s.ProposeBuilding(ctx, bob, "Bob", KindQuarry); err != nil {
		t.Fatal(err)
	}
	r, err = s.ResolveProposal(ctx, gm, repo.r.Proposals[0].ID, false)
	if err != nil || len(r.Proposals) != 0 || len(r.Buildings) != 1 {
		t.Errorf("reject = %+v %v", r, err)
	}
}

func TestAddAndCancelBuilding(t *testing.T) {
	s, repo, _, _ := setup(sheet)
	repo.r = weekRefuge(done(KindQuarry, 0))
	if _, err := s.AddBuilding(ctx, ana, KindMill); err == nil {
		t.Error("player added")
	}
	r, err := s.AddBuilding(ctx, gm, KindMill)
	if err != nil || len(r.Buildings) != 2 {
		t.Fatalf("add = %v", err)
	}
	if _, err := s.AddBuilding(ctx, gm, KindMill); !errors.Is(err, errHasKind) {
		t.Errorf("dup = %v", err)
	}
	if _, err := s.CancelBuilding(ctx, gm, KindQuarry); !errors.Is(err, errCancel) {
		t.Errorf("cancel built = %v", err)
	}
	if r, err := s.CancelBuilding(ctx, gm, r.Buildings[1].ID); err != nil || len(r.Buildings) != 1 {
		t.Errorf("cancel = %v", err)
	}
}
