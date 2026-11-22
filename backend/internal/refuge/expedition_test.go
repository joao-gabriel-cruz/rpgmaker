package refuge

import (
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

func runExpedition(t *testing.T, s *Service, fr *fakeRolls, success int) Expedition {
	t.Helper()
	fr.symbols = assimilacao.Symbols{Success: success, Adaptation: 1, Pressure: 1}
	if err := s.StartExpedition(ctx, "c", gmSender, ExpeditionInput{CharacterIDs: []string{"pc1", "pc2"}, Resource: "water"}); err != nil {
		t.Fatal(err)
	}
	id := s.mustRepo().r.Expeditions[0].ID
	if err := s.RollExpedition(ctx, "c", gmSender, id, "perception"); err != nil {
		t.Fatal(err)
	}
	s.ExpeditionFinalized(ctx, "c", "roll-2")
	return s.mustRepo().r.Expeditions[0]
}

func (s *Service) mustRepo() *fakeRepo { return s.repo.(*fakeRepo) }

func TestExpeditionByAbundanceAndCap(t *testing.T) {
	for _, c := range []struct {
		abundance string
		success   int
		found     int
		stocked   int
		discarded int
	}{{"none", 3, 0, 0, 0}, {"scarce", 3, 3, 3, 0}, {"medium", 3, 6, 6, 0}, {"abundant", 2, 6, 6, 0}, {"abundant", 5, 15, 10, 5}} {
		s, repo, fr, _, _ := crisisSetup(t)
		repo.r.Region.Abundance["water"] = c.abundance
		e := runExpedition(t, s, fr, c.success)
		if e.Found != c.found || repo.r.Resources["water"] != c.stocked || e.Discarded != c.discarded || e.Status != ExpDone {
			t.Errorf("%s x%d: %+v water %d", c.abundance, c.success, e, repo.r.Resources["water"])
		}
		if fr.last.Aptitude != "survival" || fr.last.BonusKept != 1 || fr.last.CharacterID != "pc1" {
			t.Errorf("roll input = %+v", fr.last)
		}
	}
}

func TestExpeditionExtrasAndComplications(t *testing.T) {
	s, repo, fr, fe, _ := crisisSetup(t)
	repo.r.Region.Abundance["water"] = "medium"
	e := runExpedition(t, s, fr, 1)
	if e.Extras != 1 || e.Complications != 1 {
		t.Fatalf("expedition = %+v", e)
	}
	if err := s.ExpeditionExtra(ctx, "c", ExtraInput{ID: e.ID, Resource: "food"}); err != nil || repo.r.Resources["food"] != 1 {
		t.Fatalf("extra = %v food %d", err, repo.r.Resources["food"])
	}
	if err := s.ExpeditionExtra(ctx, "c", ExtraInput{ID: e.ID, ItemID: "i1"}); err == nil {
		t.Error("extra beyond Adaptations accepted")
	}
	if err := s.ExpeditionComplication(ctx, "c", ComplicationInput{ID: e.ID, Kind: "injury", CharacterID: "pc2", Amount: 2}); err != nil {
		t.Fatal(err)
	}
	if len(fe.got) != 1 || fe.got[0].Effect != "damage" || fe.got[0].Amount != 2 || fe.got[0].CharacterName != "Davi" {
		t.Errorf("suggestion = %+v", fe.got)
	}
	if err := s.ExpeditionComplication(ctx, "c", ComplicationInput{ID: e.ID, Kind: "loss", Resource: "food"}); err == nil {
		t.Error("complication beyond Pressures accepted")
	}
}

func TestExpeditionItemExtraAndPlayerProposal(t *testing.T) {
	s, repo, fr, _, _ := crisisSetup(t)
	if err := s.StartExpedition(ctx, "c", anaSender, ExpeditionInput{CharacterIDs: []string{"pc2"}, Resource: "water"}); err == nil {
		t.Error("player led with a character that is not theirs")
	}
	if err := s.StartExpedition(ctx, "c", anaSender, ExpeditionInput{CharacterIDs: []string{"pc1"}, Resource: "water"}); err != nil {
		t.Fatal(err)
	}
	id := repo.r.Expeditions[0].ID
	if repo.r.Expeditions[0].Status != ExpProposed || s.RollExpedition(ctx, "c", anaSender, id, "sagacity") == nil {
		t.Fatal("proposal rolled before approval")
	}
	if err := s.ApproveExpedition(ctx, "c", id); err != nil {
		t.Fatal(err)
	}
	fr.symbols = assimilacao.Symbols{Adaptation: 1}
	if err := s.RollExpedition(ctx, "c", anaSender, id, "sagacity"); err != nil {
		t.Fatal(err)
	}
	s.ExpeditionFinalized(ctx, "c", "roll-2")
	if err := s.ExpeditionExtra(ctx, "c", ExtraInput{ID: id, ItemID: "i1"}); err != nil || len(repo.r.Items) != 1 {
		t.Errorf("item extra = %v items %d", err, len(repo.r.Items))
	}
}

func TestResidents(t *testing.T) {
	s, repo, _, _, _ := crisisSetup(t)
	if err := s.AddResident(ctx, "c", "u-gm", ResidentInput{CharacterID: "pc1", Role: "guard"}); err == nil {
		t.Error("PC accepted as resident")
	}
	if err := s.AddResident(ctx, "c", "u-gm", ResidentInput{CharacterID: "npc1", Role: "wizard"}); err == nil {
		t.Error("bad role accepted")
	}
	if err := s.AddResident(ctx, "c", "u-gm", ResidentInput{CharacterID: "npc1", Role: "medic", Note: "Cuida dos doentes"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddResident(ctx, "c", "u-gm", ResidentInput{CharacterID: "npc1", Role: "guard"}); err == nil {
		t.Error("duplicate resident accepted")
	}
	pop := repo.r.Population
	id := repo.r.Residents[0].ID
	_ = s.UpdateResident(ctx, "c", ResidentInput{ID: id, Role: "leader"})
	if repo.r.Residents[0].Name != "Dona Lia" || repo.r.Residents[0].Role != "leader" || repo.r.Population != pop {
		t.Errorf("residents = %+v", repo.r.Residents)
	}
	if err := s.RemoveResident(ctx, "c", id); err != nil || len(repo.r.Residents) != 0 {
		t.Errorf("remove = %v", err)
	}
}
