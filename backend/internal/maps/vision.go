package maps

import "math"

// Set is a set of visible cells.
type Set map[Coord]struct{}

const (
	// sampleStep is the ray sampling interval in cells (supercover approximation).
	sampleStep = 0.25
	// targetAim lifts the aim point above the target's terrain top.
	targetAim = 0.5
)

// Visible computes line of sight: eye = terrain height + eyeHeight; rays aim at target tops within radius.
// Wall top = terrain height (+3 if terrain blocks vision) or + object height.
// The viewer cell and neighbors are always visible.
func Visible(grid *Grid, objects []Object, viewer Coord, eyeHeight float64, radius int) Set {
	return visibleFrom(buildIndex(grid, objects), viewer, eyeHeight, radius)
}

// VisibleForMany returns the union of visibility for several viewers.
func VisibleForMany(grid *Grid, objects []Object, viewers []Coord, eyeHeight float64, radius int) Set {
	ix := buildIndex(grid, objects)
	out := Set{}
	for _, v := range viewers {
		for c := range visibleFrom(ix, v, eyeHeight, radius) {
			out[c] = struct{}{}
		}
	}
	return out
}

func visibleFrom(ix *index, v Coord, eyeHeight float64, radius int) Set {
	g := ix.grid
	out := Set{}
	vc, ok := g.At(v.X, v.Z)
	if !ok {
		return out
	}
	eye := float64(vc.Height) + eyeHeight
	for z := v.Z - radius; z <= v.Z+radius; z++ {
		for x := v.X - radius; x <= v.X+radius; x++ {
			dx, dz := x-v.X, z-v.Z
			if !g.InBounds(x, z) || dx*dx+dz*dz > radius*radius {
				continue
			}
			if abs(dx) <= 1 && abs(dz) <= 1 || rayClear(ix, v, Coord{x, z}, eye) {
				out[Coord{x, z}] = struct{}{}
			}
		}
	}
	return out
}

// rayClear walks the ray from the eye to the target aim point, testing intermediate cells.
func rayClear(ix *index, from, to Coord, eye float64) bool {
	g := ix.grid
	aim := float64(g.Cells[to.Z*g.Width+to.X].Height) + targetAim
	dx, dz := float64(to.X-from.X), float64(to.Z-from.Z)
	dist := math.Hypot(dx, dz)
	steps := int(dist / sampleStep)
	last := from
	for i := 1; i < steps; i++ {
		t := float64(i) / float64(steps)
		c := Coord{from.X + int(math.Floor(dx*t+0.5)), from.Z + int(math.Floor(dz*t+0.5))}
		if c == from || c == to || c == last {
			continue
		}
		last = c
		if ix.wallTop[c.Z*g.Width+c.X] > eye+(aim-eye)*t {
			return false
		}
	}
	return true
}
