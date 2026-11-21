package items

import (
	"regexp"
	"strings"
)

// Icons is the curated list of lucide icon names a compendium item may use.
var Icons = []string{
	"sword", "swords", "slice", "crosshair", "target", "hammer", "axe", "wrench", "key", "backpack",
	"wine", "milk", "pill", "cross", "bandage", "flashlight", "map", "radio", "battery", "cable",
	"drumstick", "apple", "droplet", "trees", "gem", "book", "skull", "shield", "shirt", "flame",
	"package", "cog", "scissors", "tent", "lock", "bomb", "syringe", "briefcase", "bone", "leaf",
}

var (
	colorRe   = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	kinds     = map[string]bool{KindMelee: true, KindRanged: true, KindTool: true, KindConsumable: true, KindClothing: true, KindResource: true, KindKey: true, KindOther: true}
	anims     = map[string]bool{"slash": true, "blunt": true, "shot": true, "arrow": true}
	categorys = map[string]bool{"combat": true, "exploration": true, "social": true, "survival": true}
	resources = map[string]bool{"water": true, "food": true, "wood": true}
)

// inputError is a validation failure whose PT-BR message may be shown to the Mestre.
type inputError string

func (e inputError) Error() string { return string(e) }

func bad(msg string) error { return inputError(msg) }

func iconOK(icon string) bool {
	for _, i := range Icons {
		if i == icon {
			return true
		}
	}
	return false
}

func inRange(n, lo, hi int) bool { return n >= lo && n <= hi }

// Clean validates a compendium item sent by the Mestre and returns it normalized; errors are PT-BR.
func Clean(in Item) (Item, error) {
	in.Name, in.Description = strings.TrimSpace(in.Name), strings.TrimSpace(in.Description)
	switch {
	case in.Name == "" || len([]rune(in.Name)) > 80:
		return in, bad("O nome do item deve ter de 1 a 80 caracteres")
	case len([]rune(in.Description)) > 500:
		return in, bad("A descrição do item deve ter até 500 caracteres")
	case !iconOK(in.Icon):
		return in, bad("Ícone inválido")
	case !colorRe.MatchString(in.Color):
		return in, bad("Cor inválida")
	case !kinds[in.Kind]:
		return in, bad("Tipo de item inválido")
	case !inRange(in.Quality, 0, 10) || !inRange(in.Scarcity, 0, 10) || !inRange(in.Pressure, 0, 10):
		return in, bad("Qualidade, Escassez e Pressão vão de 0 a 10")
	}
	var err error
	if in.Traits, err = cleanTraits(in.Traits); err != nil {
		return in, err
	}
	if err := cleanAnimation(&in); err != nil {
		return in, err
	}
	return in, cleanEffects(in)
}

func cleanTraits(in []string) ([]string, error) {
	out := []string{}
	for _, t := range in {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	if len(out) > 10 {
		return nil, bad("No máximo 10 características por item")
	}
	for _, t := range out {
		if len([]rune(t)) > 40 {
			return nil, bad("Cada característica deve ter até 40 caracteres")
		}
	}
	return out, nil
}

func cleanAnimation(in *Item) error {
	if in.AttackAnimation == "" {
		return nil
	}
	if in.Kind != KindMelee && in.Kind != KindRanged {
		return bad("Só armas têm animação de ataque")
	}
	if !anims[in.AttackAnimation] {
		return bad("Animação de ataque inválida")
	}
	return nil
}

func cleanEffects(in Item) error {
	e := in.Effects
	if b := e.KeptBonus; b != nil && (!categorys[b.ActionCategory] || !inRange(b.Amount, 1, 2)) {
		return bad("O bônus de dado mantido vai de +1 a +2 em uma categoria de ação válida")
	}
	if e.Heal != 0 && !inRange(e.Heal, 1, 6) {
		return bad("A cura vai de 1 a 6")
	}
	if r := e.RefugeResource; r != nil && (!resources[r.Resource] || !inRange(r.Amount, 1, 10)) {
		return bad("Recurso do Refúgio inválido (água, alimento ou madeira, de 1 a 10)")
	}
	if (e.Heal != 0 || e.RefugeResource != nil) && in.Kind != KindConsumable {
		return bad("Apenas consumíveis curam ou dão recursos ao Refúgio")
	}
	return nil
}
