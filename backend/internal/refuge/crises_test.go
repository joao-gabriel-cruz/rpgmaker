package refuge

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

type fakeRolls struct {
	faces   []int
	symbols assimilacao.Symbols
	last    rolls.RequestInput
}

func (f *fakeRolls) PoolEntry(context.Context, string, realtime.Sender, string, rolls.ThreatPool) (rolls.Entry, error) {
	dice := make([]assimilacao.Die, len(f.faces))
	for i, v := range f.faces {
		dice[i] = assimilacao.Die{Kind: "d6", Face: v}
	}
	return rolls.Entry{RollID: "roll-1", Dice: dice}, nil
}

func (f *fakeRolls) RequestEntry(_ context.Context, _ string, _ realtime.Sender, in rolls.RequestInput) (rolls.Entry, error) {
	f.last = in
	return rolls.Entry{RollID: "roll-2"}, nil
}

func (f *fakeRolls) Final(context.Context, string, string) (rolls.Final, error) {
	return rolls.Final{Symbols: f.symbols}, nil
}

type fakeEffects struct{ got []effects.Suggestion }

func (f *fakeEffects) Suggest(_ context.Context, in effects.Suggestion) (effects.Suggestion, error) {
	f.got = append(f.got, in)
	return in, nil
}

type fakeLookup struct{ chars map[string]characters.View }

func (f fakeLookup) Get(_ context.Context, _ campaigns.Member, id string) (characters.View, error) {
	v, ok := f.chars[id]
	if !ok {
		return v, characters.ErrNotFound
	}
	return v, nil
}

func view(id, kind, name string, owner *string) characters.View {
	return characters.View{Character: characters.Character{ID: id, Kind: kind, Name: name, OwnerID: owner}}
}

func crisisSetup(t *testing.T) (*Service, *fakeRepo, *fakeRolls, *fakeEffects, *fakeEvents) {
	t.Helper()
	s, repo, _, ev := setup(`{}`)
	repo.r.CrisisSettings = DefaultCrisisSettings()
	repo.r.Morale = 3
	owner := "u-ana"
	fr, fe := &fakeRolls{}, &fakeEffects{}
	lookup := fakeLookup{chars: map[string]characters.View{
		"pc1": view("pc1", "pc", "Rita", &owner), "pc2": view("pc2", "pc", "Davi", nil),
		"npc1": view("npc1", "npc", "Dona Lia", nil),
	}}
	s.Link(Links{Rolls: fr, Effects: fe, Chars: lookup, Compendium: func(context.Context, string, string) (json.RawMessage, error) {
		return json.RawMessage(`{"name":"Faca"}`), nil
	}})
	return s, repo, fr, fe, ev
}

var gmSender = realtime.Sender{UserID: "u-gm", Role: campaigns.RoleGM}
var anaSender = realtime.Sender{UserID: "u-ana", Role: campaigns.RolePlayer}

func TestWeekTriggersBecomeSuggestedCrisesAndConfirmCostsMorale(t *testing.T) {
	s, repo, _, _, ev := crisisSetup(t)
	report := WeekReport{Triggers: []CrisisTrigger{{Kind: "overpopulation", Detail: "x"}, {Kind: "shortage", Resource: "water", Detail: "Faltou Água"}}}
	fire(s, "c", report)
	fire(s, "c", report)
	if len(repo.r.Crises) != 2 || repo.r.Crises[0].Status != StatusSuggested || repo.r.Crises[1].Name != "Falta de Água" {
		t.Fatalf("crises = %+v", repo.r.Crises)
	}
	if repo.r.Morale != 3 || !slices.Contains(ev.types, EventCrisisSuggested) {
		t.Errorf("morale %d events %v", repo.r.Morale, ev.types)
	}
	if err := s.ConfirmCrisis(ctx, "c", repo.r.Crises[0].ID); err != nil || repo.r.Morale != 2 || repo.r.Crises[0].Status != StatusActive {
		t.Fatalf("confirm = %v morale %d", err, repo.r.Morale)
	}
	repo.r.CrisisSettings.MoralePenalty = false
	_ = s.ConfirmCrisis(ctx, "c", repo.r.Crises[1].ID)
	if repo.r.Morale != 2 {
		t.Errorf("morale with option off = %d", repo.r.Morale)
	}
}

