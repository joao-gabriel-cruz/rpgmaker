package mcp

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	maxNPCNameLen = 100
	maxNPCCount   = 10
	maxNPCThreats = 50
	maxNPCText    = 5000
	maxAptitude   = 99
)

var (
	instinctKeys  = []string{"influence", "perception", "potency", "reaction", "resolution", "sagacity"}
	knowledgeKeys = []string{"biology", "erudition", "engineering", "geography", "medicine", "security"}
	practiceKeys  = []string{"weapons", "athletics", "expression", "stealth", "crafting", "survival"}
)

type createNPCInput struct {
	Name          string         `json:"name" jsonschema:"NPC name, up to 100 characters"`
	Count         int            `json:"count,omitempty" jsonschema:"how many copies, 1-10 (default 1); names become 'Name 1'..'Name N'"`
	Instincts     map[string]int `json:"instincts,omitempty" jsonschema:"keys influence, perception, potency, reaction, resolution, sagacity; 0-99, default 1"`
	Knowledge     map[string]int `json:"knowledge,omitempty" jsonschema:"keys biology, erudition, engineering, geography, medicine, security; 0-99, default 0"`
	Practices     map[string]int `json:"practices,omitempty" jsonschema:"keys weapons, athletics, expression, stealth, crafting, survival; 0-99, default 0"`
	BoxesPerLevel *int           `json:"boxesPerLevel,omitempty" jsonschema:"health boxes per level, at least 1 (default 3)"`
	Threats       []string       `json:"threats,omitempty" jsonschema:"threat/ability descriptions shown on the NPC sheet, up to 50"`
	Notes         string         `json:"notes,omitempty" jsonschema:"free description, up to 5000 characters"`
	At            *Point         `json:"at,omitempty" jsonschema:"place the NPCs on the free cells nearest to this cell of the active scene"`
}

func validateGroup(label string, vals map[string]int, keys []string) error {
	for k, v := range vals {
		if !slices.Contains(keys, k) {
			return newUserError(fmt.Sprintf("%s: chave desconhecida %q; use %v", label, k, keys))
		}
		if v < 0 || v > maxAptitude {
			return newUserError(fmt.Sprintf("%s.%s: valores devem ficar entre 0 e 99", label, k))
		}
	}
	return nil
}

func validateCreateNPC(in createNPCInput) error {
	if n := utf8.RuneCountInString(strings.TrimSpace(in.Name)); n == 0 || n > maxNPCNameLen {
		return newUserError("o nome do NPC deve ter entre 1 e 100 caracteres")
	}
	if in.Count != 0 && (in.Count < 1 || in.Count > maxNPCCount) {
		return newUserError("a quantidade deve ficar entre 1 e 10")
	}
	for _, g := range []struct {
		label string
		vals  map[string]int
		keys  []string
	}{{"instincts", in.Instincts, instinctKeys}, {"knowledge", in.Knowledge, knowledgeKeys}, {"practices", in.Practices, practiceKeys}} {
		if err := validateGroup(g.label, g.vals, g.keys); err != nil {
			return err
		}
	}
	return validateNPCExtras(in)
}

func validateNPCExtras(in createNPCInput) error {
	if in.BoxesPerLevel != nil && *in.BoxesPerLevel < 1 {
		return newUserError("boxesPerLevel deve ser ao menos 1")
	}
	if len(in.Threats) > maxNPCThreats {
		return newUserError("no máximo 50 ameaças")
	}
	for _, th := range in.Threats {
		if len(th) > maxNPCText {
			return newUserError("texto de ameaça longo demais (máximo 5000 caracteres)")
		}
	}
	if len(in.Notes) > maxNPCText {
		return newUserError("descrição longa demais (máximo 5000 caracteres)")
	}
	return nil
}

func hasSheetFields(in createNPCInput) bool {
	return len(in.Instincts)+len(in.Knowledge)+len(in.Practices) > 0 ||
		in.BoxesPerLevel != nil || len(in.Threats) > 0 || in.Notes != ""
}

// mergeNPCSheet overlays only the given fields on the sheet Create returned, keeping unknown keys.
func mergeNPCSheet(base json.RawMessage, in createNPCInput) (json.RawMessage, error) {
	doc := map[string]any{}
	if err := json.Unmarshal(base, &doc); err != nil {
		return nil, fmt.Errorf("decode npc sheet: %w", err)
	}
	for field, vals := range map[string]map[string]int{"instincts": in.Instincts, "knowledge": in.Knowledge, "practices": in.Practices} {
		sub, _ := doc[field].(map[string]any)
		if sub == nil {
			sub = map[string]any{}
		}
		for k, v := range vals {
			sub[k] = v
		}
		doc[field] = sub
	}
	health, _ := doc["health"].(map[string]any)
	if health == nil {
		health = map[string]any{}
	}
	if in.BoxesPerLevel != nil {
		health["boxesPerLevel"] = *in.BoxesPerLevel
	}
	doc["health"] = health
	if len(in.Threats) > 0 {
		doc["threats"] = in.Threats
	}
	if in.Notes != "" {
		doc["notes"] = in.Notes
	}
	return json.Marshal(doc)
}
