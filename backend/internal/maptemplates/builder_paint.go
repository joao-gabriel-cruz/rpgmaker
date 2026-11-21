package maptemplates

import "math"

// Brush says what a paint helper writes; Height Keep leaves heights alone, Reserve protects the cells from Scatter.
type Brush struct {
	Terrain string
	Height  int
	Reserve bool
}

func (b *Builder) paint(x, z int, br Brush) {
	b.Set(x, z, br.Terrain, br.Height)
	if br.Reserve && b.In(x, z) {
		b.reserved[z*b.W+x] = true
	}
}

// FillRect paints a rectangle.
func (b *Builder) FillRect(r Rect, br Brush) {
	for z := r.Z; z < r.Z+r.D; z++ {
		for x := r.X; x < r.X+r.W; x++ {
			b.paint(x, z, br)
		}
	}
}

// FillCircle paints a disc of the given radius (in cells) centered on a cell.
func (b *Builder) FillCircle(cx, cz int, radius float64, br Brush) {
	r := int(math.Ceil(radius))
	for z := cz - r; z <= cz+r; z++ {
		for x := cx - r; x <= cx+r; x++ {
			if dx, dz := float64(x-cx), float64(z-cz); dx*dx+dz*dz <= radius*radius {
				b.paint(x, z, br)
			}
		}
	}
}

// Blob paints an organic patch: a disc whose edge wobbles with noise (roughness 0..1).
func (b *Builder) Blob(cx, cz int, radius, roughness float64, br Brush) {
	r := int(math.Ceil(radius * 1.5))
	for z := cz - r; z <= cz+r; z++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dz := float64(x-cx), float64(z-cz)
			wob := 1 + (b.Noise.FBM(float64(x)/4+31, float64(z)/4+17, 2)-0.5)*2*roughness
			if math.Hypot(dx, dz) <= radius*wob {
				b.paint(x, z, br)
			}
		}
	}
}

// Line paints a straight line of the given width (cells) between two cells.
func (b *Builder) Line(a, c Pt, width int, br Brush) {
	lo, hi := -(width / 2), width-1-width/2
	for _, p := range lineCells(a, c) {
		for dz := lo; dz <= hi; dz++ {
			for dx := lo; dx <= hi; dx++ {
				b.paint(p.X+dx, p.Z+dz, br)
			}
		}
	}
}

// lineCells is Bresenham including both ends.
func lineCells(a, c Pt) []Pt {
	dx, dz := absInt(c.X-a.X), absInt(c.Z-a.Z)
	sx, sz := sgn(c.X-a.X), sgn(c.Z-a.Z)
	out := []Pt{a}
	for err := dx - dz; a != c; {
		prev := a
		if e2 := 2 * err; e2 > -dz {
			err -= dz
			a.X += sx
		} else if e2 < dx {
			err += dx
			a.Z += sz
		}
		if a != prev {
			out = append(out, a)
		}
	}
	return out
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sgn(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}
