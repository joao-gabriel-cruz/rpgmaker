package assimilacao

import (
	"slices"
)

// Result is the final tally of a roll.
type Result struct {
	Symbols     Symbols `json:"symbols"`
	Kept        []int   `json:"kept"`
	Penalty     int     `json:"penalty"`
	Assimilated bool    `json:"assimilated"`
}

func better(a, b Symbols) bool {
	if a.Success != b.Success {
		return a.Success > b.Success
	}
	if a.Adaptation != b.Adaptation {
		return a.Adaptation > b.Adaptation
	}
	return a.Pressure < b.Pressure
}

// SuggestKept picks the non-difficulty dice maximizing Success, then Adaptation,
// then fewest Pressure. Lexicographic order is additive, so greedy top-k is optimal.
func SuggestKept(roll Roll, spec RollSpec) []int {
	var idx []int
	for i, d := range roll.Dice {
		if !d.Difficulty {
			idx = append(idx, i)
		}
	}
	sym := func(i int) Symbols { return FaceSymbols(roll.Dice[i].Kind, roll.Dice[i].Face) }
	slices.SortStableFunc(idx, func(a, b int) int {
		switch {
		case better(sym(a), sym(b)):
			return -1
		case better(sym(b), sym(a)):
			return 1
		}
		return 0
	})
	idx = idx[:min(spec.KeptCount, len(idx))]
	slices.Sort(idx)
	return idx
}

// Finalize validates the chosen dice and tallies the result. Difficulty dice
// always count (their symbols are summed, not chosen); decision to validate
// against the book. Pressure on kept dice always counts, assimilated or not.
func Finalize(roll Roll, spec RollSpec, kept []int, derived Derived) (Result, error) {
	if len(kept) != spec.KeptCount {
		return Result{}, ruleErr("Quantidade de dados mantidos incorreta.")
	}
	seen := map[int]bool{}
	var total Symbols
	for _, i := range kept {
		if i < 0 || i >= len(roll.Dice) || roll.Dice[i].Difficulty {
			return Result{}, ruleErr("Dado mantido inválido.")
		}
		if seen[i] {
			return Result{}, ruleErr("Dado mantido repetido.")
		}
		seen[i] = true
		total = total.Add(FaceSymbols(roll.Dice[i].Kind, roll.Dice[i].Face))
	}
	for _, d := range roll.Dice {
		if d.Difficulty {
			total = total.Add(FaceSymbols(d.Kind, d.Face))
		}
	}
	total.Success = max(0, total.Success-derived.SuccessPenalty)
	sorted := slices.Sorted(slices.Values(kept))
	return Result{Symbols: total, Kept: sorted, Penalty: derived.SuccessPenalty, Assimilated: spec.Assimilated}, nil
}
