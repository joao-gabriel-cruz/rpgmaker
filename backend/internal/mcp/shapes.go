package mcp

import (
	"errors"
	"sort"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

const (
	maxShapeCells   = 4096
	maxCircleRadius = 32
	maxLineWidth    = 5
	maxCoordinate   = 1000 // keeps line walking bounded for absurd input
)

// Point is a grid position as sent by the MCP client.
type Point struct {
	X int `json:"x"`
	Z int `json:"z"`
}

// Shape describes an area of the grid: rect, circle, line or an explicit cell list.
type Shape struct {
	Kind   string  `json:"kind"`
	From   *Point  `json:"from,omitempty"`
	To     *Point  `json:"to,omitempty"`
	Radius int     `json:"radius,omitempty"`
	Width  int     `json:"width,omitempty"`
	Cells  []Point `json:"cells,omitempty"`
}

// Rasterize returns the shape's cells clipped to a w×d grid, deduped, in row-major order.
func Rasterize(s Shape, w, d int) ([]maps.Coord, error) {
	pts, err := shapePoints(s, w, d)
	if err != nil {
		return nil, err
	}
	seen := map[maps.Coord]bool{}
	var out []maps.Coord
	for _, p := range pts {
		c := maps.Coord{X: p.X, Z: p.Z}
		if p.X < 0 || p.Z < 0 || p.X >= w || p.Z >= d || seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	if len(out) == 0 {
		return nil, errors.New("a forma fica toda fora do mapa")
	}
	if len(out) > maxShapeCells {
		return nil, errors.New("a forma cobre mais de 4096 células; use uma área menor")
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Z != out[j].Z {
			return out[i].Z < out[j].Z
		}
		return out[i].X < out[j].X
	})
	return out, nil
}

func shapePoints(s Shape, w, d int) ([]Point, error) {
	switch s.Kind {
	case "rect":
		return rectPoints(s, w, d)
	case "circle":
		return circlePoints(s)
	case "line":
		return linePoints(s)
	case "cells":
		return s.Cells, nil
	}
	return nil, errors.New("tipo de forma desconhecido; use rect, circle, line ou cells")
}

func checkPoint(p *Point, name string) error {
	if p == nil {
		return errors.New("a forma precisa do campo \"" + name + "\"")
	}
	if p.X < -maxCoordinate || p.X > maxCoordinate || p.Z < -maxCoordinate || p.Z > maxCoordinate {
		return errors.New("coordenada fora do intervalo permitido")
	}
	return nil
}

func rectPoints(s Shape, w, d int) ([]Point, error) {
	if err := checkPoint(s.From, "from"); err != nil {
		return nil, err
	}
	if err := checkPoint(s.To, "to"); err != nil {
		return nil, err
	}
	x0, x1 := max(min(s.From.X, s.To.X), 0), min(max(s.From.X, s.To.X), w-1)
	z0, z1 := max(min(s.From.Z, s.To.Z), 0), min(max(s.From.Z, s.To.Z), d-1)
	var pts []Point
	for z := z0; z <= z1; z++ {
		for x := x0; x <= x1; x++ {
			pts = append(pts, Point{x, z})
		}
	}
	return pts, nil
}

func circlePoints(s Shape) ([]Point, error) {
	if err := checkPoint(s.From, "from"); err != nil {
		return nil, err
	}
	if s.Radius < 0 || s.Radius > maxCircleRadius {
		return nil, errors.New("o raio deve ficar entre 0 e 32")
	}
	r := s.Radius
	var pts []Point
	for dz := -r; dz <= r; dz++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dz*dz <= r*r {
				pts = append(pts, Point{s.From.X + dx, s.From.Z + dz})
			}
		}
	}
	return pts, nil
}

func linePoints(s Shape) ([]Point, error) {
	if err := checkPoint(s.From, "from"); err != nil {
		return nil, err
	}
	if err := checkPoint(s.To, "to"); err != nil {
		return nil, err
	}
	width := s.Width
	if width == 0 {
		width = 1
	}
	if width < 1 || width > maxLineWidth {
		return nil, errors.New("a largura da linha deve ficar entre 1 e 5")
	}
	return thicken(bresenham(*s.From, *s.To), (width-1)/2, width%2 == 0), nil
}

func bresenham(a, b Point) []Point {
	dx, dz := abs(b.X-a.X), -abs(b.Z-a.Z)
	sx, sz := sign(b.X-a.X), sign(b.Z-a.Z)
	errAcc := dx + dz
	var pts []Point
	for {
		pts = append(pts, a)
		if a == b {
			return pts
		}
		e2 := 2 * errAcc
		if e2 >= dz {
			errAcc += dz
			a.X += sx
		}
		if e2 <= dx {
			errAcc += dx
			a.Z += sz
		}
	}
}

// thicken adds every cell within Chebyshev distance half of a line cell; even widths extend one extra cell toward +x/+z.
func thicken(line []Point, half int, even bool) []Point {
	if half == 0 && !even {
		return line
	}
	hi := half
	if even {
		hi++
	}
	var pts []Point
	for _, p := range line {
		for dz := -half; dz <= hi; dz++ {
			for dx := -half; dx <= hi; dx++ {
				pts = append(pts, Point{p.X + dx, p.Z + dz})
			}
		}
	}
	return pts
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}
