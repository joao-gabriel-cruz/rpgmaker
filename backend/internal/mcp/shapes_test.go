package mcp

import (
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

func pt(x, z int) *Point { return &Point{X: x, Z: z} }

func TestRasterizeCounts(t *testing.T) {
	cases := []struct {
		name string
		s    Shape
		w, d int
		want int
	}{
		{"rect", Shape{Kind: "rect", From: pt(2, 3), To: pt(4, 4)}, 10, 10, 6},
		{"rect swapped corners", Shape{Kind: "rect", From: pt(4, 4), To: pt(2, 3)}, 10, 10, 6},
		{"circle r1", Shape{Kind: "circle", From: pt(5, 5), Radius: 1}, 10, 10, 5},
		{"circle r0", Shape{Kind: "circle", From: pt(5, 5)}, 10, 10, 1},
		{"line width 1", Shape{Kind: "line", From: pt(0, 0), To: pt(4, 2)}, 10, 10, 5},
		{"circle clipped at corner", Shape{Kind: "circle", From: pt(0, 0), Radius: 3}, 10, 10, 11},
		{"cells deduped", Shape{Kind: "cells", Cells: []Point{{1, 1}, {1, 1}, {2, 2}, {99, 99}}}, 10, 10, 2},
		{"rect clipped to max", Shape{Kind: "rect", From: pt(0, 0), To: pt(69, 69)}, 64, 64, 4096},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Rasterize(c.s, c.w, c.d)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != c.want {
				t.Fatalf("got %d cells, want %d: %v", len(got), c.want, got)
			}
		})
	}
}

func TestRasterizeRectOrderAndSwap(t *testing.T) {
	a, _ := Rasterize(Shape{Kind: "rect", From: pt(2, 3), To: pt(4, 4)}, 10, 10)
	b, _ := Rasterize(Shape{Kind: "rect", From: pt(4, 4), To: pt(2, 3)}, 10, 10)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("order differs at %d: %v vs %v", i, a, b)
		}
	}
	if a[0] != (maps.Coord{X: 2, Z: 3}) || a[3] != (maps.Coord{X: 2, Z: 4}) {
		t.Fatalf("not row-major: %v", a)
	}
}

func TestRasterizeLineWidth3(t *testing.T) {
	thin, _ := Rasterize(Shape{Kind: "line", From: pt(0, 0), To: pt(4, 2)}, 10, 10)
	thick, err := Rasterize(Shape{Kind: "line", From: pt(0, 0), To: pt(4, 2), Width: 3}, 10, 10)
	if err != nil {
		t.Fatal(err)
	}
	set := map[maps.Coord]bool{}
	for _, c := range thick {
		set[c] = true
	}
	for _, c := range thin {
		for dx := -1; dx <= 1; dx++ {
			for dz := -1; dz <= 1; dz++ {
				n := maps.Coord{X: c.X + dx, Z: c.Z + dz}
				if n.X >= 0 && n.Z >= 0 && !set[n] {
					t.Fatalf("missing %v near %v", n, c)
				}
			}
		}
	}
	if len(thick) <= len(thin) {
		t.Fatal("width 3 should add cells")
	}
}

func TestRasterizeErrors(t *testing.T) {
	var cells []Point
	for i := 0; i < 4097; i++ {
		cells = append(cells, Point{i % 70, i / 70})
	}
	cases := []struct {
		name string
		s    Shape
		want string
		size int
	}{
		{"outside", Shape{Kind: "rect", From: pt(20, 20), To: pt(25, 25)}, "a forma fica toda fora do mapa", 10},
		{"unknown kind", Shape{Kind: "star"}, "tipo de forma", 64},
		{"missing to", Shape{Kind: "rect", From: pt(1, 1)}, "to", 64},
		{"missing from", Shape{Kind: "circle"}, "from", 64},
		{"radius", Shape{Kind: "circle", From: pt(1, 1), Radius: 33}, "raio", 64},
		{"width", Shape{Kind: "line", From: pt(1, 1), To: pt(2, 2), Width: 6}, "largura", 64},
		{"too many", Shape{Kind: "cells", Cells: cells}, "4096", 70},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Rasterize(c.s, c.size, c.size)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want containing %q", err, c.want)
			}
		})
	}
}
