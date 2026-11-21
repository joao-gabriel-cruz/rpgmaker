package maptemplates

// Extra helpers used by the city, forest, farm, refuge and festival templates.

// PlaceSpaced places n objects starting at from, stepping (dx, dz) cells between them; returns how many fit.
func (b *Builder) PlaceSpaced(prefab string, from Pt, dx, dz, n, rot int, props map[string]any) int {
	placed := 0
	for i := 0; i < n; i++ {
		if b.Place(prefab, from.X+dx*i, from.Z+dz*i, rot, props) {
			placed++
		}
	}
	return placed
}

// PaintTrail paints a walkable strip along path (reserved from Scatter) without touching water cells.
func (b *Builder) PaintTrail(path []Pt, width int, terrain string) {
	lo, hi := -(width / 2), width-1-width/2
	for _, p := range path {
		for dz := lo; dz <= hi; dz++ {
			for dx := lo; dx <= hi; dx++ {
				x, z := p.X+dx, p.Z+dz
				if b.In(x, z) && b.Cell(x, z).Terrain != "water" {
					b.Set(x, z, terrain, Keep)
					b.Reserve(Rect{x, z, 1, 1})
				}
			}
		}
	}
}

// WindingPath exposes the noise-driven 4-connected path used by River.
func (b *Builder) WindingPath(a, c Pt, wobble float64) []Pt { return b.windingPath(a, c, wobble) }

// WallRing places the prefab on every cell of the rectangle perimeter except the skipped cells.
func (b *Builder) WallRing(r Rect, prefab string, skip ...Pt) {
	for z := r.Z; z < r.Z+r.D; z++ {
		for x := r.X; x < r.X+r.W; x++ {
			if x != r.X && z != r.Z && x != r.X+r.W-1 && z != r.Z+r.D-1 {
				continue
			}
			if !containsPt(skip, Pt{x, z}) {
				b.Place(prefab, x, z, 0, nil)
			}
		}
	}
}

func containsPt(list []Pt, p Pt) bool {
	for _, q := range list {
		if q == p {
			return true
		}
	}
	return false
}

// sideFrom picks a random side among n consecutive ones starting at first (North, South, East, West order).
func (b *Builder) sideFrom(first Side, n int) Side { return first + Side(b.Rng.Intn(n)) }
