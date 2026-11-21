package actions

import "encoding/json"

type sheetItem struct {
	Kind            string `json:"kind"`
	AttackAnimation string `json:"attackAnimation"`
	Effects         struct {
		KeptBonus *struct {
			ActionCategory string `json:"actionCategory"`
			Amount         int    `json:"amount"`
		} `json:"keptBonus"`
	} `json:"effects"`
}

// weaponFor reads the equipped (body) items: the best kept-dice bonus for the action's category and
// the attack animation of the first weapon matching the action's style.
func weaponFor(sheet json.RawMessage, act Action) (bonus int, variant string) {
	var doc struct {
		Inventory struct {
			Body []sheetItem `json:"body"`
		} `json:"inventory"`
	}
	if json.Unmarshal(sheet, &doc) != nil {
		return 0, ""
	}
	want := map[string]string{"melee_strike": "melee", "ranged_shot": "ranged"}[act.Animation]
	for _, it := range doc.Inventory.Body {
		if b := it.Effects.KeptBonus; b != nil && b.ActionCategory == act.Category {
			bonus = max(bonus, b.Amount)
		}
		if variant == "" && want != "" && it.Kind == want {
			variant = it.AttackAnimation
		}
	}
	return min(bonus, 2), variant
}
