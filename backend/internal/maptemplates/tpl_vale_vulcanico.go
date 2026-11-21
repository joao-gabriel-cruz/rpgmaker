package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "vale-vulcanico",
		Name:        "Vale Vulcânico",
		Description: "Um vulcão ainda ativo cospe rios de lava sobre um vale de cinzas. Uma trilha estreita cruza a lava até as ruínas de uma estação de pesquisa e um mirante no alto. Cada passo em falso queima.",
		Tags:        []string{"natureza", "perigo", "combate"},
		Width:       96, Depth: 96,
		Ambience: Ambience{TimeOfDay: "dusk", Weather: "ash"},
		Generate: generateVale,
	})
}

const (
	valeVolcano = 28 // volcano center is (valeVolcano, valeVolcano-4)
	valeCrossZ  = 72 // the lava river that splits the valley runs near this row
)

func generateVale(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(96, 96, seed, "dirt", 2)
	valeRelief(b)
	valeLava(b)
	valePaths(b)
	valeStation(b)
	valeShelter(b)
	valeWreck(b)
	valeScatter(b)
	b.ClearEntry()
	return b.Build()
}

func valeRelief(b *Builder) {
	for z := 0; z < b.D; z++ {
		for x := 0; x < b.W; x++ {
			b.Set(x, z, "dirt", 1+int(b.Noise.FBM(float64(x)/10, float64(z)/10, 3)*4.2))
		}
	}
	for i := 0; i < 12; i++ {
		b.Hill(8+b.Rng.Intn(80), 8+b.Rng.Intn(80), float64(6+b.Rng.Intn(6)), 2+b.Rng.Intn(2), "")
	}
	b.Hill(valeVolcano, valeVolcano-4, 26, 8, "dirt")
	b.Hill(82, 14, 10, 7, "dirt")
	valeRidges(b)
	b.Blob(68, 50, 15, 0.1, Brush{Terrain: "dirt", Height: 3})
	for i := 0; i < 14; i++ {
		b.Blob(6+b.Rng.Intn(84), 6+b.Rng.Intn(80), 2+b.Rng.Float64()*1.5, 0.35, Brush{Terrain: "stone", Height: Keep})
	}
}

func valeLava(b *Builder) {
	vx, vz := valeVolcano, valeVolcano-4
	top := b.Cell(vx+8, vz).Height
	b.FillCircle(vx, vz, 8, Brush{Terrain: "dirt", Height: top, Reserve: true})
	b.Blob(vx, vz, 7, 0.12, Brush{Terrain: "lava", Height: top - 1, Reserve: true})
	b.FillCircle(vx, vz, 2.5, Brush{Terrain: "dirt", Height: top, Reserve: true})
	b.Place("volcano", vx-1, vz-1, 0, nil)
	b.LavaFlow(Pt{vx + 4, vz + 4}, Pt{22, 52}, 4, 0.7)
	b.LavaFlow(Pt{22, 52}, Pt{12, 70}, 4, 0.5)
	b.LavaFlow(Pt{vx + 6, vz}, Pt{62, 32}, 4, 0.6)
	b.LavaFlow(Pt{62, 32}, Pt{94, 36}, 4, 0.5)
	b.LavaFlow(Pt{2, valeCrossZ}, Pt{48, valeCrossZ + 2}, 3, 0.5)
	b.LavaFlow(Pt{48, valeCrossZ + 2}, Pt{94, valeCrossZ - 2}, 3, 0.5)
	for _, p := range []Pt{{86, 56}, {60, 10}, {8, 40}, {40, 40}} {
		b.Blob(p.X, p.Z, 2.5+b.Rng.Float64()*1.5, 0.25, Brush{Terrain: "lava", Height: 1, Reserve: true})
	}
}

// valeRidges raises stone walls across the valley; the paths carve their way through later.
func valeRidges(b *Builder) {
	for _, r := range [][2]Pt{{{58, 4}, {78, 24}}, {{2, 34}, {18, 42}}, {{66, 64}, {90, 60}}, {{28, 58}, {38, 50}}, {{4, 12}, {14, 4}}} {
		b.Ridge(r[0], r[1], 3.5, 4, "stone")
	}
}

