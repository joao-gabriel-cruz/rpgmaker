package maptemplates

import "math"

// Road paints a polyline of the given width, flattened to height and reserved from Scatter.
func (b *Builder) Road(points []Pt, width int, terrain string, height int) {
	br := Brush{Terrain: terrain, Height: height, Reserve: true}
	for i := 1; i < len(points); i++ {
		b.Line(points[i-1], points[i], width, br)
	}
}

// River carves a winding channel from one cell to another: water at height 0 flanked by banks (sand/dirt) at
// height 1. wobble is the sideways swing in cells. Returns the centerline; water and banks are reserved.
func (b *Builder) River(from, to Pt, width int, wobble float64, bank string) []Pt {
	path := b.windingPath(from, to, wobble)
	bankBrush := Brush{Terrain: bank, Height: 1, Reserve: true}
	water := Brush{Terrain: "water", Height: 0, Reserve: true}
	for _, p := range path {
		b.Line(p, p, width+4, bankBrush)
	}
	for _, p := range path {
		b.Line(p, p, width, water)
	}
	return path
}

// windingPath is a 4-connected path from a to c that swings sideways following noise.
func (b *Builder) windingPath(a, c Pt, wobble float64) []Pt {
	dx, dz := float64(c.X-a.X), float64(c.Z-a.Z)
	length := math.Hypot(dx, dz)
	nx, nz := -dz/length, dx/length
	steps := int(length * 2)
	out := []Pt{a}
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		fade := math.Min(1, t*6) * math.Min(1, (1-t)*6)
		swing := (b.Noise.FBM(t*length/10+5.5, 3.3, 2) - 0.5) * 9 * wobble * fade
		p := Pt{a.X + int(math.Round(dx*t+nx*swing)), a.Z + int(math.Round(dz*t+nz*swing))}
		if last := out[len(out)-1]; p != last {
			out = append(out, lineCells(last, p)[1:]...)
		}
	}
	return out
}

// Hill raises a smooth mound: peak extra blocks at the center fading to 0 at radius; terrain (optional)
// is applied where the rise is meaningful.
func (b *Builder) Hill(cx, cz int, radius float64, peak int, terrain string) {
	r := int(math.Ceil(radius))
	for z := cz - r; z <= cz+r; z++ {
		for x := cx - r; x <= cx+r; x++ {
			dist := math.Hypot(float64(x-cx), float64(z-cz)) / radius
			if !b.In(x, z) || dist >= 1 {
				continue
			}
			f := smooth(1 - dist)
			f *= 0.9 + 0.2*b.Noise.At(float64(x)/3, float64(z)/3)
			rise := int(math.Round(float64(peak) * f))
			t := ""
			if f > 0.35 {
				t = terrain
			}
			b.Set(x, z, t, b.Cell(x, z).Height+rise)
		}
	}
}
