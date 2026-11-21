package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "cidade-em-ruinas",
		Name:        "Cidade em Ruínas",
		Description: "Quarteirões de prédios caídos cortados por ruas largas, becos estreitos e barricadas nos cruzamentos. Uma praça com poço no centro e um prédio alto para vigiar tudo. Cada esquina esconde loot ou emboscada.",
		Tags:        []string{"urbano", "combate", "perigo"},
		Width:       96, Depth: 96,
		Ambience: Ambience{TimeOfDay: "day", Weather: "ash"},
		Generate: generateCidade,
	})
}

// The city is a 4x4 grid of 18-cell blocks separated by 4-cell streets (street i starts at 2+22*i).
const (
	cityBlock  = 18
	cityPitch  = 22
	cityOrigin = 6
)

func generateCidade(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(96, 96, seed, "dirt", 1)
	cityStreets(b)
	cityPlaza(b)
	for bi := 0; bi < 4; bi++ {
		for bj := 0; bj < 4; bj++ {
			cityLot(b, bi, bj)
		}
	}
	cityTraffic(b)
	cityCheckpoints(b)
	cityDebris(b)
	b.ClearEntry()
	return b.Build()
}

func cityStreets(b *Builder) {
	for i := 0; i < 5; i++ {
		s := 2 + cityPitch*i
		b.FillRect(Rect{s, 0, 4, 96}, Brush{Terrain: "floor", Height: 1, Reserve: true})
		b.FillRect(Rect{0, s, 96, 4}, Brush{Terrain: "floor", Height: 1, Reserve: true})
	}
}

// cityPlaza clears the four central blocks into a square with a well, benches and street lamps.
func cityPlaza(b *Builder) {
	plaza := Rect{32, 32, 32, 32}
	b.FillRect(Rect{28, 28, 40, 40}, Brush{Terrain: "dirt", Height: 1, Reserve: true})
	b.FillRect(plaza, Brush{Terrain: "floor", Height: 1, Reserve: true})
	b.FillCircle(48, 48, 4, Brush{Terrain: "stone", Height: 2, Reserve: true})
	b.FillCircle(48, 48, 2, Brush{Terrain: "floor", Height: 2, Reserve: true})
	b.Place("well", 48, 48, 0, nil)
	b.Place("campfire", 48, 52, 0, nil)
	for _, p := range []Pt{{44, 44}, {52, 44}, {44, 52}, {52, 52}} {
		b.Place("lamp_post", p.X, p.Z, 0, nil)
	}
	for _, p := range []Pt{{40, 41}, {55, 41}, {40, 56}, {55, 56}} {
		b.Place("bench", p.X, p.Z, 0, nil)
	}
	for _, p := range []Pt{{34, 34}, {61, 34}, {34, 61}, {61, 61}} {
		b.Place("tree", p.X, p.Z, 0, map[string]any{"variant": "round"})
		b.Place("sandbags", p.X+sgnTo(p.X, 48), p.Z, 0, nil)
	}
	for _, p := range []Pt{{36, 48}, {59, 47}, {47, 37}} {
		b.Place("car", p.X, p.Z, 90*b.Rng.Intn(2), nil)
	}
	b.Place("crate", 50, 53, 0, nil)
	b.Place("barrel", 51, 53, 0, nil)
	cityPlazaCover(b)
}

// cityPlazaCover scatters cover for firefights: wrecks, sandbag nests and broken wall.
func cityPlazaCover(b *Builder) {
	for i := 0; i < 26; i++ {
		x, z := 34+b.Rng.Intn(27), 34+b.Rng.Intn(27)
		if absInt(x-48)+absInt(z-48) < 9 {
			continue
		}
		switch b.Rng.Intn(4) {
		case 0:
			b.Place("car", x, z, 90*b.Rng.Intn(2), nil)
		case 1:
			b.PlaceSpaced("sandbags", Pt{x, z}, 1, 0, 3, 0, nil)
		case 2:
			b.PlaceSpaced("ruined_wall", Pt{x, z}, 0, 1, 2, 0, map[string]any{"height": 2})
		default:
			b.Place("barricade", x, z, 0, nil)
		}
	}
}

func sgnTo(v, center int) int {
	if v < center {
		return 1
	}
	return -1
}

func cityLot(b *Builder, bi, bj int) {
	bx, bz := cityOrigin+cityPitch*bi, cityOrigin+cityPitch*bj
	if (bi == 1 || bi == 2) && (bj == 1 || bj == 2) {
		return
	}
	if bi == 3 && bj == 0 {
		cityTower(b, bx, bz)
		return
	}
	switch b.Rng.Intn(7) {
	case 0, 1, 2:
		citySplit(b, bx, bz, true)
	case 3, 4:
		citySplit(b, bx, bz, false)
	case 5:
		cityBuilding(b, Rect{bx + 1, bz + 1, 16, 16}, b.sideFrom(North, 4))
	default:
		cityCollapsed(b, bx, bz)
	}
}