func TestSeverityBands(t *testing.T) {
	cases := []struct {
		faces   []int
		discard int
		symbols int
		kind    string
	}{
		{[]int{1, 2, 1, 6}, 3, 0, ""},
		{[]int{3, 1, 1, 6}, 3, 1, ""},
		{[]int{3, 4, 1, 2}, 2, 2, "internal_fight"},
		{[]int{3, 3, 6, 1}, 3, 3, "disease"},
		{[]int{5, 5, 1, 2}, 2, 4, "creature"},
		{[]int{5, 5, 6, 1}, 3, 5, "rebellion"},
		{[]int{5, 5, 5, 1}, 3, 6, "invasion"},
	}
	for _, c := range cases {
		if n := symbolsOf(c.faces, c.discard); n != c.symbols || severityFor(n).Kind != c.kind {
			t.Errorf("%v discard %d = %d (%s), want %d (%s)", c.faces, c.discard, n, severityFor(n).Kind, c.symbols, c.kind)
		}
	}
}

func TestRollDiscardDiseaseRaisesCondition(t *testing.T) {
	s, repo, fr, _, ev := crisisSetup(t)
	fr.faces = []int{3, 3, 6, 1}
	if err := s.RollSeverity(ctx, "c", gmSender, ""); err != nil || len(repo.r.Crises) != 1 || repo.r.Crises[0].Roll == nil {
		t.Fatalf("roll = %v %+v", err, repo.r.Crises)
	}
	id := repo.r.Crises[0].ID
	if err := s.DiscardDie(ctx, "c", id, 9); err == nil {
		t.Error("bad index accepted")
	}
	if err := s.DiscardDie(ctx, "c", id, 3); err != nil {
		t.Fatal(err)
	}
	c := repo.r.Crises[0]
	r, _ := s.Get(ctx, "c")
	if c.Kind != "disease" || c.Grade != 3 || !slices.Equal(r.Conditions, []string{"disease"}) || !slices.Contains(ev.types, EventCrisisRolled) {
		t.Errorf("crisis %+v conditions %v", c, r.Conditions)
	}
	if err := s.DiscardDie(ctx, "c", id, 3); err == nil {
		t.Error("discarded twice")
	}
	_ = s.UpdateCrisis(ctx, "c", CrisisPatch{ID: id, Status: ptr(StatusContained)})
	if r, _ = s.Get(ctx, "c"); len(r.Conditions) != 0 {
		t.Errorf("conditions after containing = %v", r.Conditions)
	}
}

func ptr[T any](v T) *T { return &v }

func TestInternalFightSuggestionMigratesViaMutate(t *testing.T) {
	s, repo, fr, _, _ := crisisSetup(t)
	repo.r.Population.Level = 6
	repo.r.Resources["water"], repo.r.Resources["food"] = 5, 3
	fr.faces = []int{3, 4, 1, 2}
	_ = s.RollSeverity(ctx, "c", gmSender, "")
	id := repo.r.Crises[0].ID
	if err := s.DiscardDie(ctx, "c", id, 2); err != nil {
		t.Fatal(err)
	}
	sg := repo.r.Crises[0].Suggestions
	if len(sg) != 1 || sg[0].Kind != "migration" || repo.r.Population.Level != 6 {
		t.Fatalf("suggestions = %+v pop %d", sg, repo.r.Population.Level)
	}
	if err := s.ResolveSuggestion(ctx, "c", id, sg[0].ID, true); err != nil {
		t.Fatal(err)
	}
	if repo.r.Population.Level != 3 || repo.r.Resources["water"] != 3 || repo.r.Resources["food"] != 2 {
		t.Errorf("after migration: pop %d res %v", repo.r.Population.Level, repo.r.Resources)
	}
	if err := s.ResolveSuggestion(ctx, "c", id, sg[0].ID, true); err == nil {
		t.Error("applied twice")
	}
}

func TestCrisisCRUDValidates(t *testing.T) {
	s, repo, _, _, _ := crisisSetup(t)
	if _, err := s.AddCrisis(ctx, "c", "u-gm", CrisisInput{Name: "Praga", Kind: "bogus"}); err == nil {
		t.Error("bad kind accepted")
	}
	made, err := s.AddCrisis(ctx, "c", "u-gm", CrisisInput{Name: "Praga", Kind: "disease", Grade: 4})
	if err != nil || len(repo.r.Crises) != 1 || repo.r.Morale != 3 {
		t.Fatalf("add = %v %+v", err, repo.r.Crises)
	}
	if err := s.UpdateCrisis(ctx, "c", CrisisPatch{ID: made.ID, Grade: ptr(11)}); err == nil {
		t.Error("grade 11 accepted")
	}
	if err := s.DeleteCrisis(ctx, "c", made.ID); err != nil || len(repo.r.Crises) != 0 {
		t.Errorf("delete = %v", err)
	}
}

func fire(s *Service, campaignID string, report WeekReport) {
	for _, h := range s.hooks {
		h(ctx, campaignID, report)
	}
}
