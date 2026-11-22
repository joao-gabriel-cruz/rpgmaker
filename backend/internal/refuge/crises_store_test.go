package refuge

import "testing"

func TestWeekHookPersistsSuggestedCrisesInPostgres(t *testing.T) {
	st, _, campID, _ := seedDB(t)
	s := NewService(st, &fakeChars{}, &fakeEvents{})
	r, _ := st.Get(ctx, campID)
	if !r.CrisisSettings.MoralePenalty || len(r.Crises) != 0 || len(r.Expeditions) != 0 {
		t.Fatalf("fresh refuge = %+v", r)
	}
	if _, err := s.Mutate(ctx, campID, func(r *Refuge) error { r.Morale = 4; return nil }); err != nil {
		t.Fatal(err)
	}
	fire(s, campID, WeekReport{Triggers: []CrisisTrigger{{Kind: "breach", Detail: "Defesa 1 < Perigo 3"}, {Kind: "shortage", Resource: "food", Detail: "Faltou Alimento"}}})
	r, _ = s.Get(ctx, campID)
	if len(r.Crises) != 2 || r.Crises[1].Name != "Falta de Alimento" || r.Crises[0].Status != StatusSuggested || r.Morale != 4 {
		t.Fatalf("crises = %+v morale %d", r.Crises, r.Morale)
	}
	if err := s.ConfirmCrisis(ctx, campID, r.Crises[0].ID); err != nil {
		t.Fatal(err)
	}
	if r, _ = s.Get(ctx, campID); r.Morale != 3 || r.Crises[0].Status != StatusActive {
		t.Errorf("after confirm: morale %d crises %+v", r.Morale, r.Crises)
	}
}
