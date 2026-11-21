package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "pantano",
		Name:        "Pântano",
		Description: "Poças escuras, lama e neblina de noite. Passarelas podres ligam ilhotas: uma cabana sobre palafitas com luz acesa, um atracadouro com barcos e um esconderijo de caixotes. Algumas trilhas levam a lugar nenhum.",
		Tags:        []string{"natureza", "perigo", "refúgio"},
		Width:       64, Depth: 64,
		Ambience: Ambience{TimeOfDay: "night", Weather: "fog"},
		Generate: generatePantano,
	})
}

func generatePantano(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(64, 64, seed, "grass", 1)
	pantanoGround(b)
	pantanoIslets(b)
	pantanoWalkways(b)
	pantanoHut(b)
	pantanoDock(b)
	pantanoStash(b)
	pantanoScatter(b)
	b.ClearEntry()
	return b.Build()
}

func pantanoGround(b *Builder) {
	for z := 0; z < b.D; z++ {
		for x := 0; x < b.W; x++ {
			switch n := b.Noise.FBM(float64(x)/7+3, float64(z)/7+8, 3); {
			case n < 0.44:
				b.Set(x, z, "water", 0)
			case n < 0.5:
				b.Set(x, z, "dirt", 1)
			default:
				b.Set(x, z, "grass", 1)
			}
		}
	}
}

func pantanoIslets(b *Builder) {
	land := func(cx, cz int, r float64, h int) {
		b.Blob(cx, cz, r+3, 0.15, Brush{Terrain: "water", Height: 0, Reserve: true})
		b.Blob(cx, cz, r, 0.2, Brush{Terrain: "dirt", Height: h, Reserve: true})
	}
	land(32, 20, 7, 2)
	land(52, 26, 4, 1)
	land(10, 14, 3, 1)
	land(14, 40, 3.5, 1)
	b.FillCircle(32, 58, 7, Brush{Terrain: "dirt", Height: 1, Reserve: true})
	b.FillCircle(32, 50, 3, Brush{Terrain: "dirt", Height: 1, Reserve: true})
}

func pantanoWalkways(b *Builder) {
	b.Walkway([]Pt{{32, 56}, {32, 48}, {28, 42}, {26, 34}, {30, 28}, {32, 24}}, 2, 1)
	b.Walkway([]Pt{{32, 48}, {42, 44}, {50, 36}, {52, 28}}, 1, 1)
	b.Walkway([]Pt{{28, 42}, {20, 42}, {15, 41}}, 1, 1)
	b.Walkway([]Pt{{26, 34}, {16, 30}, {10, 22}, {10, 15}}, 1, 1)
	b.Walkway([]Pt{{32, 13}, {32, 6}}, 1, 1)
}

func pantanoHut(b *Builder) {
	hut := b.Building(Rect{29, 17, 7, 5}, BuildingOpts{Door: South, Ruin: 0.12, Open: true, Floor: "floor"})
	b.FurnishCorners(hut, "crate", 2)
	b.FurnishCorners(hut, "barrel", 1)
	b.Place("lamp_post", 28, 23, 0, nil)
	b.Place("campfire", 33, 24, 0, nil)
	b.Place("sign", 31, 26, 0, nil)
	b.Place("barrel", 37, 20, 0, nil)
}

func pantanoDock(b *Builder) {
	b.Bridge(Pt{52, 28}, Pt{52, 36}, 1, 1)
	b.Bridge(Pt{52, 36}, Pt{56, 36}, 1, 1)
	b.Place("lamp_post", 53, 24, 0, nil)
	b.Place("crate", 50, 25, 0, nil)
	b.Place("crate", 54, 27, 0, nil)
	b.Place("bench", 53, 33, 90, nil)
	b.Place("bench", 49, 32, 0, nil)
	b.Place("bench", 57, 38, 90, nil)
}

func pantanoStash(b *Builder) {
	b.Place("crate", 10, 13, 0, nil)
	b.Place("crate", 11, 15, 0, nil)
	b.Place("barrel", 9, 15, 0, nil)
	b.Place("campfire", 14, 40, 0, nil)
	b.Place("tent", 12, 38, 0, nil)
	b.Place("sign", 22, 41, 0, nil)
	b.Place("sign", 30, 8, 0, nil)
}

func pantanoScatter(b *Builder) {
	all := Rect{0, 0, 64, 64}
	b.Scatter(all, Scatter{Prefab: "tree", Density: 0.2, Clump: 6, Variants: []string{"pine"}, Terrains: []string{"grass", "dirt"}})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.3, Clump: 5, Terrains: []string{"grass", "dirt"}})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.015, Terrains: []string{"grass"}})
}
