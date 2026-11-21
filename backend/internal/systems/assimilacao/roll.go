package assimilacao

import (
	"crypto/rand"
	"math/big"
)

// Attr is a named sheet value (instinct or aptitude) in a fixed order.
type Attr struct {
	Group string
	Key   string
	Value int
}

// Field returns the JSON path of the attribute.
func (a Attr) Field() string { return a.Group + "." + a.Key }

func instinctList(s Sheet) []Attr {
	i := s.Instincts
	return []Attr{
		{"instincts", "influence", i.Influence}, {"instincts", "perception", i.Perception},
		{"instincts", "potency", i.Potency}, {"instincts", "reaction", i.Reaction},
		{"instincts", "resolution", i.Resolution}, {"instincts", "sagacity", i.Sagacity},
	}
}

func aptitudeList(s Sheet) []Attr {
	k, p := s.Knowledge, s.Practices
	return []Attr{
		{"knowledge", "biology", k.Biology}, {"knowledge", "erudition", k.Erudition},
		{"knowledge", "engineering", k.Engineering}, {"knowledge", "geography", k.Geography},
		{"knowledge", "medicine", k.Medicine}, {"knowledge", "security", k.Security},
		{"practices", "weapons", p.Weapons}, {"practices", "athletics", p.Athletics},
		{"practices", "expression", p.Expression}, {"practices", "stealth", p.Stealth},
		{"practices", "crafting", p.Crafting}, {"practices", "survival", p.Survival},
	}
}

func lookup(list []Attr, key string) (int, bool) {
	for _, a := range list {
		if a.Key == key {
			return a.Value, true
		}
	}
	return 0, false
}

// RollRequest is what the player asks for.
type RollRequest struct {
	Instinct        string `json:"instinct"`
	Aptitude        string `json:"aptitude,omitempty"`
	SecondInstinct  string `json:"secondInstinct,omitempty"`
	Effort          bool   `json:"effort"`
	UseOccupation   bool   `json:"useOccupation"`
	UseMarkingEvent bool   `json:"useMarkingEvent"`
	BonusKept       int    `json:"bonusKept"`
	DifficultyD12   int    `json:"difficultyD12"`
}

// DieSpec is one die of the pool.
type DieSpec struct {
	Kind       string `json:"kind"`
	Difficulty bool   `json:"difficulty"`
}

// RollSpec is the assembled pool.
type RollSpec struct {
	Dice        []DieSpec `json:"dice"`
	KeptCount   int       `json:"keptCount"`
	Assimilated bool      `json:"assimilated"`
}

// BuildRoll validates the request against the sheet and assembles the pool.
func BuildRoll(s Sheet, req RollRequest) (RollSpec, error) {
	if err := checkRequest(s, req); err != nil {
		return RollSpec{}, err
	}
	instincts := instinctList(s)
	v1, _ := lookup(instincts, req.Instinct)
	spec := RollSpec{Assimilated: req.SecondInstinct != ""}
	spec.Dice = repeat(spec.Dice, poolKind(spec.Assimilated), v1, false)
	if spec.Assimilated {
		v2, _ := lookup(instincts, req.SecondInstinct)
		spec.Dice = repeat(spec.Dice, D12, v2, false)
	}
	if req.Aptitude != "" {
		va, _ := lookup(aptitudeList(s), req.Aptitude)
		spec.Dice = repeat(spec.Dice, D10, va, false)
	}
	pool := len(spec.Dice)
	spec.Dice = repeat(spec.Dice, D12, req.DifficultyD12, true)
	spec.KeptCount = keptCount(req, spec.Assimilated, pool)
	return spec, nil
}

func poolKind(assimilated bool) string {
	if assimilated {
		return D12
	}
	return D6
}

func repeat(dice []DieSpec, kind string, n int, difficulty bool) []DieSpec {
	for range n {
		dice = append(dice, DieSpec{Kind: kind, Difficulty: difficulty})
	}
	return dice
}

func keptCount(req RollRequest, assimilated bool, pool int) int {
	n := 1 + req.BonusKept
	for _, on := range []bool{req.Effort, req.UseOccupation, req.UseMarkingEvent} {
		if on {
			n++
		}
	}
	if assimilated {
		n = max(n, 2)
	}
	return min(n, pool)
}

func checkRequest(s Sheet, req RollRequest) error {
	if _, ok := lookup(instinctList(s), req.Instinct); !ok {
		return ruleErr("Instinto inválido.")
	}
	if req.SecondInstinct != "" {
		if _, ok := lookup(instinctList(s), req.SecondInstinct); !ok {
			return ruleErr("Segundo instinto inválido.")
		}
		if req.SecondInstinct == req.Instinct {
			return ruleErr("Os dois instintos da rolagem assimilada devem ser diferentes.")
		}
	}
	if req.Aptitude != "" {
		if _, ok := lookup(aptitudeList(s), req.Aptitude); !ok {
			return ruleErr("Aptidão inválida.")
		}
	}
	return checkModifiers(s, req)
}

func checkModifiers(s Sheet, req RollRequest) error {
	if req.Effort && s.TugOfWar.DeterminationPoints < 1 {
		return ruleErr("Sem pontos de Determinação para o Empenho.")
	}
	if req.BonusKept < 0 {
		return ruleErr("O bônus de dados mantidos não pode ser negativo.")
	}
	if req.DifficultyD12 < 0 || req.DifficultyD12 > 3 {
		return ruleErr("A dificuldade deve ser de 0 a 3 d12.")
	}
	return nil
}

// RuleError is a PT-BR, user-facing rule violation.
type RuleError string

func (e RuleError) Error() string { return string(e) }

func ruleErr(msg string) error { return RuleError(msg) }

// RNG abstracts randomness so tests can be deterministic.
type RNG interface {
	Intn(n int) int
}

// CryptoRNG uses crypto/rand.
type CryptoRNG struct{}

// Intn returns a uniform value in [0, n).
func (CryptoRNG) Intn(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return int(v.Int64())
}

// Die is a rolled die.
type Die struct {
	Kind       string `json:"kind"`
	Face       int    `json:"face"`
	Difficulty bool   `json:"difficulty"`
}

// Roll is the outcome of drawing every die in a spec.
type Roll struct {
	Dice []Die `json:"dice"`
}

// Resolve draws a face for each die of the spec.
func Resolve(spec RollSpec, rng RNG) Roll {
	dice := make([]Die, len(spec.Dice))
	for i, d := range spec.Dice {
		dice[i] = Die{Kind: d.Kind, Face: rng.Intn(Sides(d.Kind)) + 1, Difficulty: d.Difficulty}
	}
	return Roll{Dice: dice}
}

// RollCosts are the side effects of a roll. Occupation and marking event are
// once-per-session flags tracked by the campaign, not the sheet.
type RollCosts struct {
	DeterminationPoints int  `json:"determinationPoints"`
	OccupationUsed      bool `json:"occupationUsed"`
	MarkingEventUsed    bool `json:"markingEventUsed"`
}

// Costs returns what the request consumes.
func Costs(req RollRequest) RollCosts {
	c := RollCosts{OccupationUsed: req.UseOccupation, MarkingEventUsed: req.UseMarkingEvent}
	if req.Effort {
		c.DeterminationPoints = 1
	}
	return c
}

// ApplyCosts returns a copy of the sheet with sheet-level costs applied.
func ApplyCosts(s Sheet, req RollRequest) Sheet {
	s.TugOfWar.DeterminationPoints = max(0, s.TugOfWar.DeterminationPoints-Costs(req).DeterminationPoints)
	return s
}
