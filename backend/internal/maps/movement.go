package maps

// DefaultMaxStep is the default max height difference between adjacent cells.
const DefaultMaxStep = 2

// CanEnter checks a cell-to-cell move; maxStep <= 0 uses DefaultMaxStep.
func CanEnter(grid *Grid, objects []Object, from, to Coord, maxStep int) (bool, string) {
	return canEnter(buildIndex(grid, objects), from, to, maxStep)
}

func canEnter(ix *index, from, to Coord, maxStep int) (bool, string) {
	g := ix.grid
	if maxStep <= 0 {
		maxStep = DefaultMaxStep
	}
	dst, ok := g.At(to.X, to.Z)
	if !ok {
		return false, "Destino fora do mapa."
	}
	dstBridge, srcBridge := ix.deck[to.Z*g.Width+to.X], ix.deck[from.Z*g.Width+from.X]
	if t, _ := TerrainByID(dst.Terrain); t.BlocksMovement && dstBridge < 0 {
		return false, "O terreno bloqueia o movimento."
	}
	if ix.moveBlock[to.Z*g.Width+to.X] {
		return false, "Há um objeto bloqueando a célula."
	}
	src, ok := g.At(from.X, from.Z)
	if !ok {
		return false, "Origem fora do mapa."
	}
	if abs(standHeight(dst, dstBridge)-standHeight(src, srcBridge)) > maxStep {
		return false, "A diferença de altura é grande demais."
	}
	return true, ""
}

// standHeight is the cell height, or the deck height when a bridge spans the cell.
func standHeight(c Cell, deck int) int {
	if deck >= 0 {
		return deck
	}
	return c.Height
}

// PathBlocked walks the straight line from->to cell by cell and reports whether any step is illegal.
func PathBlocked(grid *Grid, objects []Object, from, to Coord) bool {
	ix := buildIndex(grid, objects)
	prev := from
	for _, c := range lineCells(from, to) {
		if ok, _ := canEnter(ix, prev, c, 0); !ok {
			return true
		}
		prev = c
	}
	return false
}

// lineCells returns the Bresenham cells after start up to and including end.
func lineCells(a, b Coord) []Coord {
	dx, dz := abs(b.X-a.X), abs(b.Z-a.Z)
	sx, sz := sign(b.X-a.X), sign(b.Z-a.Z)
	var out []Coord
	err := dx - dz
	for a != b {
		e2 := 2 * err
		if e2 > -dz {
			err -= dz
			a.X += sx
		}
		if e2 < dx {
			err += dx
			a.Z += sz
		}
		out = append(out, a)
	}
	return out
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}
