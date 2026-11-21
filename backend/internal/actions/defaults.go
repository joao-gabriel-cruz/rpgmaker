package actions

func opt(instinct, aptitude, label string) RollOption {
	return RollOption{Instinct: instinct, Aptitude: aptitude, Label: label}
}

// defaults is the Assimilação catalog; a campaign may override any entry by id.
var defaults = []Action{
	{ID: "melee_attack", Name: "Atacar corpo a corpo", Icon: "swords", Category: CatCombat, Target: TargetToken, Range: 1,
		Animation: "melee_strike", Outcome: OutcomeDamage,
		Rolls: []RollOption{opt("potency", "weapons", "Potência + Armas"), opt("reaction", "weapons", "Reação + Armas")}},
	{ID: "ranged_attack", Name: "Atacar à distância", Icon: "crosshair", Category: CatCombat, Target: TargetToken, Range: 8,
		Animation: "ranged_shot", Outcome: OutcomeDamage,
		Rolls: []RollOption{opt("perception", "weapons", "Percepção + Armas")}},
	{ID: "perceive", Name: "Perceber", Icon: "eye", Category: CatExploration, Target: TargetNone,
		Animation: "look_around", Outcome: OutcomeLoot,
		Rolls: []RollOption{opt("perception", "survival", "Percepção + Sobrevivência"), opt("perception", "security", "Percepção + Segurança")}},
	{ID: "search", Name: "Vasculhar", Icon: "search", Category: CatExploration, Target: TargetObject, Range: 2,
		Animation: "search", Outcome: OutcomeLoot,
		Rolls: []RollOption{opt("sagacity", "survival", "Sagacidade + Sobrevivência"), opt("perception", "crafting", "Percepção + Manufaturas")}},
	{ID: "sneak", Name: "Esgueirar-se", Icon: "footprints", Category: CatExploration, Target: TargetNone,
		Animation: "sneak", Outcome: OutcomeNotice,
		Rolls: []RollOption{opt("reaction", "stealth", "Reação + Furtividade")}},
	{ID: "repair", Name: "Consertar", Icon: "hammer", Category: CatSurvival, Target: TargetObject, Range: 2,
		Animation: "repair", Outcome: OutcomeNotice,
		Rolls: []RollOption{opt("sagacity", "crafting", "Sagacidade + Manufaturas"), opt("sagacity", "engineering", "Sagacidade + Engenharia")}},
	{ID: "first_aid", Name: "Primeiros socorros", Icon: "heart-pulse", Category: CatSurvival, Target: TargetToken, Range: 1,
		Animation: "first_aid", Outcome: OutcomeHeal,
		Rolls: []RollOption{opt("sagacity", "medicine", "Sagacidade + Medicina")}},
	{ID: "rest", Name: "Descansar", Icon: "moon", Category: CatSurvival, Target: TargetNone,
		Animation: "rest", Outcome: OutcomeRest,
		Rolls: []RollOption{opt("resolution", "survival", "Resolução + Sobrevivência")}},
	{ID: "dance", Name: "Dançar", Icon: "music", Category: CatSocial, Target: TargetNone,
		Animation: "dance", Outcome: OutcomeMorale,
		Rolls: []RollOption{opt("influence", "expression", "Influência + Expressão")}},
	{ID: "persuade", Name: "Convencer", Icon: "message-circle", Category: CatSocial, Target: TargetToken, Range: 6,
		Animation: "persuade", Outcome: OutcomeNotice,
		Rolls: []RollOption{opt("influence", "expression", "Influência + Expressão")}},
	{ID: "intimidate", Name: "Intimidar", Icon: "skull", Category: CatSocial, Target: TargetToken, Range: 6,
		Animation: "intimidate", Outcome: OutcomeNotice,
		Rolls: []RollOption{opt("potency", "expression", "Potência + Expressão")}},
}

func defaultByID(id string) (Action, bool) {
	for _, a := range defaults {
		if a.ID == id {
			a.Enabled, a.Builtin = true, true
			return a, true
		}
	}
	return Action{}, false
}
