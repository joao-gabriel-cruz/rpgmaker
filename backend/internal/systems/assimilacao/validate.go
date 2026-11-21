package assimilacao

import "fmt"

// Inventory limits.
const (
	BodySlots     = 3
	BackpackSlots = 6
)

// Issue is a validation finding. Blocking issues are structural and always reject
// the sheet; non-blocking ones are creation limits the campaign may only warn about.
type Issue struct {
	Code     string `json:"code"`
	Field    string `json:"field"`
	Message  string `json:"message"`
	Blocking bool   `json:"blocking"`
}

func issue(blocking bool, code, field, msg string, args ...any) Issue {
	return Issue{Code: code, Field: field, Message: fmt.Sprintf(msg, args...), Blocking: blocking}
}

// Validate checks structural rules and creation limits.
func Validate(s Sheet) []Issue {
	var out []Issue
	out = append(out, structural(s)...)
	out = append(out, creationInstincts(s)...)
	out = append(out, creationAptitudes(s)...)
	out = append(out, creationMisc(s)...)
	return out
}

func structural(s Sheet) []Issue {
	var out []Issue
	for _, a := range append(instinctList(s), aptitudeList(s)...) {
		if a.Value < 0 {
			out = append(out, issue(true, "negative_value", a.Field(), "O valor não pode ser negativo."))
		}
	}
	return append(out, structuralMisc(s)...)
}

func structuralMisc(s Sheet) []Issue {
	var out []Issue
	if s.Health.Damage < 0 {
		out = append(out, issue(true, "negative_value", "health.damage", "O dano não pode ser negativo."))
	}
	if s.XP.Total < 0 {
		out = append(out, issue(true, "negative_value", "xp.total", "O XP não pode ser negativo."))
	}
	if s.TugOfWar.DeterminationPoints < 0 || s.TugOfWar.AssimilationPoints < 0 {
		out = append(out, issue(true, "negative_value", "tugOfWar", "Os pontos não podem ser negativos."))
	}
	if len(s.Inventory.Body) > BodySlots {
		out = append(out, issue(true, "inventory_body_full", "inventory.body", "O corpo comporta no máximo %d itens.", BodySlots))
	}
	if len(s.Inventory.Backpack) > BackpackSlots {
		out = append(out, issue(true, "inventory_backpack_full", "inventory.backpack", "A mochila comporta no máximo %d itens.", BackpackSlots))
	}
	for i, a := range s.Assimilations {
		if !validSuit(a.Suit) {
			out = append(out, issue(true, "suit_invalid", fmt.Sprintf("assimilations[%d].suit", i), "Naipe inválido."))
		}
	}
	return append(out, structuralChars(s)...)
}

func structuralChars(s Sheet) []Issue {
	var out []Issue
	switch s.Identity.Generation {
	case "pre_collapse", "during_collapse", "post_collapse":
	default:
		out = append(out, issue(true, "generation_invalid", "identity.generation", "Geração inválida."))
	}
	for i, c := range s.Characteristics {
		if c.Level < 1 {
			out = append(out, issue(true, "characteristic_level", fmt.Sprintf("characteristics[%d].level", i), "O nível da característica deve ser ao menos 1."))
		}
	}
	return out
}

func validSuit(v string) bool {
	return v == "hearts" || v == "diamonds" || v == "spades" || v == "clubs"
}

func creationInstincts(s Sheet) []Issue {
	var out []Issue
	sum := 0
	for _, a := range instinctList(s) {
		v := a.Value
		sum += v
		if v >= 0 && v < 1 {
			out = append(out, issue(false, "instinct_min", a.Field(), "O instinto deve ser ao menos 1."))
		}
		if v > 3 {
			out = append(out, issue(false, "instinct_max", a.Field(), "O instinto deve ser no máximo 3."))
		}
	}
	if sum > 9 {
		out = append(out, issue(false, "instincts_total", "instincts", "A soma dos instintos deve ser no máximo 9 (atual: %d).", sum))
	}
	return out
}

func creationAptitudes(s Sheet) []Issue {
	var out []Issue
	sum := 0
	for _, a := range aptitudeList(s) {
		sum += a.Value
		if a.Value > 2 {
			out = append(out, issue(false, "aptitude_max", a.Field(), "A aptidão deve ser no máximo 2."))
		}
	}
	if sum > 7 {
		out = append(out, issue(false, "aptitudes_total", "aptitudes", "A soma das aptidões deve ser no máximo 7 (atual: %d).", sum))
	}
	return out
}

func creationMisc(s Sheet) []Issue {
	var out []Issue
	t := s.TugOfWar
	if t.DeterminationLevel+t.AssimilationLevel != 10 {
		out = append(out, issue(false, "tug_total", "tugOfWar", "Determinação e Assimilação devem somar 10."))
	}
	if t.DeterminationPoints > t.DeterminationLevel {
		out = append(out, issue(false, "determination_points", "tugOfWar.determinationPoints", "Os pontos de Determinação não podem superar o nível."))
	}
	if t.AssimilationPoints > t.AssimilationLevel {
		out = append(out, issue(false, "assimilation_points", "tugOfWar.assimilationPoints", "Os pontos de Assimilação não podem superar o nível."))
	}
	if d := Derive(s); d.XPSpent > s.XP.Total {
		out = append(out, issue(false, "xp_overspent", "xp", "O XP gasto (%d) supera o total (%d).", d.XPSpent, s.XP.Total))
	}
	if len(s.Identity.PersonalPurposes) != 2 {
		out = append(out, issue(false, "personal_purposes", "identity.personalPurposes", "São necessários exatamente 2 propósitos pessoais."))
	}
	if len(s.Identity.CollectivePurposes) != 2 {
		out = append(out, issue(false, "collective_purposes", "identity.collectivePurposes", "São necessários exatamente 2 propósitos coletivos."))
	}
	return out
}
