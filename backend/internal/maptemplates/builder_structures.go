package maptemplates

// Side is a side of a rectangle; NoSide means none.
type Side int

// Rectangle sides (North is the low-Z edge).
const (
	NoSide Side = iota
	North
	South
	East
	West
)

// BuildingOpts configures Building.
type BuildingOpts struct {
	Door     Side    // side with the door (default South)
	DoorAt   int     // offset of the door from the corner along that side (0 = centered)
	Floor    string  // interior terrain (default "floor")
	Open     bool    // door starts open
	Ruin     float64 // fraction (0..1) of wall segments missing, never next to the door
	NoDoor   bool    // leave a plain wall instead of a door
	Material string  // wall prefab (default "wall")
}

// BuildingInfo describes a placed building.
type BuildingInfo struct {
	Rect  Rect
	Door  Pt // door cell
	Front Pt // cell just outside the door (kept free)
}

// Building raises a rectangular building: flattened floor, perimeter walls with a door, interior reserved.
// Objects already in Rect are removed; Rect must be at least 3x3. The area plus a 1 cell margin is reserved from Scatter.
func (b *Builder) Building(r Rect, o BuildingOpts) BuildingInfo {
	if o.Door == NoSide {
		o.Door = South
	}
	if o.Floor == "" {
		o.Floor = "floor"
	}
	if o.Material == "" {
		o.Material = "wall"
	}
	b.Clear(r)
	h := b.Cell(r.Center().X, r.Center().Z).Height
	b.FillRect(r, Brush{Terrain: o.Floor, Height: h})
	b.Reserve(r.Grow(1))
	info := BuildingInfo{Rect: r}
	info.Door, info.Front = doorCells(r, o)
	b.placeWalls(r, o, info.Door)
	if !o.NoDoor {
		b.Clear(Rect{info.Door.X, info.Door.Z, 1, 1})
		b.Place("door", info.Door.X, info.Door.Z, 0, map[string]any{"open": o.Open})
		b.Clear(Rect{info.Front.X, info.Front.Z, 1, 1})
	}
	return info
}

func doorCells(r Rect, o BuildingOpts) (door, front Pt) {
	at := func(span int) int {
		if o.DoorAt > 0 {
			return min(max(o.DoorAt, 1), span-2)
		}
		return span / 2
	}
	switch o.Door {
	case North:
		door, front = Pt{r.X + at(r.W), r.Z}, Pt{r.X + at(r.W), r.Z - 1}
	case East:
		door, front = Pt{r.X + r.W - 1, r.Z + at(r.D)}, Pt{r.X + r.W, r.Z + at(r.D)}
	case West:
		door, front = Pt{r.X, r.Z + at(r.D)}, Pt{r.X - 1, r.Z + at(r.D)}
	default:
		door, front = Pt{r.X + at(r.W), r.Z + r.D - 1}, Pt{r.X + at(r.W), r.Z + r.D}
	}
	return door, front
}

func (b *Builder) placeWalls(r Rect, o BuildingOpts, door Pt) {
	for z := r.Z; z < r.Z+r.D; z++ {
		for x := r.X; x < r.X+r.W; x++ {
			edge := x == r.X || z == r.Z || x == r.X+r.W-1 || z == r.Z+r.D-1
			dist := absInt(x-door.X) + absInt(z-door.Z)
			if !edge || (dist == 0 && !o.NoDoor) || (dist > 1 && o.Ruin > 0 && b.Rng.Float64() < o.Ruin) {
				continue
			}
			b.Place(o.Material, x, z, 0, nil)
		}
	}
}

// Fence lays fence segments along a line, leaving the listed indexes (0 = first cell) open as gaps.
func (b *Builder) Fence(from, to Pt, gaps ...int) {
	for i, p := range lineCells(from, to) {
		if !contains(gaps, i) && b.Free(p.X, p.Z) {
			b.Place("fence", p.X, p.Z, 0, nil)
		}
	}
}

// FenceRect fences the rectangle perimeter with a 2 cell gap in the middle of the gap side.
func (b *Builder) FenceRect(r Rect, gap Side) {
	x1, z1 := r.X+r.W-1, r.Z+r.D-1
	mid := func(span int) []int { return []int{span/2 - 1, span / 2} }
	pick := func(s Side, g []int) []int {
		if gap == s {
			return g
		}
		return nil
	}
	b.Fence(Pt{r.X, r.Z}, Pt{x1, r.Z}, pick(North, mid(r.W))...)
	b.Fence(Pt{r.X, z1}, Pt{x1, z1}, pick(South, mid(r.W))...)
	b.Fence(Pt{r.X, r.Z + 1}, Pt{r.X, z1 - 1}, pick(West, mid(r.D-2))...)
	b.Fence(Pt{x1, r.Z + 1}, Pt{x1, z1 - 1}, pick(East, mid(r.D-2))...)
}

func contains(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// FurnishCorners places up to n objects (crates, barrels...) in random interior corners of a building,
// which never blocks the door. Returns how many were placed.
func (b *Builder) FurnishCorners(info BuildingInfo, prefab string, n int) int {
	r := info.Rect
	corners := []Pt{{r.X + 1, r.Z + 1}, {r.X + r.W - 2, r.Z + 1}, {r.X + 1, r.Z + r.D - 2}, {r.X + r.W - 2, r.Z + r.D - 2}}
	b.Rng.Shuffle(len(corners), func(i, j int) { corners[i], corners[j] = corners[j], corners[i] })
	placed := 0
	for _, c := range corners {
		if placed < n && absInt(c.X-info.Door.X)+absInt(c.Z-info.Door.Z) > 2 && b.Place(prefab, c.X, c.Z, 0, nil) {
			placed++
		}
	}
	return placed
}
