package maps

import "sort"

// Object is a prop placed on the map. X,Z is the min corner of its footprint.
type Object struct {
	ID       string         `json:"id"`
	Prefab   string         `json:"prefab"`
	X        int            `json:"x"`
	Z        int            `json:"z"`
	Rotation int            `json:"rotation"` // degrees, multiples of 90
	Props    map[string]any `json:"props,omitempty"`
}

// PrefabRule describes how a prefab affects movement and vision.
type PrefabRule struct {
	BlocksMovement bool
	BlocksVision   bool
	W, D           int // footprint at rotation 0
	Height         int // blocks above the terrain top
}

var prefabs = map[string]PrefabRule{
	"wall":       {true, true, 1, 1, 3},
	"barricade":  {true, false, 1, 1, 1},
	"tree":       {true, true, 1, 1, 3},
	"bush":       {false, false, 1, 1, 1},
	"rock":       {true, true, 1, 1, 1},
	"car":        {true, true, 2, 1, 1},
	"crate":      {true, true, 1, 1, 1},
	"campfire":   {false, false, 1, 1, 1},
	"volcano":    {true, true, 3, 3, 3},
	"watchtower": {true, false, 1, 1, 5},
	"post":       {true, false, 1, 1, 2},
	"fence":      {true, false, 1, 1, 1},
	"door":       {true, true, 1, 1, 3},

	"tent":         {true, false, 2, 2, 2},
	"barrel":       {false, false, 1, 1, 1},
	"bench":        {true, false, 2, 1, 1},
	"market_stall": {false, false, 2, 2, 2},
	"lamp_post":    {false, false, 1, 1, 3},
	"sandbags":     {true, false, 1, 1, 1},
	"ruined_wall":  {true, true, 1, 1, 2},
	"well":         {true, false, 1, 1, 1},
	"sign":         {false, false, 1, 1, 1},
	"bridge":       {false, false, 1, 1, 1},
	"bunting":      {false, false, 2, 1, 2},
}

// RuleFor returns the effective rule; unknown prefabs never block. An open door blocks nothing; a ruined wall is 1 or 2 high (prop height).
func RuleFor(o Object) PrefabRule {
	r, ok := prefabs[o.Prefab]
	if !ok {
		return PrefabRule{W: 1, D: 1}
	}
	if o.Prefab == "door" {
		if open, _ := o.Props["open"].(bool); open {
			r.BlocksMovement, r.BlocksVision = false, false
		}
	}
	if o.Prefab == "ruined_wall" {
		r.Height = 1
		if h, ok := propInt(o.Props, "height"); ok && h >= 2 {
			r.Height = 2
		}
	}
	return r
}

// Footprint returns the cells covered by the object, honoring rotation.
func Footprint(o Object) []Coord {
	r := RuleFor(o)
	w, d := r.W, r.D
	if ((o.Rotation%360)+360)%360%180 == 90 {
		w, d = d, w
	}
	out := make([]Coord, 0, w*d)
	for dz := 0; dz < d; dz++ {
		for dx := 0; dx < w; dx++ {
			out = append(out, Coord{o.X + dx, o.Z + dz})
		}
	}
	return out
}

// KnownPrefab reports whether the prefab id exists in the catalog.
func KnownPrefab(id string) bool {
	_, ok := prefabs[id]
	return ok
}

// PrefabIDs lists every prefab id of the catalog, sorted.
func PrefabIDs() []string {
	ids := make([]string, 0, len(prefabs))
	for id := range prefabs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// PrefabRules returns a copy of the prefab catalog keyed by id.
func PrefabRules() map[string]PrefabRule {
	out := make(map[string]PrefabRule, len(prefabs))
	for id, r := range prefabs {
		out[id] = r
	}
	return out
}
