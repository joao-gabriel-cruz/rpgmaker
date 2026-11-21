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

// NewNPCSheet returns a blank simplified NPC sheet carrying the NPC name.
func (assimilacaoModule) NewNPCSheet(name string) (json.RawMessage, error) {
	s := assimilacao.NewNPCSheet()
	s.Identity.Name = name
	return json.Marshal(s)
}

// ValidateNPCSheet reports structural problems only; NPCs have no creation limits.
func (assimilacaoModule) ValidateNPCSheet(sheet json.RawMessage) []Issue {
	s, err := assimilacao.ParseNPCSheet(sheet)
	if err != nil {
		return []Issue{{Code: "invalid_sheet", Severity: SeverityBlocking, Message: "Ficha inválida"}}
	}
	out := []Issue{}
	for _, i := range assimilacao.ValidateNPC(s) {
		out = append(out, Issue{Code: i.Code, Severity: SeverityBlocking, Field: i.Field, Message: i.Message})
	}
	return out
}

// DeriveNPCSheet returns the health level computed from the manual boxes per level.
func (assimilacaoModule) DeriveNPCSheet(sheet json.RawMessage) (json.RawMessage, error) {
	s, err := assimilacao.ParseNPCSheet(sheet)
	if err != nil {
		return nil, err
	}
	return json.Marshal(assimilacao.DeriveNPC(s))
}
