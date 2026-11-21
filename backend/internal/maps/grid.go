// Package maps holds the pure map model: grid, objects, movement and line of sight.
package maps

import (
	"fmt"
	"strconv"
)

// Cell height bounds.
const (
	MinHeight = 0  // lowest cell height
	MaxHeight = 10 // highest cell height
)

// Coord is a cell position on the grid.
type Coord struct {
	X int `json:"x"`
	Z int `json:"z"`
}

// Cell is one grid block.
type Cell struct {
	Terrain string `json:"terrain"`
	Height  int    `json:"height"`
}

// Grid is a width x depth block map; Cells is row-major (z*Width+x).
type Grid struct {
	Width int
	Depth int
	Cells []Cell
}

// Serialized is the compact wire format shared with the frontend.
type Serialized struct {
	Width int    `json:"width"`
	Depth int    `json:"depth"`
	Data  string `json:"data"`
}

// CellPatch changes one cell.
type CellPatch struct {
	X       int    `json:"x"`
	Z       int    `json:"z"`
	Terrain string `json:"terrain"`
	Height  int    `json:"height"`
}

// Patch is a batch of cell changes.
type Patch struct {
	Cells []CellPatch `json:"cells"`
}

// Terrain describes the rules of a terrain type.
type Terrain struct {
	ID             string
	BlocksMovement bool
	BlocksVision   bool
}

// TerrainOrder must match TERRAIN_IDS in the frontend (index is serialized).
var TerrainOrder = []Terrain{
	{"floor", false, false},
	{"grass", false, false},
	{"dirt", false, false},
	{"sand", false, false},
	{"water", false, false},
	{"stone", true, true},
	{"snow", false, false},
	{"lava", false, false},
	{"ruin", true, true},
}

// TerrainByID returns the terrain rules for an id.
func TerrainByID(id string) (Terrain, bool) {
	for _, t := range TerrainOrder {
		if t.ID == id {
			return t, true
		}
	}
	return Terrain{}, false
}

func terrainIndex(id string) int {
	for i, t := range TerrainOrder {
		if t.ID == id {
			return i
		}
	}
	return -1
}

// NewGrid creates a grid filled with the default cell (floor, height 1).
func NewGrid(width, depth int) (*Grid, error) {
	if width < 1 || depth < 1 {
		return nil, fmt.Errorf("o tamanho do mapa deve ser positivo")
	}
	cells := make([]Cell, width*depth)
	for i := range cells {
		cells[i] = Cell{Terrain: "floor", Height: 1}
	}
	return &Grid{Width: width, Depth: depth, Cells: cells}, nil
}

// InBounds reports whether the coordinate is inside the grid.
func (g *Grid) InBounds(x, z int) bool {
	return x >= 0 && z >= 0 && x < g.Width && z < g.Depth
}

// At returns the cell at x,z.
func (g *Grid) At(x, z int) (Cell, bool) {
	if !g.InBounds(x, z) {
		return Cell{}, false
	}
	return g.Cells[z*g.Width+x], true
}

func clampHeight(h int) int {
	return min(MaxHeight, max(MinHeight, h))
}

// Serialize encodes the grid as 2 base36 chars per cell: terrain index + height.
func (g *Grid) Serialize() Serialized {
	buf := make([]byte, 0, len(g.Cells)*2)
	for _, c := range g.Cells {
		buf = strconv.AppendInt(buf, int64(max(terrainIndex(c.Terrain), 0)), 36)
		buf = strconv.AppendInt(buf, int64(clampHeight(c.Height)), 36)
	}
	return Serialized{Width: g.Width, Depth: g.Depth, Data: string(buf)}
}

// Deserialize decodes the compact format.
func Deserialize(s Serialized) (*Grid, error) {
	errInvalid := fmt.Errorf("mapa serializado inválido")
	if s.Width < 1 || s.Depth < 1 || len(s.Data) != s.Width*s.Depth*2 {
		return nil, errInvalid
	}
	cells := make([]Cell, 0, s.Width*s.Depth)
	for i := 0; i < len(s.Data); i += 2 {
		ti, err1 := strconv.ParseInt(s.Data[i:i+1], 36, 32)
		h, err2 := strconv.ParseInt(s.Data[i+1:i+2], 36, 32)
		if err1 != nil || err2 != nil || int(ti) >= len(TerrainOrder) {
			return nil, errInvalid
		}
		cells = append(cells, Cell{Terrain: TerrainOrder[ti].ID, Height: clampHeight(int(h))})
	}
	return &Grid{Width: s.Width, Depth: s.Depth, Cells: cells}, nil
}

// ApplyPatch validates the whole patch first, then applies it atomically.
func (g *Grid) ApplyPatch(p Patch) error {
	for _, c := range p.Cells {
		if !g.InBounds(c.X, c.Z) {
			return fmt.Errorf("célula fora do mapa: (%d, %d)", c.X, c.Z)
		}
		if _, ok := TerrainByID(c.Terrain); !ok {
			return fmt.Errorf("terreno desconhecido: %q", c.Terrain)
		}
		if c.Height < MinHeight || c.Height > MaxHeight {
			return fmt.Errorf("altura inválida (%d): use de %d a %d", c.Height, MinHeight, MaxHeight)
		}
	}
	for _, c := range p.Cells {
		g.Cells[c.Z*g.Width+c.X] = Cell{Terrain: c.Terrain, Height: c.Height}
	}
	return nil
}
