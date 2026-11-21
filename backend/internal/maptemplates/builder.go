package maptemplates

// Builder helpers (deterministic per seed): b := NewBuilder(w, d, seed, "grass", 1) with b.Rng and b.Noise.
// Paint: FillRect/FillCircle/Blob/Line/Road with Brush{Terrain, Height (Keep = unchanged), Reserve};
// River(from, to, width, wobble, bank) -> path; Hill(cx, cz, radius, peak, terrain).

// Structures: Building(rect, BuildingOpts) -> {Rect, Door, Front}; FenceRect/Fence; FurnishCorners(info, prefab, n).
// Objects: Scatter(area, Scatter{...}) by density avoiding reserved/occupied cells; Place; Clear; Reserve.
// ClearEntry() must be the last step (entry area stays walkable and empty); Build() -> grid, objects.

// Check(grid, objects) validates a result; every template test runs it plus Reachable.

import (
	"fmt"
	"math/rand"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// Keep leaves a cell's height unchanged in a Brush or Set.
const Keep = -1

// Pt is a cell position.
type Pt struct{ X, Z int }

// Rect is a cell rectangle (X,Z is the min corner).
type Rect struct{ X, Z, W, D int }

// Contains reports whether the cell is inside the rectangle.
func (r Rect) Contains(x, z int) bool { return x >= r.X && z >= r.Z && x < r.X+r.W && z < r.Z+r.D }

// Grow expands the rectangle by n cells on every side.
func (r Rect) Grow(n int) Rect { return Rect{r.X - n, r.Z - n, r.W + 2*n, r.D + 2*n} }

// Center returns the middle cell.
func (r Rect) Center() Pt { return Pt{r.X + r.W/2, r.Z + r.D/2} }

// Builder accumulates terrain and objects of one map.
type Builder struct {
	W, D     int
	Rng      *rand.Rand
	Noise    Noise
	grid     *maps.Grid
	objs     []maps.Object
	occ      []bool
	reserved []bool
	nextID   int
}

// NewBuilder starts a map filled with the given terrain and height.
func NewBuilder(w, d int, seed int64, terrain string, height int) *Builder {
	g, _ := maps.NewGrid(w, d)
	for i := range g.Cells {
		g.Cells[i] = maps.Cell{Terrain: terrain, Height: height}
	}
	return &Builder{W: w, D: d, Rng: rand.New(rand.NewSource(seed)), Noise: NewNoise(seed), //nolint:gosec // map generation, not security
		grid: g, occ: make([]bool, w*d), reserved: make([]bool, w*d)}
}

// In reports whether the cell is inside the map.
func (b *Builder) In(x, z int) bool { return x >= 0 && z >= 0 && x < b.W && z < b.D }

// Cell returns the cell (zero value outside the map).
func (b *Builder) Cell(x, z int) maps.Cell {
	c, _ := b.grid.At(x, z)
	return c
}

// Set changes a cell; an empty terrain or a Keep height leaves that part unchanged.
func (b *Builder) Set(x, z int, terrain string, height int) {
	if !b.In(x, z) {
		return
	}
	c := &b.grid.Cells[z*b.W+x]
	if terrain != "" {
		c.Terrain = terrain
	}
	if height != Keep {
		c.Height = min(maps.MaxHeight, max(maps.MinHeight, height))
	}
}

// Reserve protects cells from Scatter (roads, water and buildings reserve themselves).
func (b *Builder) Reserve(r Rect) {
	for z := r.Z; z < r.Z+r.D; z++ {
		for x := r.X; x < r.X+r.W; x++ {
			if b.In(x, z) {
				b.reserved[z*b.W+x] = true
			}
		}
	}
}

// Free reports whether the cell is inside the map, unreserved and without objects.
func (b *Builder) Free(x, z int) bool {
	return b.In(x, z) && !b.reserved[z*b.W+x] && !b.occ[z*b.W+x]
}

// Place adds an object when its whole footprint is inside the map and free of other objects.
func (b *Builder) Place(prefab string, x, z, rot int, props map[string]any) bool {
	o := maps.Object{ID: fmt.Sprintf("t%d", b.nextID+1), Prefab: prefab, X: x, Z: z, Rotation: rot, Props: props}
	cells := maps.Footprint(o)
	for _, c := range cells {
		if !b.In(c.X, c.Z) || b.occ[c.Z*b.W+c.X] {
			return false
		}
	}
	for _, c := range cells {
		b.occ[c.Z*b.W+c.X] = true
	}
	b.nextID++
	b.objs = append(b.objs, o)
	return true
}

// Clear removes every object touching the rectangle and turns blocking terrain in it into dirt.
func (b *Builder) Clear(r Rect) {
	kept := b.objs[:0]
	for _, o := range b.objs {
		if touches(o, r) {
			for _, c := range maps.Footprint(o) {
				b.occ[c.Z*b.W+c.X] = false
			}
			continue
		}
		kept = append(kept, o)
	}
	b.objs = kept
	for z := r.Z; z < r.Z+r.D; z++ {
		for x := r.X; x < r.X+r.W; x++ {
			if t, _ := maps.TerrainByID(b.Cell(x, z).Terrain); b.In(x, z) && t.BlocksMovement {
				b.Set(x, z, "dirt", Keep)
			}
		}
	}
}

func touches(o maps.Object, r Rect) bool {
	for _, c := range maps.Footprint(o) {
		if r.Contains(c.X, c.Z) {
			return true
		}
	}
	return false
}

// Build returns the finished grid and objects (copies, safe to keep).
func (b *Builder) Build() (maps.Grid, []maps.Object) {
	g := maps.Grid{Width: b.W, Depth: b.D, Cells: append([]maps.Cell(nil), b.grid.Cells...)}
	return g, append([]maps.Object{}, b.objs...)
}
