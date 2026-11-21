package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "litoral-pier",
		Name:        "Litoral e Píer",
		Description: "O mar bate numa praia de dunas, com barracos de pescador, barcos virados, um casco encalhado cheio de caixotes e um píer longo que avança sobre a água. Um farol velho vigia a ponta norte.",
		Tags:        []string{"natureza", "social", "perigo"},
		Width:       64, Depth: 96,
		Ambience: Ambience{TimeOfDay: "day", Weather: "clear"},
		Generate: generateLitoral,
	})
}

// litoralShore is the x of the first water cell on row z (the sea is on the east side).
func litoralShore(b *Builder, z int) int {
	return 46 + int((b.Noise.FBM(float64(z)/14, 4.5, 2)-0.5)*12)
}

func generateLitoral(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(64, 96, seed, "grass", 1)
	litoralLand(b)
	litoralRoad(b)
	litoralPiers(b)
	litoralShacks(b)
	litoralWreck(b)
	litoralInland(b)
	litoralLighthouse(b)
	litoralBeach(b)
	litoralScatter(b)
	b.ClearEntry()
	return b.Build()
}

func litoralLand(b *Builder) {
	for z := 0; z < b.D; z++ {
		sx := litoralShore(b, z)
		for x := 0; x < b.W; x++ {
			switch {
			case x >= sx:
				b.Set(x, z, "water", 0)
			case x >= sx-7:
				b.Set(x, z, "sand", 1)
			default:
				b.Set(x, z, "grass", 1+int(b.Noise.FBM(float64(x)/9, float64(z)/9, 2)*2.5))
			}
		}
	}
	for i := 0; i < 7; i++ {
		z := 6 + b.Rng.Intn(84)
		b.Hill(litoralShore(b, z)-10-b.Rng.Intn(8), z, float64(5+b.Rng.Intn(3)), 2+b.Rng.Intn(2), "sand")
	}
}

func litoralRoad(b *Builder) {
	b.Road([]Pt{{32, 95}, {31, 84}, {27, 70}, {24, 52}, {25, 36}, {27, 20}, {30, 8}, {31, 0}}, 3, "dirt", Keep)
}

func litoralPiers(b *Builder) {
	z := 40
	sx := litoralShore(b, z)
	b.Bridge(Pt{sx - 6, z}, Pt{sx + 11, z}, 3, -1)
	b.Bridge(Pt{sx + 11, z - 4}, Pt{sx + 11, z + 4}, 3, -1)
	b.Line(Pt{25, z}, Pt{sx - 7, z}, 2, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
	z = 68
	sx = litoralShore(b, z)
	b.Bridge(Pt{sx - 5, z}, Pt{sx + 8, z}, 2, -1)
	b.Line(Pt{26, z}, Pt{sx - 6, z}, 2, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
	b.Place("sign", sx-8, z-2, 0, nil)
}

func litoralShacks(b *Builder) {
	shacks := []struct {
		z    int
		dx   int
		door Side
	}{{31, 18, East}, {47, 17, North}, {62, 15, East}, {75, 16, East}}
	for _, s := range shacks {
		sx := litoralShore(b, s.z)
		info := b.Building(Rect{sx - s.dx, s.z, 5, 4}, BuildingOpts{Door: s.door, Ruin: b.Rng.Float64() * 0.25, Open: true, Floor: "floor"})
		b.FurnishCorners(info, "crate", 1+b.Rng.Intn(2))
		b.Place("barrel", info.Front.X, info.Front.Z+1, 0, nil)
	}
}

func litoralWreck(b *Builder) {
	z := 14
	sx := litoralShore(b, z)
	hull := b.Building(Rect{sx - 12, z, 11, 5}, BuildingOpts{NoDoor: true, Ruin: 0.3, Floor: "floor", Material: "ruined_wall"})
	b.FurnishCorners(hull, "crate", 3)
	b.FurnishCorners(hull, "barrel", 2)
	b.Place("campfire", sx-14, z+2, 0, nil)
}

func litoralLighthouse(b *Builder) {
	cx, cz := 44, 5
	b.Hill(cx, cz, 9, 4, "dirt")
	b.FillCircle(cx, cz, 2, Brush{Terrain: "dirt", Height: b.Cell(cx, cz).Height, Reserve: true})
	b.Place("watchtower", cx, cz, 0, nil)
	b.Line(Pt{31, 8}, Pt{cx - 2, cz + 2}, 2, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
}

func litoralBeach(b *Builder) {
	for i := 0; i < 10; i++ {
		z := 8 + b.Rng.Intn(80)
		if z > 35 && z < 45 || z > 63 && z < 73 {
			continue
		}
		rot := 90 * b.Rng.Intn(2)
		b.Place("bench", litoralShore(b, z)-3-b.Rng.Intn(3), z, rot, nil)
	}
	for _, z := range []int{26, 54, 84} {
		b.Place("crate", litoralShore(b, z)-3, z, 0, nil)
		b.Place("barrel", litoralShore(b, z)-4, z+1, 0, nil)
	}
	b.Place("car", 20, 58, 0, nil)
	b.Place("car", 29, 12, 90, nil)
}

func litoralScatter(b *Builder) {
	all := Rect{0, 0, 64, 96}
	b.Scatter(all, Scatter{Prefab: "tree", Density: 0.14, Clump: 8, Variants: []string{"round"}, Terrains: []string{"sand", "grass"}})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.05, Terrains: []string{"grass", "sand"}})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.03, Terrains: []string{"sand"}})
}

// litoralInland adds a flooded warehouse, a farmhouse and a survivors' camp west of the coastal road.
func litoralInland(b *Builder) {
	house := b.Building(Rect{6, 44, 11, 8}, BuildingOpts{Door: East, Ruin: 0.25, Floor: "floor"})
	b.FurnishCorners(house, "crate", 3)
	b.FurnishCorners(house, "barrel", 1)
	farm := b.Building(Rect{8, 20, 6, 5}, BuildingOpts{Door: South, Ruin: 0.2, Open: true, Floor: "floor"})
	b.FurnishCorners(farm, "crate", 1)
	for _, in := range []BuildingInfo{house, farm} {
		b.Line(in.Front, Pt{25, in.Front.Z}, 1, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
	}
	b.Line(Pt{farm.Front.X, farm.Front.Z}, Pt{farm.Front.X, farm.Front.Z + 3}, 1, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
	b.Place("campfire", 14, 72, 0, nil)
	b.Place("tent", 11, 70, 0, nil)
	b.Place("tent", 17, 74, 0, nil)
	b.Place("crate", 13, 75, 0, nil)
	b.FenceRect(Rect{4, 28, 10, 8}, South)
	b.Place("well", 9, 32, 0, nil)
}
