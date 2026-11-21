package maptemplates

import (
	"fmt"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// MaxObjects is the most objects a template may place (the scene limit is higher; rendering cost is the reason).
const MaxObjects = 2500

// Check validates a generated map against the rules of internal/maps: size, known terrains and heights,
// known prefabs, unique ids, footprints inside the grid, no overlapping objects, free entry area.
func Check(g maps.Grid, objects []maps.Object) error {
	if g.Width < MinSide || g.Depth < MinSide || g.Width > MaxSide || g.Depth > MaxSide || len(g.Cells) != g.Width*g.Depth {
		return fmt.Errorf("invalid grid %dx%d with %d cells", g.Width, g.Depth, len(g.Cells))
	}
	if err := checkCells(g); err != nil {
		return err
	}
	if len(objects) > MaxObjects {
		return fmt.Errorf("too many objects: %d", len(objects))
	}
	occ, err := checkObjects(g, objects)
	if err != nil {
		return err
	}
	return checkEntry(g, occ)
}

func checkCells(g maps.Grid) error {
	for i, c := range g.Cells {
		if _, ok := maps.TerrainByID(c.Terrain); !ok {
			return fmt.Errorf("cell %d has unknown terrain %q", i, c.Terrain)
		}
		if c.Height < maps.MinHeight || c.Height > maps.MaxHeight {
			return fmt.Errorf("cell %d has invalid height %d", i, c.Height)
		}
	}
	return nil
}

func checkObjects(g maps.Grid, objects []maps.Object) ([]bool, error) {
	occ := make([]bool, len(g.Cells))
	ids := map[string]bool{}
	for _, o := range objects {
		if o.ID == "" || len(o.ID) > 64 || ids[o.ID] {
			return nil, fmt.Errorf("object id %q is empty, too long or duplicated", o.ID)
		}
		ids[o.ID] = true
		if !maps.KnownPrefab(o.Prefab) {
			return nil, fmt.Errorf("object %s has unknown prefab %q", o.ID, o.Prefab)
		}
		if o.Rotation%90 != 0 || o.Rotation < 0 || o.Rotation > 270 {
			return nil, fmt.Errorf("object %s has invalid rotation %d", o.ID, o.Rotation)
		}
		for _, c := range maps.Footprint(o) {
			if !g.InBounds(c.X, c.Z) {
				return nil, fmt.Errorf("object %s (%s) leaves the grid at (%d,%d)", o.ID, o.Prefab, c.X, c.Z)
			}
			if occ[c.Z*g.Width+c.X] {
				return nil, fmt.Errorf("object %s (%s) overlaps another at (%d,%d)", o.ID, o.Prefab, c.X, c.Z)
			}
			occ[c.Z*g.Width+c.X] = true
		}
	}
	return occ, nil
}

func checkEntry(g maps.Grid, occ []bool) error {
	r := EntryRect(g.Width, g.Depth)
	for z := r.Z; z < r.Z+r.D; z++ {
		for x := r.X; x < r.X+r.W; x++ {
			c := g.Cells[z*g.Width+x]
			t, _ := maps.TerrainByID(c.Terrain)
			if occ[z*g.Width+x] || t.BlocksMovement || c.Terrain == "water" || c.Terrain == "lava" {
				return fmt.Errorf("entry area is blocked at (%d,%d)", x, z)
			}
		}
	}
	return nil
}

// Reachable returns the fraction of walkable cells (no blocking terrain or object, not lava) that can be
// walked to from the entry area by 8-way steps (height step <= maps.DefaultMaxStep); templates are tested for >= 0.9.
func Reachable(g maps.Grid, objects []maps.Object) float64 {
	open := make([]bool, len(g.Cells))
	for i, c := range g.Cells {
		t, _ := maps.TerrainByID(c.Terrain)
		open[i] = !t.BlocksMovement && c.Terrain != "lava"
	}
	for _, o := range objects {
		if !maps.RuleFor(o).BlocksMovement {
			continue
		}
		for _, c := range maps.Footprint(o) {
			open[c.Z*g.Width+c.X] = false
		}
	}
	total := 0
	for _, ok := range open {
		if ok {
			total++
		}
	}
	return float64(flood(g, open)) / float64(max(total, 1))
}

func flood(g maps.Grid, open []bool) int {
	r := EntryRect(g.Width, g.Depth)
	start := r.Center()
	seen := map[Pt]bool{start: true}
	queue := []Pt{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for dz := -1; dz <= 1; dz++ {
			for dx := -1; dx <= 1; dx++ {
				n := Pt{cur.X + dx, cur.Z + dz}
				if seen[n] || !g.InBounds(n.X, n.Z) || !open[n.Z*g.Width+n.X] {
					continue
				}
				dh := g.Cells[n.Z*g.Width+n.X].Height - g.Cells[cur.Z*g.Width+cur.X].Height
				if absInt(dh) <= maps.DefaultMaxStep {
					seen[n] = true
					queue = append(queue, n)
				}
			}
		}
	}
	return len(seen)
}
