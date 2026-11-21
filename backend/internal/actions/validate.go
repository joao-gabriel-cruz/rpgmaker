package actions

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	categories = map[string]bool{CatCombat: true, CatExploration: true, CatSocial: true, CatSurvival: true}
	targets    = map[string]bool{TargetNone: true, TargetCell: true, TargetToken: true, TargetObject: true}
	outcomes   = map[string]bool{OutcomeNone: true, OutcomeDamage: true, OutcomeHeal: true, OutcomeLoot: true,
		OutcomeRest: true, OutcomeMorale: true, OutcomeNotice: true}
	animations = map[string]bool{"melee_strike": true, "ranged_shot": true, "look_around": true, "search": true, "sneak": true,
		"repair": true, "first_aid": true, "rest": true, "dance": true, "persuade": true, "intimidate": true}
	instincts = map[string]bool{"influence": true, "perception": true, "potency": true, "reaction": true, "resolution": true, "sagacity": true}
	aptitudes = map[string]bool{"biology": true, "erudition": true, "engineering": true, "geography": true, "medicine": true,
		"security": true, "weapons": true, "athletics": true, "expression": true, "stealth": true, "crafting": true, "survival": true}
	iconRe = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)
)

type inputError string

func (e inputError) Error() string { return string(e) }

func bad(msg string) error { return inputError(msg) }

// Clean validates and normalizes an action coming from the Mestre; the id is assigned by the caller.
func Clean(in Action) (Action, error) {
	in.Name = strings.TrimSpace(in.Name)
	if n := utf8.RuneCountInString(in.Name); n < 1 || n > MaxName {
		return Action{}, bad("O nome da ação deve ter de 1 a 40 caracteres")
	}
	if !iconRe.MatchString(in.Icon) {
		return Action{}, bad("Ícone inválido (use o nome de um ícone lucide)")
	}
	switch {
	case !categories[in.Category]:
		return Action{}, bad("Categoria inválida")
	case !targets[in.Target]:
		return Action{}, bad("Alvo inválido")
	case !outcomes[in.Outcome]:
		return Action{}, bad("Consequência inválida")
	case !animations[in.Animation]:
		return Action{}, bad("Animação inválida")
	case in.Range < 0 || in.Range > MaxRange:
		return Action{}, bad("O alcance vai de 0 a 20 células")
	}
	rolls, err := cleanRolls(in.Rolls)
	in.Rolls = rolls
	return in, err
}

func cleanRolls(in []RollOption) ([]RollOption, error) {
	if len(in) < 1 || len(in) > MaxRolls {
		return nil, bad("Informe de 1 a 4 rolagens")
	}
	out := make([]RollOption, 0, len(in))
	for _, r := range in {
		if !instincts[r.Instinct] || (r.Aptitude != "" && !aptitudes[r.Aptitude]) {
			return nil, bad("Instinto ou aptidão inválidos na rolagem")
		}
		r.Label = strings.TrimSpace(r.Label)
		if r.Label == "" || utf8.RuneCountInString(r.Label) > MaxName {
			return nil, bad("Cada rolagem precisa de um rótulo de até 40 caracteres")
		}
		out = append(out, r)
	}
	return out, nil
}
