package rolls

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	maxDice  = 50
	maxSides = 1000
	maxMod   = 10000
)

var exprRe = regexp.MustCompile(`^(\d*)d(\d+)(?:([+-])(\d+))?$`)

// parseExpression parses XdY, XdY+Z or XdY-Z (X defaults to 1); errors are PT-BR user messages.
func parseExpression(raw string) (genericSpec, error) {
	expr := strings.ToLower(strings.Join(strings.Fields(raw), ""))
	m := exprRe.FindStringSubmatch(expr)
	if m == nil {
		return genericSpec{}, ruleErr("Expressão inválida. Use o formato 2d6+3.")
	}
	count := 1
	if m[1] != "" {
		count, _ = strconv.Atoi(m[1])
	}
	sides, _ := strconv.Atoi(m[2])
	mod := 0
	if m[4] != "" {
		mod, _ = strconv.Atoi(m[4])
		if m[3] == "-" {
			mod = -mod
		}
	}
	switch {
	case count < 1 || count > maxDice:
		return genericSpec{}, ruleErr("Use de 1 a 50 dados.")
	case sides < 2 || sides > maxSides:
		return genericSpec{}, ruleErr("Os dados devem ter de 2 a 1000 faces.")
	case mod > maxMod || mod < -maxMod:
		return genericSpec{}, ruleErr("O modificador é grande demais.")
	}
	return genericSpec{Expression: expr, Count: count, Sides: sides, Modifier: mod}, nil
}

// RuleError is a PT-BR, user-facing validation failure.
type RuleError string

func (e RuleError) Error() string { return string(e) }

func ruleErr(msg string) error { return RuleError(msg) }
