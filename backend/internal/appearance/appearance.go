// Package appearance validates the optional 3D miniature description stored in a character sheet.
package appearance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

// Height and infection bounds.
const (
	MinHeight   = 0.9
	MaxHeight   = 1.1
	MaxInfected = 10
	maxBytes    = 4096
)

var colorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

var enums = map[string][]string{
	"body.build":           {"slim", "average", "broad"},
	"hair.style":           {"none", "short", "long", "ponytail", "mohawk", "curly"},
	"beard":                {"none", "stubble", "full"},
	"outfit.top":           {"tshirt", "jacket", "hoodie", "vest", "coat"},
	"outfit.bottom":        {"pants", "shorts", "skirt"},
	"outfit.shoes":         {"boots", "sneakers", "barefoot"},
	"accessories.weapon":   {"none", "pipe", "axe", "rifle", "bow", "knife"},
	"accessories.mask":     {"none", "gas", "bandana"},
	"accessories.headwear": {"none", "cap", "hood", "helmet"},
}

var colors = []string{
	"body.skin", "hair.color", "outfit.topColor", "outfit.bottomColor", "outfit.shoesColor", "tokenColor",
}

var bools = []string{"accessories.backpack", "accessories.glasses"}

func bad(field, msg string) systems.Issue {
	return systems.Issue{Code: "appearance_invalid", Severity: systems.SeverityBlocking, Field: "appearance." + field, Message: msg}
}

func lookup(root map[string]any, path string) (any, bool) {
	var cur any = root
	start := 0
	for i := 0; i <= len(path); i++ {
		if i < len(path) && path[i] != '.' {
			continue
		}
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[path[start:i]]
		if !ok {
			return nil, false
		}
		start = i + 1
	}
	return cur, true
}

// Validate checks the "appearance" block of a sheet; a sheet without it is valid.
func Validate(sheet json.RawMessage) []systems.Issue {
	var top struct {
		Appearance json.RawMessage `json:"appearance"`
	}
	if json.Unmarshal(sheet, &top) != nil || len(top.Appearance) == 0 || bytes.Equal(top.Appearance, []byte("null")) {
		return nil
	}
	if len(top.Appearance) > maxBytes {
		return []systems.Issue{bad("", "Aparência grande demais.")}
	}
	var root map[string]any
	if json.Unmarshal(top.Appearance, &root) != nil {
		return []systems.Issue{bad("", "Aparência inválida.")}
	}
	var out []systems.Issue
	out = append(out, checkEnums(root)...)
	out = append(out, checkColors(root)...)
	out = append(out, checkMisc(root)...)
	return out
}

func checkEnums(root map[string]any) []systems.Issue {
	var out []systems.Issue
	for path, allowed := range enums {
		v, ok := lookup(root, path)
		if !ok {
			continue
		}
		s, isStr := v.(string)
		if !isStr || !contains(allowed, s) {
			out = append(out, bad(path, "Opção de aparência desconhecida."))
		}
	}
	return out
}

func checkColors(root map[string]any) []systems.Issue {
	var out []systems.Issue
	for _, path := range colors {
		if v, ok := lookup(root, path); ok {
			if s, isStr := v.(string); !isStr || !colorRe.MatchString(s) {
				out = append(out, bad(path, "Cor inválida; use o formato #rrggbb."))
			}
		}
	}
	return out
}

func checkMisc(root map[string]any) []systems.Issue {
	var out []systems.Issue
	if v, ok := lookup(root, "body.height"); ok {
		if f, isNum := v.(float64); !isNum || f < MinHeight || f > MaxHeight {
			out = append(out, bad("body.height", fmt.Sprintf("A altura deve ficar entre %.1f e %.1f.", MinHeight, MaxHeight)))
		}
	}
	if v, ok := lookup(root, "infected"); ok {
		if f, isNum := v.(float64); !isNum || f < 0 || f > MaxInfected || f != float64(int(f)) {
			out = append(out, bad("infected", "O nível de infecção deve ser um inteiro de 0 a 10."))
		}
	}
	for _, path := range bools {
		if v, ok := lookup(root, path); ok {
			if _, isBool := v.(bool); !isBool {
				out = append(out, bad(path, "Valor inválido."))
			}
		}
	}
	return out
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
