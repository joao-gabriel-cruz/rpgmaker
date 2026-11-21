package maptemplates

import (
	"math"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// bridgeSpans opens bridge cells (a bridge spans blocking terrain) and returns the stand height per cell.
func bridgeSpans(g maps.Grid, objects []maps.Object, open []bool) []int {
	heights := make([]int, len(g.Cells))
	for i, c := range g.Cells {
		heights[i] = c.Height
	}
	for _, o := range objects {
		if o.Prefab != "bridge" {
			continue
		}
		deck, has := propDeck(o)
		for _, c := range maps.Footprint(o) {
			if g.InBounds(c.X, c.Z) {
				open[c.Z*g.Width+c.X] = true
				if has {
					heights[c.Z*g.Width+c.X] = deck
				}
			}
		}
	}
	return heights
}

// Ramp paints a walkable slope between two cells: heights are interpolated from the first cell's height
// to the last one's (keep the gradient at 2 per cell or less); the strip is reserved from Scatter.
func (b *Builder) Ramp(a, c Pt, width int, terrain string) {
	cells := lineCells(a, c)
	h0, h1 := b.Cell(a.X, a.Z).Height, b.Cell(c.X, c.Z).Height
	lo := -(width / 2)
	for i, p := range cells {
		h := h0
		if len(cells) > 1 {
			h = h0 + int(math.Round(float64(h1-h0)*float64(i)/float64(len(cells)-1)))
		}
		b.FillRect(Rect{p.X + lo, p.Z + lo, width, width}, Brush{Terrain: terrain, Height: h, Reserve: true})
	}
}

// Ridge raises a long smooth mountain chain along a segment: peak extra blocks on the spine fading out at
// halfWidth cells; terrain (optional) is applied where the rise is meaningful.
func (b *Builder) Ridge(a, c Pt, halfWidth float64, peak int, terrain string) {
	r := int(math.Ceil(halfWidth))
	x0, x1, z0, z1 := min(a.X, c.X)-r, max(a.X, c.X)+r, min(a.Z, c.Z)-r, max(a.Z, c.Z)+r
	for z := z0; z <= z1; z++ {
		for x := x0; x <= x1; x++ {
			dist := segDist(float64(x), float64(z), a, c) / halfWidth
			if !b.In(x, z) || dist >= 1 {
				continue
			}
			f := smooth(1-dist) * (0.8 + 0.4*b.Noise.FBM(float64(x)/5, float64(z)/5, 2))
			t := ""
			if f > 0.3 {
				t = terrain
			}
			b.Set(x, z, t, b.Cell(x, z).Height+int(math.Round(float64(peak)*f)))
		}
	}
}

func segDist(px, pz float64, a, c Pt) float64 {
	ax, az, dx, dz := float64(a.X), float64(a.Z), float64(c.X-a.X), float64(c.Z-a.Z)
	t := 0.0
	if l := dx*dx + dz*dz; l > 0 {
		t = math.Min(1, math.Max(0, ((px-ax)*dx+(pz-az)*dz)/l))
	}
	return math.Hypot(px-ax-dx*t, pz-az-dz*t)
}

// LavaFlow carves a winding lava channel, one block below the ground it crosses; the channel is reserved.
func (b *Builder) LavaFlow(from, to Pt, width int, wobble float64) []Pt {
	path := b.windingPath(from, to, wobble)
	lo := -(width / 2)
	for _, p := range path {
		h := max(0, b.Cell(p.X, p.Z).Height-1)
		b.FillRect(Rect{p.X + lo, p.Z + lo, width, width}, Brush{Terrain: "lava", Height: h, Reserve: true})
	}
	return path
}

// strip calls fn for every cell of a line widened to width cells (perpendicular to its main direction).
func strip(a, c Pt, width int, fn func(x, z int)) {
	horizontal := absInt(c.X-a.X) > absInt(c.Z-a.Z)
	lo := -(width / 2)
	for _, p := range lineCells(a, c) {
		for k := 0; k < width; k++ {
			if horizontal {
				fn(p.X, p.Z+lo+k)
			} else {
				fn(p.X+lo+k, p.Z)
			}
		}
	}
}

// plank places one bridge plank at the deck height and protects it from Scatter.
func (b *Builder) plank(x, z, deck int) bool {
	if !b.Place("bridge", x, z, 0, map[string]any{"deck": deck}) {
		return false
	}
	b.Reserve(Rect{x, z, 1, 1})
	return true
}

// Bridge lays bridge planks along a line (width cells wide) at the given deck height (deck < 0 = the higher
// end of the span); returns how many planks were placed. Call it before Scatter so the planks keep their cells.
func (b *Builder) Bridge(a, c Pt, width, deck int) int {
	if deck < 0 {
		deck = max(b.Cell(a.X, a.Z).Height, b.Cell(c.X, c.Z).Height)
	}
	placed := 0
	strip(a, c, width, func(x, z int) {
		if b.plank(x, z, deck) {
			placed++
		}
	})
	return placed
}

// Walkway lays a trail through waypoints: dirt over land and planks (at the deck height) over water.
// Returns how many planks were placed.
func (b *Builder) Walkway(points []Pt, width, deck int) int {
	placed := 0
	for i := 1; i < len(points); i++ {
		strip(points[i-1], points[i], width, func(x, z int) {
			switch {
			case !b.In(x, z):
			case b.Cell(x, z).Terrain == "water":
				if b.plank(x, z, deck) {
					placed++
				}
			default:
				b.Set(x, z, "dirt", Keep)
				b.Reserve(Rect{x, z, 1, 1})
			}
		})
	}
	return placed
}

func propDeck(o maps.Object) (int, bool) {
	switch v := o.Props["deck"].(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	}
	return 0, false
}

// ReachSet returns every cell that can be walked to from the entry area (8-way steps, bridges honored).
func ReachSet(g maps.Grid, objects []maps.Object) map[Pt]bool {
	open, heights := walkGrid(g, objects)
	return flood(g, open, heights)
}

// SealOff turns every walkable cell the entry cannot reach (cut off by cliffs, lava or lakes) into a tall
// mass of the given blocking terrain and removes the objects on it. Call it after ClearEntry.
func (b *Builder) SealOff(terrain string, minHeight int) {
	open, heights := walkGrid(*b.grid, b.objs)
	reached := flood(*b.grid, open, heights)
	for z := 0; z < b.D; z++ {
		for x := 0; x < b.W; x++ {
			if b.sealable(x, z, open[z*b.W+x], reached) {
				b.Set(x, z, terrain, max(b.Cell(x, z).Height, minHeight+int(b.Noise.FBM(float64(x)/4, float64(z)/4, 2)*4)))
			}
		}
	}
	kept := b.objs[:0]
	for _, o := range b.objs {
		if t, _ := maps.TerrainByID(b.Cell(o.X, o.Z).Terrain); t.BlocksMovement && o.Prefab != "bridge" {
			for _, c := range maps.Footprint(o) {
				b.occ[c.Z*b.W+c.X] = false
			}
			continue
		}
		kept = append(kept, o)
	}
	b.objs = kept
}

// sealable reports whether a terrain-walkable cell is cut off; an object cell touching the reached area is not.
func (b *Builder) sealable(x, z int, open bool, reached map[Pt]bool) bool {
	c := b.Cell(x, z)
	if t, _ := maps.TerrainByID(c.Terrain); t.BlocksMovement || c.Terrain == "lava" || reached[Pt{x, z}] {
		return false
	}
	if open {
		return true
	}
	for dz := -1; dz <= 1; dz++ {
		for dx := -1; dx <= 1; dx++ {
			if reached[Pt{x + dx, z + dz}] {
				return false
			}
		}
	}
	return true
}
