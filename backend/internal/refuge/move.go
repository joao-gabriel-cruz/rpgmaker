package refuge

import (
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Slot names within a sheet inventory.
const (
	SlotBody     = "body"
	SlotBackpack = "backpack"
)

var (
	errNoItem    = realtime.UserError("Item não encontrado")
	errNoChar    = realtime.UserError("Personagem não encontrado")
	errNotYours  = realtime.UserError("Você só pode mover itens dos seus personagens")
	errBadSlot   = realtime.UserError("Local inválido")
	errBodyFull  = realtime.UserError("O corpo está cheio (máximo de 3 itens)")
	errPackFull  = realtime.UserError("A mochila está cheia (máximo de 6 itens)")
	errRefugeMax = realtime.UserError("O Refúgio não comporta mais itens")
)

func slotLimit(slot string) (int, error) {
	switch slot {
	case SlotBody:
		return BodySlots, nil
	case SlotBackpack:
		return BackpackSlots, nil
	}
	return 0, errBadSlot
}

func fullError(slot string) error {
	if slot == SlotBody {
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

func (v *inventory) slot(name string) ([]json.RawMessage, error) {
	items := []json.RawMessage{}
	if raw, ok := v.inv[name]; ok {
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (v *inventory) set(name string, items []json.RawMessage) {
	v.inv[name], _ = json.Marshal(items)
}

func (v *inventory) encode() (json.RawMessage, error) {
	v.sheet["inventory"], _ = json.Marshal(v.inv)
	return json.Marshal(v.sheet)
}

func removeAt(items []json.RawMessage, i int) []json.RawMessage {
	out := make([]json.RawMessage, 0, len(items))
	out = append(out, items[:i]...)
	return append(out, items[i+1:]...)
}

// deposit moves sheet item `index` of slot `from` to the end of the refuge items.
func deposit(c CharState, items []json.RawMessage, from string, index int) (json.RawMessage, []json.RawMessage, error) {
	if _, err := slotLimit(from); err != nil {
		return nil, nil, err
	}
	v, err := openInventory(c.Sheet)
	if err != nil {
		return nil, nil, err
	}
	src, err := v.slot(from)
	if err != nil {
		return nil, nil, err
	}
	if index < 0 || index >= len(src) {
		return nil, nil, errNoItem
	}
	if len(items) >= maxItems {
		return nil, nil, errRefugeMax
	}
	item := src[index]
	v.set(from, removeAt(src, index))
	sheet, err := v.encode()
	return sheet, append(append([]json.RawMessage{}, items...), item), err
}

// withdraw moves refuge item `index` into slot `to` of the sheet.
func withdraw(c CharState, items []json.RawMessage, to string, index int) (json.RawMessage, []json.RawMessage, error) {
	limit, err := slotLimit(to)
	if err != nil {
		return nil, nil, err
	}
	if index < 0 || index >= len(items) {
		return nil, nil, errNoItem
	}
	v, err := openInventory(c.Sheet)
	if err != nil {
		return nil, nil, err
	}
	dst, err := v.slot(to)
	if err != nil {
		return nil, nil, err
	}
	if len(dst) >= limit {
		return nil, nil, fullError(to)
	}
	v.set(to, append(dst, items[index]))
	sheet, err := v.encode()
	return sheet, removeAt(items, index), err
}
