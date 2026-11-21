package items

import (
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

var (
	errNoItem   = realtime.UserError("Item não encontrado")
	errNoChar   = realtime.UserError("Personagem não encontrado")
	errBadPlace = realtime.UserError("Local inválido")
	errBodyFull = realtime.UserError("O corpo não comporta mais itens (máximo de 3)")
	errPackFull = realtime.UserError("A mochila não comporta mais itens (máximo de 6)")
)

// SheetItem converts a compendium item into the sheet item format (plus the optional compendium fields).
func SheetItem(it Item) json.RawMessage {
	m := map[string]any{
		"name": it.Name, "quality": it.Quality, "scarcity": it.Scarcity, "pressure": it.Pressure,
		"traits": it.Traits, "itemId": it.ID, "icon": it.Icon, "color": it.Color, "kind": it.Kind,
		"description": it.Description, "effects": it.Effects,
	}
	if it.AttackAnimation != "" {
		m["attackAnimation"] = it.AttackAnimation
	}
	raw, _ := json.Marshal(m)
	return raw
}

func placeLimit(place string) (int, error) {
	switch place {
	case PlaceBody:
		return BodySlots, nil
	case PlaceBackpack:
		return BackpackSlots, nil
	}
	return 0, errBadPlace
}

func fullError(place string) error {
	if place == PlaceBody {
		return errBodyFull
	}
	return errPackFull
}

// inventory edits the sheet's inventory slots while keeping every other field verbatim.
type inventory struct {
	sheet map[string]json.RawMessage
	inv   map[string]json.RawMessage
}

func openInventory(sheet json.RawMessage) (*inventory, error) {
	v := &inventory{sheet: map[string]json.RawMessage{}, inv: map[string]json.RawMessage{}}
	if err := json.Unmarshal(sheet, &v.sheet); err != nil {
		return nil, err
	}
	if raw, ok := v.sheet["inventory"]; ok {
		if err := json.Unmarshal(raw, &v.inv); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (v *inventory) slot(place string) ([]json.RawMessage, error) {
	items := []json.RawMessage{}
	if raw, ok := v.inv[place]; ok {
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (v *inventory) set(place string, items []json.RawMessage) {
	v.inv[place], _ = json.Marshal(items)
}

func (v *inventory) encode() (json.RawMessage, error) {
	v.sheet["inventory"], _ = json.Marshal(v.inv)
	return json.Marshal(v.sheet)
}

// addCopies appends up to qty copies of item to the places in order and reports how many fit.
func addCopies(sheet json.RawMessage, item json.RawMessage, qty int, places []string) (json.RawMessage, int, error) {
	v, err := openInventory(sheet)
	if err != nil {
		return nil, 0, err
	}
	added := 0
	for _, place := range places {
		limit, err := placeLimit(place)
		if err != nil {
			return nil, 0, err
		}
		list, err := v.slot(place)
		if err != nil {
			return nil, 0, err
		}
		for len(list) < limit && added < qty {
			list, added = append(list, item), added+1
		}
		v.set(place, list)
	}
	out, err := v.encode()
	return out, added, err
}

// takeAt removes the item at index of place and returns the new sheet and the item.
func takeAt(sheet json.RawMessage, place string, index int) (json.RawMessage, json.RawMessage, error) {
	if _, err := placeLimit(place); err != nil {
		return nil, nil, err
	}
	v, err := openInventory(sheet)
	if err != nil {
		return nil, nil, err
	}
	list, err := v.slot(place)
	if err != nil {
		return nil, nil, err
	}
	if index < 0 || index >= len(list) {
		return nil, nil, errNoItem
	}
	taken := list[index]
	v.set(place, append(append([]json.RawMessage{}, list[:index]...), list[index+1:]...))
	out, err := v.encode()
	return out, taken, err
}
