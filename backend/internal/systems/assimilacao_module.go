package systems

import (
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

// assimilacaoModule binds the rules package to the sheet interfaces here; it cannot live in
// the assimilacao package because systems already imports it.
type assimilacaoModule struct{ assimilacao.System }

// ValidateSheet reports structural problems as blocking and creation limits as limit issues.
func (assimilacaoModule) ValidateSheet(sheet, _ json.RawMessage) []Issue {
	s, err := assimilacao.ParseSheet(sheet)
	if err != nil {
		return []Issue{{Code: "invalid_sheet", Severity: SeverityBlocking, Message: "Ficha inválida"}}
	}
	out := []Issue{}
	for _, i := range assimilacao.Validate(s) {
		sev := SeverityLimit
		if i.Blocking {
			sev = SeverityBlocking
		}
		out = append(out, Issue{Code: i.Code, Severity: sev, Field: i.Field, Message: i.Message})
	}
	return out
}

// DeriveSheet returns the computed values (health boxes, penalties, XP) as JSON.
func (assimilacaoModule) DeriveSheet(sheet json.RawMessage) (json.RawMessage, error) {
	s, err := assimilacao.ParseSheet(sheet)
	if err != nil {
		return nil, err
	}
	return json.Marshal(assimilacao.Derive(s))
}

// NewSheet returns a blank sheet carrying the character name.
func (assimilacaoModule) NewSheet(name string) (json.RawMessage, error) {
	s := assimilacao.NewSheet()
	s.Identity.Name = name
	return json.Marshal(s)
}
