package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

// Scatter describes a vegetation/props spread.
type Scatter struct {
	Prefab    string
	Density   float64  // chance per eligible cell, 0..1
	Terrains  []string // allowed ground (empty = any non-water, non-blocking)
	Variants  []string // tree variants picked at random ("pine", "round"); empty = no variant prop
	Clump     float64  // noise scale in cells for clustering (0 = uniform); denser inside noise peaks
	MaxHeight int      // skip cells taller than this (0 = no limit)
	Rotate    bool     // random rotation
	Props     map[string]any
}

// Scatter places objects over the area by density, skipping reserved cells, occupied cells and water.
// Returns how many were placed. Call it after roads, rivers and buildings so they stay clear.
func (b *Builder) Scatter(area Rect, s Scatter) int {
	placed := 0
	for z := area.Z; z < area.Z+area.D; z++ {
		for x := area.X; x < area.X+area.W; x++ {
			if !b.eligible(x, z, s) || b.Rng.Float64() >= s.Density*b.clumpFactor(x, z, s) {
				continue
			}
			if b.Place(s.Prefab, x, z, b.scatterRotation(s), b.scatterProps(s)) {
				placed++
			}
		}
	}
	return placed
}

func (b *Builder) eligible(x, z int, s Scatter) bool {
	if !b.Free(x, z) {
		return false
	}
	c := b.Cell(x, z)
	t, _ := maps.TerrainByID(c.Terrain)
	if t.BlocksMovement || c.Terrain == "water" || c.Terrain == "lava" || (s.MaxHeight > 0 && c.Height > s.MaxHeight) {
		return false
	}
	return len(s.Terrains) == 0 || containsStr(s.Terrains, c.Terrain)
}

func (b *Builder) clumpFactor(x, z int, s Scatter) float64 {
	if s.Clump <= 0 {
		return 1
	}
	n := b.Noise.FBM(float64(x)/s.Clump+90, float64(z)/s.Clump+40, 2)
	return min(1, max(0, (n-0.4)*4))
}

func (b *Builder) scatterRotation(s Scatter) int {
	if !s.Rotate {
		return 0
	}
	return b.Rng.Intn(4) * 90
}

func (b *Builder) scatterProps(s Scatter) map[string]any {
	if len(s.Variants) == 0 {
		return s.Props
	}
	props := map[string]any{"variant": s.Variants[b.Rng.Intn(len(s.Variants))]}
	for k, v := range s.Props {
		props[k] = v
	}
	return props
}

func containsStr(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// EntryRect is the area where players arrive (bottom center); templates must keep it walkable and empty.
func EntryRect(w, d int) Rect { return Rect{X: w/2 - 3, Z: d - 7, W: 6, D: 5} }

// ClearEntry frees the entry area (objects removed, blocking terrain, water and lava turned to dirt) and reserves it.
func (b *Builder) ClearEntry() {
	r := EntryRect(b.W, b.D)
	b.Clear(r)
	b.Reserve(r)
	for z := r.Z; z < r.Z+r.D; z++ {
		for x := r.X; x < r.X+r.W; x++ {
			if t := b.Cell(x, z).Terrain; t == "water" || t == "lava" {
				b.Set(x, z, "dirt", 1)
			}
		}
	}
}

// ClearPath frees a walkable corridor of the given width between two cells.
func (b *Builder) ClearPath(a, c Pt, width int) {
	lo, hi := -(width / 2), width-1-width/2
	for _, p := range lineCells(a, c) {
		b.Clear(Rect{p.X + lo, p.Z + lo, hi - lo + 1, hi - lo + 1})
	}
}
