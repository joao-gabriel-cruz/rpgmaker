package refuge

import (
	"errors"
	"testing"
)

// weekRefuge is a refuge with population level 6 (5 workers), roomy reserves and a safe region.
func weekRefuge(bs ...Building) Refuge {
	s := DefaultSheet()
	s.Population = Population{Level: 6, Max: 10}
	s.ReservesMax = 500
	s.Resources["water"], s.Resources["food"], s.Resources["wood"] = 100, 100, 100
	s.Region.Abundance["minerals"], s.Region.Abundance["biomass"], s.Region.Abundance["water"] = "abundant", "scarce", "medium"
	return Refuge{CampaignID: "c", Editable: Editable{Text: Text{Name: "R"}, Sheet: s}, Buildings: bs}
}

func done(kind string, workers int) Building {
	return Building{ID: kind, Kind: kind, Level: 1, Built: 1, State: StateDone, Workers: workers}
}

func building(kind string, workers int) Building {
	return Building{ID: kind, Kind: kind, Level: 1, State: StateBuilding, Workers: workers}
}

func TestExtractionFollowsAbundance(t *testing.T) {
	for ab, want := range map[string]int{"none": 0, "scarce": 1, "medium": 3, "abundant": 6} {
		r := weekRefuge(done(KindQuarry, 3))
		r.Region.Abundance["minerals"] = ab
		if ab == "abundant" {
			r.Buildings[0].Workers = 3
		}
		_, rep := Simulate(r, nil)
		if got := rep.Entered["minerals"]; got != want {
			t.Errorf("%s: minerals = %d, want %d", ab, got, want)
		}
	}
}

func TestExtractionBuildings(t *testing.T) {
	r := weekRefuge(done(KindBiomass, 4), done(KindFountain, 2))
	r.Region.Abundance["biomass"] = "abundant"
	_, rep := Simulate(r, nil)
	if rep.Entered["biomass"] != 8 || rep.Entered["water"] != 2 {
		t.Errorf("entered = %v", rep.Entered)
	}
}

func TestConversions(t *testing.T) {
	cases := []struct{ kind, in, out string }{
		{KindApothecary, "plants", "medicine"}, {KindAmmoFactory, "minerals", "ammo"}, {KindWorksFactory, "minerals", "buildingMaterial"},
		{KindMill, "plants", "food"}, {KindSawmill, "wood", "buildingMaterial"}, {KindRefineryVeg, "plants", "fuel"},
		{KindRefineryAnim, "animals", "fuel"}, {KindRefineryBio, "biomass", "fuel"},
	}
	for _, c := range cases {
		r := weekRefuge(done(c.kind, 4))
		r.Resources[c.in] += 50
		before := r.Resources[c.out]
		next, rep := Simulate(r, nil)
		if next.Resources[c.in] != r.Resources[c.in]-4-consumed(c.in) || rep.Entered[c.out] != 4 {
			t.Errorf("%s: in=%d entered=%v", c.kind, next.Resources[c.in], rep.Entered)
		}
		if got := next.Resources[c.out] - before + consumed(c.out); got != 4 {
			t.Errorf("%s: output stock +%d", c.kind, got)
		}
	}
}

// consumed is the weekly consumption of level 6 for the resource.
func consumed(k string) int {
	if k == "water" || k == "food" || k == "wood" {
		return 6
	}
	return 0
}

func TestConversionLimitedByInputAndRefineryCap(t *testing.T) {
	r := weekRefuge(done(KindApothecary, 4))
	r.Resources["plants"] = 2
	_, rep := Simulate(r, nil)
	if rep.Entered["medicine"] != 2 {
		t.Errorf("limited by input: %v", rep.Entered)
	}
	r = weekRefuge(done(KindRefineryVeg, 5))
	r.Population = Population{Level: 20, Max: 20}
	r.Buildings[0].Workers = 15
	r.Resources["plants"] = 99
	_, rep = Simulate(r, nil)
	if rep.Entered["fuel"] != RefineryWeekly {
		t.Errorf("refinery cap: %v", rep.Entered)
	}
}

func TestCanteenAndTailorDouble(t *testing.T) {
	r := weekRefuge(done(KindMill, 3), done(KindCanteen, 0), done(KindTailor, 2))
	r.Resources["plants"] = 20
	_, rep := Simulate(r, nil)
	if rep.Entered["food"] != 6 || rep.Entered["clothing"] != 4 {
		t.Errorf("entered = %v", rep.Entered)
	}
}

