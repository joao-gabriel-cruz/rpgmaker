// Package items owns the campaign item compendium, giving and using items, and loot hidden on maps.
package items

import (
	"encoding/json"
	"errors"
)

// Item kinds.
const (
	KindMelee      = "melee"
	KindRanged     = "ranged"
	KindTool       = "tool"
	KindConsumable = "consumable"
	KindClothing   = "clothing"
	KindResource   = "resource"
	KindKey        = "key"
	KindOther      = "other"
)

// Inventory places of a sheet.
const (
	PlaceBody     = "body"
	PlaceBackpack = "backpack"
	TargetRefuge  = "refuge"
)

// Sheet inventory limits.
const (
	BodySlots     = 3
	BackpackSlots = 6
)

// Events emitted to clients.
const (
	EventItemsChanged = "items.changed"
	EventLootPlaced   = "loot.placed"
	EventLootRemoved  = "loot.removed"
	EventLootFound    = "loot.found"
	EventItemUsed     = "item.used"
)

// ErrNotFound means the item or loot does not exist in the campaign.
var ErrNotFound = errors.New("items: not found")

// KeptBonus adds kept dice to actions of one category.
type KeptBonus struct {
	ActionCategory string `json:"actionCategory"`
	Amount         int    `json:"amount"`
}

// RefugeResource is added to the Refúgio when the item is consumed.
type RefugeResource struct {
	Resource string `json:"resource"`
	Amount   int    `json:"amount"`
}

// ItemEffects are the optional mechanical effects of an item.
type ItemEffects struct {
	KeptBonus      *KeptBonus      `json:"keptBonus,omitempty"`
	Heal           int             `json:"heal,omitempty"`
	RefugeResource *RefugeResource `json:"refugeResource,omitempty"`
}

// Item is a compendium entry.
type Item struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Icon            string      `json:"icon"`
	Color           string      `json:"color"`
	Kind            string      `json:"kind"`
	Quality         int         `json:"quality"`
	Scarcity        int         `json:"scarcity"`
	Pressure        int         `json:"pressure"`
	Traits          []string    `json:"traits"`
	Effects         ItemEffects `json:"effects"`
	AttackAnimation string      `json:"attackAnimation,omitempty"`
}

// Loot is an item hidden (or lying) on a scene cell or inside an object.
type Loot struct {
	ID         string          `json:"id"`
	SceneID    string          `json:"sceneId"`
	X          int             `json:"x"`
	Z          int             `json:"z"`
	ObjectID   *string         `json:"objectId"`
	ItemID     *string         `json:"itemId"`
	Item       json.RawMessage `json:"item"`
	Qty        int             `json:"qty"`
	Difficulty int             `json:"difficulty"`
	Hidden     bool            `json:"hidden"`
	FoundBy    *string         `json:"foundBy"`
}

// Found reports one loot revealed by TryFind and how many copies reached the finder's backpack.
type Found struct {
	Loot      Loot `json:"loot"`
	Delivered int  `json:"delivered"`
	Left      int  `json:"left"`
}