// citySplit puts two buildings in the block with a 2-cell alley between them.
func citySplit(b *Builder, bx, bz int, alongX bool) {
	w1 := 6 + b.Rng.Intn(3)
	w2 := 14 - w1
	d1, d2 := 11+b.Rng.Intn(6), 11+b.Rng.Intn(6)
	if alongX {
		cityBuilding(b, Rect{bx + 1, bz + 1, w1, d1}, b.sideFrom(North, 2))
		cityBuilding(b, Rect{bx + 3 + w1, bz + 1, w2, d2}, b.sideFrom(North, 2))
		return
	}
	cityBuilding(b, Rect{bx + 1, bz + 1, d1, w1}, b.sideFrom(East, 2))
	cityBuilding(b, Rect{bx + 1, bz + 3 + w1, d2, w2}, b.sideFrom(East, 2))
}

func cityBuilding(b *Builder, r Rect, door Side) {
	mat := "wall"
	if b.Rng.Intn(3) == 0 {
		mat = "ruined_wall"
	}
	info := b.Building(r, BuildingOpts{Door: door, Floor: "floor", Open: b.Rng.Intn(4) != 0, Ruin: b.Rng.Float64() * 0.45, Material: mat})
	b.FurnishCorners(info, []string{"crate", "barrel"}[b.Rng.Intn(2)], 1+b.Rng.Intn(3))
	if b.Rng.Intn(3) == 0 {
		b.Place("campfire", r.X+r.W/2, r.Z+r.D/2, 0, nil)
	}
}

// cityTower raises a stepped platform with a tall building and a watchtower for overwatch.
func cityTower(b *Builder, bx, bz int) {
	r := Rect{bx + 3, bz + 3, 12, 12}
	b.FillRect(r.Grow(2), Brush{Terrain: "stone", Height: 2, Reserve: true})
	b.FillRect(r.Grow(1), Brush{Terrain: "stone", Height: 3, Reserve: true})
	b.FillRect(r, Brush{Terrain: "floor", Height: 4})
	info := b.Building(r, BuildingOpts{Door: South, Open: true, Floor: "floor"})
	b.Place("watchtower", r.X+r.W-2, r.Z+1, 0, nil)
	b.FurnishCorners(info, "crate", 2)
	b.Place("sandbags", r.X+1, r.Z+r.D-2, 0, nil)
	b.Place("ruined_wall", bx+1, bz+16, 0, map[string]any{"height": 2})
}

// cityCollapsed leaves a rubble lot: a ruin heap ringed by broken wall.
func cityCollapsed(b *Builder, bx, bz int) {
	c := Pt{bx + 9, bz + 9}
	b.Blob(c.X, c.Z, 4, 0.3, Brush{Terrain: "ruin", Height: 2, Reserve: true})
	b.Reserve(Rect{bx, bz, cityBlock, cityBlock})
	for i := 0; i < 14; i++ {
		x, z := bx+1+b.Rng.Intn(16), bz+1+b.Rng.Intn(16)
		if t := b.Cell(x, z).Terrain; t != "ruin" {
			b.Place("ruined_wall", x, z, 0, map[string]any{"height": 1 + b.Rng.Intn(2)})
		}
	}
	b.Place("crate", bx+3, bz+14, 0, nil)
	b.Place("barrel", bx+4, bz+14, 0, nil)
}

func cityTraffic(b *Builder) {
	for i := 0; i < 5; i++ {
		s := 2 + cityPitch*i
		for n := 0; n < 5; n++ {
			b.Place("car", 4+b.Rng.Intn(88), s+1+b.Rng.Intn(2), 0, nil)
			b.Place("car", s+1+b.Rng.Intn(2), 4+b.Rng.Intn(80), 90, nil)
		}
	}
}

// cityCheckpoints blocks some intersections with sandbags and barricades, always leaving a gap.
func cityCheckpoints(b *Builder) {
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if (i+j)%2 == 1 || i == 0 || i == 4 || j == 0 || j == 4 || b.Rng.Intn(4) == 0 {
				continue
			}
			x, z := 2+cityPitch*i, 2+cityPitch*j
			b.PlaceSpaced("sandbags", Pt{x + 5, z}, 0, 1, 2, 0, nil)
			b.Place("barricade", x+5, z+3, 0, nil)
			b.Place("sandbags", x-2, z, 0, nil)
			b.Place("barricade", x-2, z+2, 0, nil)
		}
	}
	b.Place("sign", 47, 80, 0, nil)
}

func cityDebris(b *Builder) {
	all := Rect{0, 0, 96, 96}
	b.Scatter(all, Scatter{Prefab: "ruined_wall", Density: 0.008, Props: map[string]any{"height": 1}})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.006})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.02, Terrains: []string{"dirt"}})
	b.Scatter(all, Scatter{Prefab: "crate", Density: 0.002, Terrains: []string{"floor"}})
}