func TestOnlyDoneBuildingsProduce(t *testing.T) {
	r := weekRefuge(building(KindQuarry, 2))
	_, rep := Simulate(r, nil)
	if rep.Entered["minerals"] != 0 {
		t.Errorf("entered = %v", rep.Entered)
	}
}

func TestWorksCompleteAndApplyEffects(t *testing.T) {
	cases := map[string]func(Sheet) bool{
		KindCommunity: func(s Sheet) bool { return s.Morale == 1 },
		KindFortify:   func(s Sheet) bool { return s.Defense == 1 },
		KindDormitory: func(s Sheet) bool { return s.Population.Max == 10+DormitoryPopulation },
		KindReserves:  func(s Sheet) bool { return s.ReservesMax == 500+ExpandedReserves },
		KindPort:      func(s Sheet) bool { return s.Mobility == PortMobility },
		KindFountain:  func(s Sheet) bool { return s.HasWaterSource },
	}
	for kind, ok := range cases {
		d, _ := DefOf(kind)
		b := building(kind, 5)
		b.Progress, b.Paid = d.CostAt(1)-5, true
		next, rep := Simulate(weekRefuge(b), nil)
		nb := next.Buildings[0]
		if nb.State != StateDone || nb.Built != 1 || !ok(next.Sheet) || len(rep.Completed) != 1 || !rep.Works[0].Done {
			t.Errorf("%s: %+v %+v %v", kind, nb, next.Sheet, rep.Completed)
		}
	}
}

func TestWorkProgressAccumulatesAndPaysMaterials(t *testing.T) {
	r := weekRefuge(building(KindQuarry, 5))
	r.Resources["buildingMaterial"] = 5
	next, rep := Simulate(r, nil)
	b := next.Buildings[0]
	if b.Progress != 5 || !b.Paid || b.State != StateBuilding || rep.Spent["buildingMaterial"] != 2 || rep.Works[0].Points != 5 {
		t.Errorf("week 1: %+v %v", b, rep.Spent)
	}
	next.Resources["buildingMaterial"] = 0
	next, _ = Simulate(next, nil)
	if next.Buildings[0].Progress != 10 || next.Resources["buildingMaterial"] != 0 {
		t.Errorf("paid only once: %+v", next.Buildings[0])
	}
}

func TestWoodStandsInForMissingMaterial(t *testing.T) {
	r := weekRefuge(building(KindQuarry, 1))
	r.Resources["buildingMaterial"] = 1
	_, rep := Simulate(r, nil)
	if rep.Spent["buildingMaterial"] != 1 || rep.Spent["wood"] != WoodPerMaterial+6 {
		t.Errorf("spent = %v", rep.Spent)
	}
	r.Resources["wood"] = 1
	next, rep := Simulate(r, nil)
	if next.Buildings[0].Progress != 0 || rep.Works[0].Blocked == "" {
		t.Errorf("should be blocked: %+v %+v", next.Buildings[0], rep.Works)
	}
}

func TestLeveledCostAndUpgrade(t *testing.T) {
	r := weekRefuge(done(KindFortify, 1))
	if err := r.upgrade(KindFortify); err != nil {
		t.Fatal(err)
	}
	b := r.Buildings[0]
	d, _ := DefOf(KindFortify)
	if b.Level != 2 || b.State != StatePlanned || d.CostAt(2) != 20 || d.MaterialCost(2) != 2 {
		t.Errorf("upgrade = %+v", b)
	}
	r.Defense = 1
	r.Buildings[0].Workers, r.Buildings[0].State, r.Buildings[0].Progress, r.Buildings[0].Paid = 5, StateBuilding, 15, true
	next, _ := Simulate(r, nil)
	if next.Defense != 2 || next.Buildings[0].Built != 2 {
		t.Errorf("level 2 done: %+v", next.Sheet)
	}
	if !errors.Is((&Refuge{Buildings: []Building{done(KindQuarry, 0)}}).upgrade(KindQuarry), errNotLevel) {
		t.Error("non leveled must not upgrade")
	}
}

