package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "festival",
		Name:        "Festival da Comunidade",
		Description: "Noite de festa na praça: barracas de feira coloridas, uma grande fogueira para dançar, bandeirolas, postes de luz e um palco improvisado. Boa hora para negociar, conspirar... ou para algo dar errado.",
		Tags:        []string{"social", "urbano", "refúgio"},
		Width:       64, Depth: 64,
		Ambience: Ambience{TimeOfDay: "night", Weather: "clear"},
		Generate: generateFestival,
	})
}

var festivalColors = []string{"red", "blue", "green", "yellow", "orange", "purple"}

var festivalPlaza = Rect{12, 12, 40, 36}

func generateFestival(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(64, 64, seed, "grass", 1)
	festivalGround(b)
	festivalStage(b)
	festivalStalls(b)
	festivalDanceFloor(b)
	festivalLights(b)
	festivalHouses(b)
	festivalVegetation(b)
	b.ClearEntry()
	return b.Build()
}

func festivalGround(b *Builder) {
	b.FillRect(festivalPlaza, Brush{Terrain: "dirt", Height: 1, Reserve: true})
	b.Road([]Pt{{32, 63}, {32, 48}}, 4, "dirt", 1)
	b.Road([]Pt{{2, 51}, {61, 51}}, 3, "dirt", 1)
	b.Road([]Pt{{7, 13}, {7, 50}}, 2, "dirt", 1)
	b.Road([]Pt{{56, 13}, {56, 50}}, 2, "dirt", 1)
}

// festivalStage is a raised platform north of the plaza with a ramp of steps, speakers and corner lights.
func festivalStage(b *Builder) {
	stage := Rect{22, 13, 20, 6}
	b.FillRect(stage.Grow(1), Brush{Terrain: "dirt", Height: 2, Reserve: true})
	b.FillRect(stage, Brush{Terrain: "floor", Height: 3, Reserve: true})
	for _, x := range []int{stage.X, stage.X + stage.W - 1} {
		b.Place("lamp_post", x, stage.Z, 0, nil)
		b.Place("lamp_post", x, stage.Z+stage.D-1, 0, nil)
	}
	for _, x := range []int{stage.X + 2, stage.X + stage.W - 3} {
		b.PlaceSpaced("crate", Pt{x, stage.Z + 1}, 0, 1, 2, 0, nil)
	}
	b.PlaceSpaced("barrel", Pt{stage.X + 8, stage.Z + 1}, 1, 0, 4, 0, nil)
	b.PlaceSpaced("bunting", Pt{stage.X + 1, stage.Z}, 2, 0, 9, 0, nil)
}

// festivalStalls lines the plaza with colored fair stalls leaving aisles between them.
func festivalStalls(b *Builder) {
	n := b.Rng.Intn(len(festivalColors))
	put := func(x, z int) {
		n++
		b.Place("market_stall", x, z, 0, map[string]any{"color": festivalColors[n%len(festivalColors)]})
		b.Place("barrel", x+b.Rng.Intn(2), z+2, 0, nil)
	}
	for _, z := range []int{22, 38} {
		for _, x := range []int{14, 18, 22, 40, 44, 48} {
			put(x, z)
		}
	}
	for _, x := range []int{14, 48} {
		for _, z := range []int{26, 30, 34} {
			put(x, z)
		}
	}
}

// festivalDanceFloor is the great central bonfire ringed by a dance floor and benches.
func festivalDanceFloor(b *Builder) {
	b.FillCircle(32, 30, 5, Brush{Terrain: "floor", Height: 1, Reserve: true})
	b.FillCircle(32, 30, 1, Brush{Terrain: "stone", Height: 1, Reserve: true})
	b.Place("campfire", 32, 30, 0, nil)
	for _, p := range []Pt{{31, 24}, {31, 36}} {
		b.Place("bench", p.X, p.Z, 0, nil)
	}
	for _, p := range []Pt{{25, 29}, {38, 29}} {
		b.Place("bench", p.X, p.Z, 90, nil)
	}
	b.Place("campfire", 26, 42, 0, nil)
	b.Place("campfire", 37, 42, 0, nil)
	b.PlaceSpaced("barrel", Pt{29, 42}, 1, 0, 6, 0, nil)
}

// festivalLights hangs bunting between lamp posts and lights the access roads.
func festivalLights(b *Builder) {
	for _, x := range []int{12, 24, 39, 51} {
		b.Place("lamp_post", x, 20, 0, nil)
		b.Place("lamp_post", x, 42, 0, nil)
	}
	for _, z := range []int{26, 34} {
		b.Place("lamp_post", 24, z, 0, nil)
		b.Place("lamp_post", 39, z, 0, nil)
	}
	b.PlaceSpaced("bunting", Pt{25, 20}, 2, 0, 7, 0, nil)
	b.PlaceSpaced("bunting", Pt{25, 42}, 2, 0, 7, 0, nil)
	b.PlaceSpaced("bunting", Pt{25, 26}, 2, 0, 7, 0, nil)
	b.PlaceSpaced("bunting", Pt{25, 34}, 2, 0, 7, 0, nil)
	b.PlaceSpaced("lamp_post", Pt{29, 49}, 6, 0, 2, 0, nil)
	b.PlaceSpaced("lamp_post", Pt{34, 55}, 0, 4, 2, 0, nil)
	b.Place("sign", 30, 60, 0, nil)
}

// festivalHouses rings the plaza with lived-in houses facing their access roads.
func festivalHouses(b *Builder) {
	var doors = []struct {
		r Rect
		d Side
	}{
		{Rect{1, 14, 5, 7}, East}, {Rect{1, 24, 5, 7}, East}, {Rect{1, 34, 5, 7}, East},
		{Rect{57, 14, 6, 7}, West}, {Rect{57, 24, 6, 7}, West}, {Rect{57, 34, 6, 7}, West},
		{Rect{6, 54, 8, 7}, North}, {Rect{18, 54, 8, 7}, North}, {Rect{42, 54, 8, 7}, North}, {Rect{53, 54, 8, 7}, North},
	}
	for _, h := range doors {
		info := b.Building(h.r, BuildingOpts{Door: h.d, Open: b.Rng.Intn(3) != 0, Ruin: 0.05, Floor: "floor"})
		b.FurnishCorners(info, "crate", 1+b.Rng.Intn(2))
		b.Place("lamp_post", info.Front.X, info.Front.Z+sgnTo(info.Front.Z, 51), 0, nil)
	}
}

func festivalVegetation(b *Builder) {
	all := Rect{0, 0, 64, 64}
	b.Scatter(all, Scatter{Prefab: "tree", Density: 0.1, Clump: 8, Variants: []string{"round", "pine"}, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.05, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.008, Terrains: []string{"grass"}})
}
