package effects

import (
	"strings"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func knownKey(kind, key string) bool {
	switch kind {
	case KindTerrain:
		_, ok := maps.TerrainByID(key)
		return ok
	case KindPrefab:
		return maps.KnownPrefab(key)
	}
	return false
}

// normalize validates and trims the rules the Mestre submits.
func normalize(in []Rule) ([]Rule, error) {
	if len(in) > MaxRulesPerKey {
		return nil, badInput("Há regras demais para este item")
	}
	out := make([]Rule, 0, len(in))
	for _, r := range in {
		r.Message = strings.TrimSpace(r.Message)
		if err := checkRule(r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func checkRule(r Rule) error {
	if r.Trigger != TriggerEnter && r.Trigger != TriggerTurnStart {
		return badInput("Gatilho inválido")
	}
	if len([]rune(r.Message)) > MaxMessageLen {
		return badInput("Mensagem longa demais")
	}
	switch r.Effect {
	case EffectDamage:
		return amountBetween(r.Amount, 1, MaxDamage)
	case EffectDifficulty:
		return amountBetween(r.Amount, 1, MaxDifficulty)
	case EffectNotice:
		return amountBetween(r.Amount, 0, 0)
	}
	return badInput("Efeito inválido")
}

func amountBetween(n, lo, hi int) error {
	if n < lo || n > hi {
		return badInput("Quantidade fora do limite permitido")
	}
	return nil
}

// inputError is a validation failure whose PT-BR message may be shown to the Mestre.
type inputError string

func (e inputError) Error() string { return string(e) }

func badInput(msg string) error { return inputError(msg) }
