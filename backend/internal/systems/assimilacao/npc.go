package assimilacao

import (
	"encoding/json"
	"fmt"
)

// NPC sheet bounds; structural only, NPCs have no creation limits.
const (
	maxNPCName     = 100
	maxNPCThreats  = 50
	maxNPCText     = 5000
	maxNPCAptitude = 99
)

// NPCHealth is damage plus the manually set boxes per level (NPCs ignore 1+Potency+Resolution).
type NPCHealth struct {
	Damage        int `json:"damage"`
	BoxesPerLevel int `json:"boxesPerLevel"`
}

// NPCIdentity is the NPC name and portrait.
type NPCIdentity struct {
	Name        string  `json:"name"`
	PortraitURL *string `json:"portraitUrl"`
}

// NPCSheet is the simplified NPC sheet. Instincts, Knowledge and Practices share the PC keys so
// ParseSheet accepts this JSON and BuildRoll works on NPCs unchanged.
type NPCSheet struct {
	Identity  NPCIdentity `json:"identity"`
	Instincts Instincts   `json:"instincts"`
	Knowledge Knowledge   `json:"knowledge"`
	Practices Practices   `json:"practices"`
	Health    NPCHealth   `json:"health"`
	Threats   []string    `json:"threats"`
	Notes     string      `json:"notes"`
}

// NewNPCSheet returns a blank NPC sheet.
func NewNPCSheet() NPCSheet {
	return NPCSheet{
		Instincts: Instincts{1, 1, 1, 1, 1, 1},
		Health:    NPCHealth{BoxesPerLevel: 3},
		Threats:   []string{},
	}
}

// ParseNPCSheet decodes JSON over the NPC defaults.
func ParseNPCSheet(data []byte) (NPCSheet, error) {
	s := NewNPCSheet()
	if err := json.Unmarshal(data, &s); err != nil {
		return NewNPCSheet(), err
	}
	if s.Threats == nil {
		s.Threats = []string{}
	}
	return s, nil
}

// ValidateNPC reports structural problems only; every issue is blocking.
func ValidateNPC(s NPCSheet) []Issue {
	out := []Issue{}
	add := func(code, field, msg string) {
		out = append(out, Issue{Code: code, Field: field, Message: msg, Blocking: true})
	}
	if len(s.Identity.Name) > maxNPCName {
		add("name_too_long", "identity.name", "Nome longo demais")
	}
	if len(s.Notes) > maxNPCText {
		add("notes_too_long", "notes", "Descrição longa demais")
	}
	if len(s.Threats) > maxNPCThreats {
		add("too_many_threats", "threats", "Ameaças demais")
	}
	for i, t := range s.Threats {
		if len(t) > maxNPCText {
			add("threat_too_long", fmt.Sprintf("threats.%d", i), "Texto longo demais")
		}
	}
	if s.Health.BoxesPerLevel < 1 {
		add("boxes_invalid", "health.boxesPerLevel", "Caixas por nível deve ser ao menos 1")
	}
	if s.Health.Damage < 0 {
		add("damage_negative", "health.damage", "Dano não pode ser negativo")
	}
	return append(out, npcAptitudeIssues(s)...)
}

func npcAptitudeIssues(s NPCSheet) []Issue {
	out := []Issue{}
	groups := map[string][]int{
		"instincts": {s.Instincts.Influence, s.Instincts.Perception, s.Instincts.Potency, s.Instincts.Reaction, s.Instincts.Resolution, s.Instincts.Sagacity},
		"knowledge": {s.Knowledge.Biology, s.Knowledge.Erudition, s.Knowledge.Engineering, s.Knowledge.Geography, s.Knowledge.Medicine, s.Knowledge.Security},
		"practices": {s.Practices.Weapons, s.Practices.Athletics, s.Practices.Expression, s.Practices.Stealth, s.Practices.Crafting, s.Practices.Survival},
	}
	for field, vals := range groups {
		for _, v := range vals {
			if v < 0 || v > maxNPCAptitude {
				out = append(out, Issue{Code: "value_range", Field: field, Message: "Valores devem ficar entre 0 e 99", Blocking: true})
				break
			}
		}
	}
	return out
}

// DeriveNPC computes the health level from the manual boxes per level.
func DeriveNPC(s NPCSheet) Derived {
	boxes := max(s.Health.BoxesPerLevel, 1)
	level := min(max(s.Health.Damage, 0)/boxes, len(levelNames)-1)
	return Derived{
		BoxesPerLevel: boxes, CurrentLevel: level, LevelName: levelNames[level],
		SuccessPenalty: levelPenalty[level], Debilitated: level >= 4, Incapacitated: level >= 5,
	}
}
