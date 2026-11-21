package assimilacao

import "testing"

var validateCases = []struct {
	name     string
	mutate   func(*Sheet)
	code     string
	field    string
	blocking bool
}{
	{"instinct max", func(s *Sheet) { s.Instincts.Potency = 4 }, "instinct_max", "instincts.potency", false},
	{"instinct min", func(s *Sheet) { s.Instincts.Reaction = 0 }, "instinct_min", "instincts.reaction", false},
	{"instinct negative", func(s *Sheet) { s.Instincts.Reaction = -1 }, "negative_value", "instincts.reaction", true},
	{"instincts total", func(s *Sheet) { s.Instincts = Instincts{3, 3, 3, 1, 1, 1} }, "instincts_total", "instincts", false},
	{"aptitude max", func(s *Sheet) { s.Practices.Weapons = 3 }, "aptitude_max", "practices.weapons", false},
	{"aptitude negative", func(s *Sheet) { s.Knowledge.Biology = -1 }, "negative_value", "knowledge.biology", true},
	{"aptitudes total", func(s *Sheet) { s.Knowledge = Knowledge{2, 2, 2, 2, 0, 0} }, "aptitudes_total", "aptitudes", false},
	{"tug total", func(s *Sheet) { s.TugOfWar.AssimilationLevel = 2 }, "tug_total", "tugOfWar", false},
	{"det points", func(s *Sheet) { s.TugOfWar.DeterminationPoints = 10 }, "determination_points", "tugOfWar.determinationPoints", false},
	{"assim points", func(s *Sheet) { s.TugOfWar.AssimilationPoints = 2 }, "assimilation_points", "tugOfWar.assimilationPoints", false},
	{"negative points", func(s *Sheet) { s.TugOfWar.DeterminationPoints = -1 }, "negative_value", "tugOfWar", true},
	{"negative damage", func(s *Sheet) { s.Health.Damage = -1 }, "negative_value", "health.damage", true},
	{"negative xp", func(s *Sheet) { s.XP.Total = -1 }, "negative_value", "xp.total", true},
	{"xp overspent", func(s *Sheet) { s.Characteristics = []Characteristic{{Level: 4}} }, "xp_overspent", "xp", false},
	{"char level", func(s *Sheet) { s.Characteristics = []Characteristic{{Level: 0}} }, "characteristic_level", "characteristics[0].level", true},
	{"body full", func(s *Sheet) { s.Inventory.Body = make([]Item, 4) }, "inventory_body_full", "inventory.body", true},
	{"backpack full", func(s *Sheet) { s.Inventory.Backpack = make([]Item, 7) }, "inventory_backpack_full", "inventory.backpack", true},
	{"suit", func(s *Sheet) { s.Assimilations = []Assimilation{{Suit: "x"}} }, "suit_invalid", "assimilations[0].suit", true},
	{"generation", func(s *Sheet) { s.Identity.Generation = "x" }, "generation_invalid", "identity.generation", true},
	{"personal purposes", func(s *Sheet) { s.Identity.PersonalPurposes = []string{"a"} }, "personal_purposes", "identity.personalPurposes", false},
	{"collective purposes", func(s *Sheet) { s.Identity.CollectivePurposes = nil }, "collective_purposes", "identity.collectivePurposes", false},
}

func TestValidate(t *testing.T) {
	for _, tc := range validateCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSheet()
			tc.mutate(&s)
			for _, is := range Validate(s) {
				if is.Code == tc.code && is.Field == tc.field {
					if is.Blocking != tc.blocking || is.Message == "" {
						t.Fatalf("bad issue: %+v", is)
					}
					return
				}
			}
			t.Fatalf("issue %s/%s not found in %+v", tc.code, tc.field, Validate(s))
		})
	}
}