func valePaths(b *Builder) {
	dirt := Keep
	b.Road([]Pt{{48, 92}, {48, 84}}, 3, "dirt", dirt)
	b.Ramp(Pt{48, 84}, Pt{47, 64}, 3, "dirt") // causeway across the cross lava river
	b.Road([]Pt{{47, 64}, {46, 58}, {52, 52}, {60, 48}}, 3, "dirt", dirt)
	b.Ramp(Pt{20, 82}, Pt{21, 62}, 1, "dirt") // the risky one-cell shortcut
	b.Road([]Pt{{48, 82}, {70, 86}, {82, 84}}, 2, "dirt", dirt)
	b.Ramp(Pt{70, 44}, Pt{80, 22}, 2, "dirt") // crosses the eastern lava river toward the lookout
	b.Road([]Pt{{80, 22}, {82, 16}}, 2, "dirt", dirt)
	b.FillCircle(82, 14, 2, Brush{Terrain: "dirt", Height: b.Cell(82, 14).Height, Reserve: true})
	b.Place("watchtower", 82, 14, 0, nil)
	b.Place("sign", 49, 78, 0, nil)
}

func valeStation(b *Builder) {
	main := b.Building(Rect{58, 41, 13, 9}, BuildingOpts{Door: South, Ruin: 0.3, Floor: "floor"})
	dorm := b.Building(Rect{75, 44, 8, 7}, BuildingOpts{Door: West, Ruin: 0.5, Floor: "floor"})
	shed := b.Building(Rect{59, 55, 6, 5}, BuildingOpts{Door: North, Ruin: 0.2, Floor: "floor"})
	b.FurnishCorners(main, "crate", 3)
	b.FurnishCorners(dorm, "crate", 2)
	b.FurnishCorners(shed, "barrel", 3)
	for _, in := range []BuildingInfo{main, dorm, shed} {
		b.Line(in.Front, Pt{in.Front.X, 52}, 1, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
	}
	for i := 0; i < 4; i++ {
		b.Place("ruined_wall", 70+i*2, 58, 0, map[string]any{"height": 1 + i%2})
	}
	b.Place("campfire", 68, 53, 0, nil)
	b.Place("sandbags", 55, 50, 0, nil)
	b.Place("sandbags", 55, 52, 0, nil)
	b.Place("barrel", 73, 49, 0, nil)
	b.Place("crate", 66, 57, 0, nil)
}

func valeShelter(b *Builder) {
	cx, cz := 80, 84
	b.FillCircle(cx, cz, 5, Brush{Terrain: "stone", Height: 3, Reserve: true})
	b.FillCircle(cx, cz, 3, Brush{Terrain: "dirt", Height: 2, Reserve: true})
	b.Line(Pt{cx - 3, cz}, Pt{cx - 6, cz}, 2, Brush{Terrain: "dirt", Height: 2, Reserve: true})
	b.Place("campfire", cx, cz, 0, nil)
	b.Place("crate", cx+2, cz-1, 0, nil)
	b.Place("barrel", cx+1, cz+2, 0, nil)
	b.Place("crate", cx+2, cz+1, 0, nil)
}

func valeScatter(b *Builder) {
	all := Rect{0, 0, 96, 96}
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.08, Clump: 7, Terrains: []string{"dirt"}, MaxHeight: 9})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.02, Terrains: []string{"dirt"}, MaxHeight: 4})
}

func valeWreck(b *Builder) {
	b.Place("car", 14, 86, 0, nil)
	b.Place("car", 24, 82, 90, nil)
	b.Place("crate", 17, 84, 0, nil)
	b.Place("crate", 18, 85, 0, nil)
	b.Place("barrel", 22, 85, 0, nil)
	b.Place("campfire", 20, 84, 0, nil)
	b.Place("sign", 27, 79, 0, nil)
	b.Place("rock", 12, 80, 0, nil)
	b.Road([]Pt{{20, 84}, {34, 88}, {47, 90}}, 1, "dirt", Keep)
}
