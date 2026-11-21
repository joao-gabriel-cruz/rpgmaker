package assimilacao

import (
	"reflect"
	"testing"
)

func mkRoll(faces ...int) (Roll, RollSpec) {
	var r Roll
	for _, f := range faces {
		r.Dice = append(r.Dice, Die{Kind: D12, Face: f})
	}
	return r, RollSpec{}
}

func TestSuggestKept(t *testing.T) {
	roll, spec := mkRoll(12, 7, 6, 10, 8, 1)
	roll.Dice[1].Difficulty = true
	spec.KeptCount = 3
	// non-difficulty: idx0=12(P2) idx2=6(S1) idx3=10(S2P1) idx4=8(S1A1) idx5=1
	if got := SuggestKept(roll, spec); !reflect.DeepEqual(got, []int{2, 3, 4}) {
		t.Fatalf("got %v", got)
	}
	spec.KeptCount = 9
	if got := SuggestKept(roll, spec); len(got) != 5 {
		t.Fatalf("capped: %v", got)
	}
}

func TestSuggestKeptTieBreaks(t *testing.T) {
	roll, spec := mkRoll(6, 8, 9)
	spec.KeptCount = 1
	// 8 beats 9 (same S/A, less pressure) and 6 (adaptation)
	if got := SuggestKept(roll, spec); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("got %v", got)
	}
}

func TestFinalize(t *testing.T) {
	roll, spec := mkRoll(7, 8, 3, 10)
	roll.Dice[3].Difficulty = true
	spec.KeptCount = 2
	spec.Assimilated = true
	res, err := Finalize(roll, spec, []int{1, 0}, Derived{SuccessPenalty: 1})
	if err != nil {
		t.Fatal(err)
	}
	// kept 7 (S2) + 8 (S1A1) + difficulty 10 (S2P1) = S5 A1 P1, minus 1
	want := Result{Symbols: Symbols{4, 1, 1}, Kept: []int{0, 1}, Penalty: 1, Assimilated: true}
	if !reflect.DeepEqual(res, want) {
		t.Fatalf("got %+v want %+v", res, want)
	}
}

func TestFinalizePenaltyFloor(t *testing.T) {
	roll, spec := mkRoll(3)
	spec.KeptCount = 1
	res, err := Finalize(roll, spec, []int{0}, Derived{SuccessPenalty: 2})
	if err != nil || res.Symbols.Success != 0 || res.Symbols.Pressure != 1 {
		t.Fatalf("got %+v %v", res, err)
	}
}

func TestFinalizeErrors(t *testing.T) {
	roll, spec := mkRoll(7, 8, 9)
	roll.Dice[2].Difficulty = true
	spec.KeptCount = 2
	tests := map[string][]int{
		"too few": {0}, "too many": {0, 1, 1}, "difficulty": {0, 2},
		"out of range": {0, 5}, "negative": {0, -1}, "duplicate": {1, 1},
	}
	for name, kept := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Finalize(roll, spec, kept, Derived{}); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