func TestWeeklyConsumptionWithAndWithoutWaterSource(t *testing.T) {
	r := weekRefuge()
	next, rep := Simulate(r, nil)
	if next.Resources["water"] != 94 || next.Resources["food"] != 94 || next.Resources["wood"] != 94 || next.Week != 2 || rep.Week != 1 {
		t.Errorf("without source: %v week %d", next.Resources, next.Week)
	}
	r.HasWaterSource = true
	next, _ = Simulate(r, nil)
	if next.Resources["water"] != 100 || next.Resources["food"] != 94 {
		t.Errorf("with source: %v", next.Resources)
	}
}

func TestYearlyAndBiennialMilestones(t *testing.T) {
	for week, want := range map[int][2]int{51: {0, 0}, 52: {6, 0}, 104: {6, 6}} {
		r := weekRefuge()
		r.Week, r.Resources["medicine"], r.Resources["clothing"] = week, 10, 10
		next, _ := Simulate(r, nil)
		if 10-next.Resources["medicine"] != want[0] || 10-next.Resources["clothing"] != want[1] {
			t.Errorf("week %d: %v", week, next.Resources)
		}
	}
}

func TestShortageBecomesTriggerAndStockNeverNegative(t *testing.T) {
	r := weekRefuge()
	r.Resources["water"] = 1
	next, rep := Simulate(r, nil)
	if next.Resources["water"] != 0 || rep.Missing["water"] != 5 {
		t.Errorf("missing = %v", rep.Missing)
	}
	if len(rep.Triggers) != 1 || rep.Triggers[0] != (CrisisTrigger{Kind: "shortage", Resource: "water", Detail: "Faltou Água (precisava 6, tinha 1)"}) {
		t.Errorf("triggers = %+v", rep.Triggers)
	}
}

func TestReservesCapCutsExcess(t *testing.T) {
	r := weekRefuge(done(KindQuarry, 4))
	r.ReservesMax = 280
	_, rep := Simulate(r, nil)
	if rep.Lost["minerals"]+rep.Lost["water"]+rep.Lost["food"]+rep.Lost["wood"] == 0 {
		t.Errorf("lost = %v", rep.Lost)
	}
	next, _ := Simulate(r, nil)
	if next.ReservesTotal() > 280 {
		t.Errorf("total = %d", next.ReservesTotal())
	}
}

func TestBreachAndOverpopulationTriggers(t *testing.T) {
	r := weekRefuge()
	r.Region.Danger, r.Defense = 4, 3
	_, rep := Simulate(r, nil)
	if len(rep.Triggers) != 1 || rep.Triggers[0].Kind != "breach" {
		t.Errorf("breach = %+v", rep.Triggers)
	}
	r.Population.Level = 11
	_, rep = Simulate(r, nil)
	if len(rep.Triggers) != 2 || rep.Triggers[0].Kind != "overpopulation" {
		t.Errorf("overpop = %+v", rep.Triggers)
	}
}

func TestFortificationNeedsAGuard(t *testing.T) {
	r := weekRefuge(done(KindFortify, 0))
	r.Defense, r.Region.Danger = 2, 2
	_, rep := Simulate(r, nil)
	if len(rep.Triggers) != 1 || rep.Triggers[0].Kind != "breach" || len(rep.Notes) == 0 {
		t.Errorf("unguarded: %+v %v", rep.Triggers, rep.Notes)
	}
	r.Buildings[0].Workers = 1
	_, rep = Simulate(r, nil)
	if len(rep.Triggers) != 0 {
		t.Errorf("guarded: %+v", rep.Triggers)
	}
}

func TestPortUpkeepEvery26Weeks(t *testing.T) {
	r := weekRefuge(done(KindPort, 0))
	r.Week = 26
	_, rep := Simulate(r, nil)
	if rep.Spent["wood"] != 6+PortUpkeepWood {
		t.Errorf("spent = %v", rep.Spent)
	}
}

func TestAdjustmentsApplyBeforeConsumption(t *testing.T) {
	r := weekRefuge()
	r.Resources["water"] = 2
	next, rep := Simulate(r, map[string]int{"water": 4, "food": -1000})
	if next.Resources["water"] != 0 || rep.Missing["water"] != 0 || rep.Adjustments["water"] != 4 || next.Resources["food"] != 0 {
		t.Errorf("adjust = %v %v", next.Resources, rep.Missing)
	}
	if ValidateAdjust(map[string]int{"gold": 1}) == nil || ValidateAdjust(map[string]int{"water": 99999}) == nil {
		t.Error("invalid adjust accepted")
	}
}
