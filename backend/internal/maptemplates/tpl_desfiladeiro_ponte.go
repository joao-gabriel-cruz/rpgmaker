package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "desfiladeiro-ponte",
		Name:        "Desfiladeiro da Ponte",
		Description: "Um cânion fundo, com rio no fundo, corta o mapa. A única ponte liga as duas bordas, cheias de rochas e arbustos para emboscada. Há um acampamento na borda sul e uma trilha íngreme que desce até o rio, onde alguém escondeu suprimentos.",
		Tags:        []string{"natureza", "combate", "perigo"},
		Width:       64, Depth: 96,
		Ambience: Ambience{TimeOfDay: "day", Weather: "clear"},
		Generate: generateDesfiladeiro,
	})
}

const (
	canyonRim    = 7 // typical height of the edges
	canyonBridge = 32
)

// canyonZ is the row of the river center at column x.
func canyonZ(b *Builder, x int) int {
	return 46 + int((b.Noise.FBM(float64(x)/18, 7.7, 2)-0.5)*14)
}

func generateDesfiladeiro(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(64, 96, seed, "grass", canyonRim)
	canyonCarve(b)
	canyonPaths(b)
	canyonBridgeAndPosts(b)
	canyonCamp(b)
	canyonOutpost(b)
	canyonRiverbed(b)
	canyonAmbush(b)
	b.ClearEntry()
	return b.Build()
}

func canyonCarve(b *Builder) {
	for x := 0; x < b.W; x++ {
		cz := canyonZ(b, x)
		for z := 0; z < b.D; z++ {
			d := absInt(z - cz)
			switch {
			case d <= 2:
				b.Set(x, z, "water", 0)
			case d <= 5:
				b.Set(x, z, "sand", 1)
			case d <= 9:
				b.Set(x, z, "stone", maps.MaxHeight)
			default:
				t := "grass"
				if b.Noise.FBM(float64(x)/5+20, float64(z)/5, 2) > 0.58 {
					t = "dirt"
				}
				b.Set(x, z, t, canyonRim-1+int(b.Noise.FBM(float64(x)/8, float64(z)/8, 2)*2.5))
			}
		}
	}
}

func canyonPaths(b *Builder) {
	cz := canyonZ(b, canyonBridge)
	b.Road([]Pt{{32, 94}, {31, 80}, {32, 66}, {32, cz + 15}, {32, cz + 11}}, 3, "dirt", Keep)
	b.Road([]Pt{{32, cz - 11}, {32, cz - 17}, {34, 20}, {32, 4}, {32, 0}}, 3, "dirt", Keep)
	a, c := Pt{10, canyonZ(b, 10) + 11}, Pt{16, canyonZ(b, 16) + 4}
	b.Road([]Pt{{32, cz + 16}, {20, a.Z + 3}, {a.X + 1, a.Z}}, 2, "dirt", Keep)
	b.Ramp(a, c, 2, "dirt")
}

func canyonBridgeAndPosts(b *Builder) {
	cz := canyonZ(b, canyonBridge)
	b.Bridge(Pt{canyonBridge - 1, cz - 12}, Pt{canyonBridge - 1, cz + 12}, 2, canyonRim)
	for _, z := range []int{cz - 13, cz + 13} {
		b.Place("post", canyonBridge-3, z, 0, nil)
		b.Place("post", canyonBridge+2, z, 0, nil)
	}
	b.Place("sandbags", canyonBridge-3, cz-15, 0, nil)
	b.Place("sandbags", canyonBridge+2, cz-15, 0, nil)
	b.Place("barricade", canyonBridge-4, cz-16, 0, nil)
	b.Place("sign", canyonBridge+3, cz+15, 0, nil)
}

func canyonCamp(b *Builder) {
	cz := canyonZ(b, 48)
	cx, z := 46, cz+18
	b.FillCircle(cx, z, 5, Brush{Terrain: "dirt", Height: canyonRim, Reserve: true})
	b.Place("campfire", cx, z, 0, nil)
	b.Place("tent", cx-4, z-2, 0, nil)
	b.Place("tent", cx+3, z-3, 0, nil)
	b.Place("bench", cx-1, z+2, 0, nil)
	b.Place("crate", cx+2, z+2, 0, nil)
	b.Place("barrel", cx+3, z+1, 0, nil)
	b.Line(Pt{cx - 6, z}, Pt{34, z}, 1, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
}

func canyonOutpost(b *Builder) {
	cz := canyonZ(b, 14)
	info := b.Building(Rect{10, cz - 21, 7, 6}, BuildingOpts{Door: East, Ruin: 0.4, Floor: "floor"})
	b.FurnishCorners(info, "crate", 2)
	b.Line(info.Front, Pt{32, info.Front.Z}, 1, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
	b.Place("campfire", info.Front.X+2, info.Front.Z+2, 0, nil)
}

func canyonRiverbed(b *Builder) {
	cz := canyonZ(b, 36)
	b.Place("campfire", 37, cz+3, 0, nil)
	b.Place("crate", 38, cz+4, 0, nil)
	b.Place("crate", 36, cz+4, 0, nil)
	b.Place("barrel", 39, cz+3, 0, nil)
	b.Scatter(Rect{0, 0, 64, 96}, Scatter{Prefab: "rock", Density: 0.05, Terrains: []string{"sand"}})
}

func canyonAmbush(b *Builder) {
	for _, side := range []int{-1, 1} {
		for x := 4; x < 60; x++ {
			if absInt(x-canyonBridge) < 4 {
				continue
			}
			z := canyonZ(b, x) + side*(10+b.Rng.Intn(3))
			if b.Rng.Float64() < 0.65 {
				b.Place([]string{"rock", "bush", "bush"}[b.Rng.Intn(3)], x, z, 0, nil)
			}
		}
	}
	all := Rect{0, 0, 64, 96}
	b.Scatter(all, Scatter{Prefab: "tree", Density: 0.07, Clump: 7, Variants: []string{"pine", "round"}, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.05, Terrains: []string{"grass", "dirt"}})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.02, Terrains: []string{"grass", "dirt"}})
}
