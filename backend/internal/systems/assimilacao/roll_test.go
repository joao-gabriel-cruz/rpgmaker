package assimilacao

import "testing"

type seqRNG struct {
	vals []int
	pos  int
}

func (r *seqRNG) Intn(n int) int {
	v := r.vals[r.pos%len(r.vals)] % n
	r.pos++
	return v
}

func count(spec RollSpec, kind string, difficulty bool) int {
	n := 0
	for _, d := range spec.Dice {
		if d.Kind == kind && d.Difficulty == difficulty {
			n++
		}
	}
	return n
}

func testSheet() Sheet {
	s := NewSheet()
	s.Instincts.Potency, s.Instincts.Reaction = 3, 2
	s.Knowledge.Biology = 2
	return s
}

func TestBuildRoll(t *testing.T) {
	tests := []struct {
		name                  string
		req                   RollRequest
		d6, d10, d12, diff, k int
		assimilated           bool
	}{
		{"normal", RollRequest{Instinct: "potency", Aptitude: "biology"}, 3, 2, 0, 0, 1, false},
		{"instinct only", RollRequest{Instinct: "potency"}, 3, 0, 0, 0, 1, false},
		{"modifiers", RollRequest{Instinct: "potency", Aptitude: "biology", Effort: true, UseOccupation: true, UseMarkingEvent: true, BonusKept: 1}, 3, 2, 0, 0, 5, false},
		{"kept capped", RollRequest{Instinct: "reaction", BonusKept: 9}, 2, 0, 0, 0, 2, false},
		{"difficulty", RollRequest{Instinct: "potency", DifficultyD12: 2}, 3, 0, 0, 2, 1, false},
		{"assimilated", RollRequest{Instinct: "potency", SecondInstinct: "reaction", Aptitude: "biology"}, 0, 2, 5, 0, 2, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := BuildRoll(testSheet(), tc.req)
			if err != nil {
				t.Fatal(err)
			}
			if count(spec, D6, false) != tc.d6 || count(spec, D10, false) != tc.d10 ||
				count(spec, D12, false) != tc.d12 || count(spec, D12, true) != tc.diff ||
				spec.KeptCount != tc.k || spec.Assimilated != tc.assimilated {
				t.Fatalf("unexpected spec: %+v", spec)
			}
		})
	}
}

func TestBuildRollErrors(t *testing.T) {
	noDet := testSheet()
	noDet.TugOfWar.DeterminationPoints = 0
	tests := []struct {
		name string
		s    Sheet
		req  RollRequest
	}{
		{"bad instinct", testSheet(), RollRequest{Instinct: "x"}},
		{"bad aptitude", testSheet(), RollRequest{Instinct: "potency", Aptitude: "x"}},
		{"bad second", testSheet(), RollRequest{Instinct: "potency", SecondInstinct: "x"}},
		{"same instinct", testSheet(), RollRequest{Instinct: "potency", SecondInstinct: "potency"}},
		{"effort without determination", noDet, RollRequest{Instinct: "potency", Effort: true}},
		{"difficulty high", testSheet(), RollRequest{Instinct: "potency", DifficultyD12: 4}},
		{"negative bonus", testSheet(), RollRequest{Instinct: "potency", BonusKept: -1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := BuildRoll(tc.s, tc.req); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestResolveDeterministic(t *testing.T) {
	spec := RollSpec{Dice: []DieSpec{{D6, false}, {D10, false}, {D12, true}}}
	roll := Resolve(spec, &seqRNG{vals: []int{5, 9, 11}})
	want := []Die{{D6, 6, false}, {D10, 10, false}, {D12, 12, true}}
	for i, d := range want {
		if roll.Dice[i] != d {
			t.Fatalf("die %d: got %+v want %+v", i, roll.Dice[i], d)
		}
	}
}

func TestCryptoRNGRange(t *testing.T) {
	var r CryptoRNG
	for range 200 {
		if v := r.Intn(6); v < 0 || v > 5 {
			t.Fatalf("out of range: %d", v)
		}
	}
}

func TestCosts(t *testing.T) {
	req := RollRequest{Instinct: "potency", Effort: true, UseOccupation: true}
	if c := Costs(req); c != (RollCosts{1, true, false}) {
		t.Fatalf("costs: %+v", c)
	}
	s := testSheet()
	if got := ApplyCosts(s, req).TugOfWar.DeterminationPoints; got != 8 || s.TugOfWar.DeterminationPoints != 9 {
		t.Fatalf("apply: %d", got)
	}
	if got := ApplyCosts(s, RollRequest{}).TugOfWar.DeterminationPoints; got != 9 {
		t.Fatalf("no effort: %d", got)
	}
}
