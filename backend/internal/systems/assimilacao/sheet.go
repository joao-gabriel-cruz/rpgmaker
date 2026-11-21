// Package assimilacao holds the pure rules of the Assimilação RPG system.
package assimilacao

import "encoding/json"

// Identity describes who the character is.
type Identity struct {
	Name               string   `json:"name"`
	PortraitURL        *string  `json:"portraitUrl"`
	Generation         string   `json:"generation"`
	Occupation         string   `json:"occupation"`
	MarkingEvent       string   `json:"markingEvent"`
	PersonalPurposes   []string `json:"personalPurposes"`
	CollectivePurposes []string `json:"collectivePurposes"`
}

// Instincts are the six base attributes.
type Instincts struct {
	Influence  int `json:"influence"`
	Perception int `json:"perception"`
	Potency    int `json:"potency"`
	Reaction   int `json:"reaction"`
	Resolution int `json:"resolution"`
	Sagacity   int `json:"sagacity"`
}

// Knowledge aptitudes.
type Knowledge struct {
	Biology     int `json:"biology"`
	Erudition   int `json:"erudition"`
	Engineering int `json:"engineering"`
	Geography   int `json:"geography"`
	Medicine    int `json:"medicine"`
	Security    int `json:"security"`
}

// Practices aptitudes.
type Practices struct {
	Weapons    int `json:"weapons"`
	Athletics  int `json:"athletics"`
	Expression int `json:"expression"`
	Stealth    int `json:"stealth"`
	Crafting   int `json:"crafting"`
	Survival   int `json:"survival"`
}

// TugOfWar is the Determination versus Assimilation track.
type TugOfWar struct {
	DeterminationLevel  int `json:"determinationLevel"`
	DeterminationPoints int `json:"determinationPoints"`
	AssimilationLevel   int `json:"assimilationLevel"`
	AssimilationPoints  int `json:"assimilationPoints"`
}

// Health stores only the raw damage; levels are derived.
type Health struct {
	Damage int `json:"damage"`
}

// Item is an inventory entry.
type Item struct {
	Name     string   `json:"name"`
	Quality  int      `json:"quality"`
	Scarcity int      `json:"scarcity"`
	Pressure int      `json:"pressure"`
	Traits   []string `json:"traits"`
	// Optional fields set when the item comes from the campaign compendium.
	ItemID          string          `json:"itemId,omitempty"`
	Description     string          `json:"description,omitempty"`
	Icon            string          `json:"icon,omitempty"`
	Color           string          `json:"color,omitempty"`
	Kind            string          `json:"kind,omitempty"`
	AttackAnimation string          `json:"attackAnimation,omitempty"`
	Effects         json.RawMessage `json:"effects,omitempty"`
}

// Inventory has body slots and backpack slots.
type Inventory struct {
	Body     []Item `json:"body"`
	Backpack []Item `json:"backpack"`
}

// Characteristic is a purchasable trait costing 2x its level in XP.
type Characteristic struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

// XP holds the total experience earned.
type XP struct {
	Total int `json:"total"`
}

// Assimilation is a free-form entry tied to a card suit.
type Assimilation struct {
	Name        string `json:"name"`
	Suit        string `json:"suit"`
	Description string `json:"description"`
}

// Sheet is the full character sheet (contract shared with the frontend).
type Sheet struct {
	Identity        Identity         `json:"identity"`
	Instincts       Instincts        `json:"instincts"`
	Knowledge       Knowledge        `json:"knowledge"`
	Practices       Practices        `json:"practices"`
	TugOfWar        TugOfWar         `json:"tugOfWar"`
	Health          Health           `json:"health"`
	Inventory       Inventory        `json:"inventory"`
	Characteristics []Characteristic `json:"characteristics"`
	XP              XP               `json:"xp"`
	Assimilations   []Assimilation   `json:"assimilations"`
	Notes           string           `json:"notes"`
	Conditions      []string         `json:"conditions,omitempty"`
}

// NewSheet returns a fresh sheet with initial values.
func NewSheet() Sheet {
	return Sheet{
		Identity: Identity{
			Generation:         "pre_collapse",
			PersonalPurposes:   []string{"", ""},
			CollectivePurposes: []string{"", ""},
		},
		Instincts:       Instincts{1, 1, 1, 1, 1, 1},
		TugOfWar:        TugOfWar{DeterminationLevel: 9, DeterminationPoints: 9, AssimilationLevel: 1},
		Inventory:       Inventory{Body: []Item{}, Backpack: []Item{}},
		Characteristics: []Characteristic{},
		XP:              XP{Total: 7},
		Assimilations:   []Assimilation{},
	}
}

// ParseSheet decodes JSON over the defaults, so missing fields keep initial values.
func ParseSheet(data []byte) (Sheet, error) {
	s := NewSheet()
	if err := json.Unmarshal(data, &s); err != nil {
		return NewSheet(), err
	}
	if s.Inventory.Body == nil {
		s.Inventory.Body = []Item{}
	}
	if s.Inventory.Backpack == nil {
		s.Inventory.Backpack = []Item{}
	}
	if s.Characteristics == nil {
		s.Characteristics = []Characteristic{}
	}
	if s.Assimilations == nil {
		s.Assimilations = []Assimilation{}
	}
	return s, nil
}
