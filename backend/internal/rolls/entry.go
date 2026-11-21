package rolls

import (
	"encoding/json"
	"strconv"
)

func itoa(n int) string { return strconv.Itoa(n) }

// toEntry decodes a stored record into its client representation.
func toEntry(r Record) (Entry, error) {
	e := Entry{
		RollID: r.ID, UserID: r.UserID, UserName: r.UserName, CharacterID: r.CharacterID, CharacterName: r.CharacterName,
		Kind: r.Kind, Visibility: r.Visibility, Status: r.Status, Seed: r.Seed, CreatedAt: r.CreatedAt,
	}
	if err := json.Unmarshal(r.Dice, &e.Dice); err != nil {
		return Entry{}, err
	}
	if len(r.Result) > 0 {
		e.Result = r.Result
	}
	if r.Kind == KindGeneric {
		return e, decodeGeneric(r, &e)
	}
	var spec assimilacaoSpec
	if err := json.Unmarshal(r.Spec, &spec); err != nil {
		return Entry{}, err
	}
	e.KeptCount, e.Request = spec.Pool.KeptCount, &spec.Request
	return e, nil
}

func decodeGeneric(r Record, e *Entry) error {
	var spec genericSpec
	if err := json.Unmarshal(r.Spec, &spec); err != nil {
		return err
	}
	var res struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(r.Result, &res); err != nil {
		return err
	}
	e.Expression, e.Modifier, e.Total = spec.Expression, spec.Modifier, &res.Total
	e.Result = nil
	return nil
}
