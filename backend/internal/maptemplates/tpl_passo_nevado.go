package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "passo-nevado",
		Name:        "Passo Nevado",
		Description: "Duas cadeias de montanhas fecham um desfiladeiro coberto de neve. Há um posto abandonado no estreito, uma fenda com uma única ponte e uma cabana com fogueira num patamar alto, atrás de uma trilha íngreme.",
		Tags:        []string{"natureza", "perigo", "refúgio"},
		Width:       80, Depth: 80,
		Ambience: Ambience{TimeOfDay: "day", Weather: "snow"},
		Generate: generatePasso,
	})
}

const passoFissureZ = 58 // the crack that cuts the pass; one bridge crosses it

func generatePasso(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(80, 80, seed, "snow", 1)
	passoRelief(b)
	passoLake(b)
	passoLedges(b)
	passoTrail(b)
	passoFissure(b)
	passoCheckpoint(b)
	passoCabin(b)
	passoScatter(b)
	b.ClearEntry()
	b.SealOff("stone", 6)
	return b.Build()
}

func passoRelief(b *Builder) {
	for z := 0; z < b.D; z++ {
		for x := 0; x < b.W; x++ {
			b.Set(x, z, "snow", 1+int(b.Noise.FBM(float64(x)/8, float64(z)/8, 2)*2.5))
		}
	}
	west := []Pt{{3, 0}, {17, 30}, {13, 50}, {5, 79}}
	east := []Pt{{77, 0}, {63, 30}, {67, 50}, {75, 79}}
	for _, chain := range [][]Pt{west, east} {
		for i := 1; i < len(chain); i++ {
			b.Ridge(chain[i-1], chain[i], 13, 8, "")
		}
	}
	for z := 0; z < b.D; z++ {
		for x := 0; x < b.W; x++ {
			h := b.Cell(x, z).Height
			if h >= 8 || (h >= 4 && b.Noise.FBM(float64(x)/4+70, float64(z)/4, 2) > 0.6) {
				b.Set(x, z, "stone", Keep)
			}
		}
	}
}

func passoLake(b *Builder) {
	b.Blob(44, 14, 7, 0.2, Brush{Terrain: "water", Height: 1, Reserve: true})
	b.FillCircle(50, 19, 1.5, Brush{Terrain: "water", Height: 1, Reserve: true})
	b.Blob(44, 14, 10, 0.2, Brush{Terrain: "", Height: Keep, Reserve: true})
}

func passoFissure(b *Builder) {
	b.Line(Pt{0, passoFissureZ}, Pt{79, passoFissureZ}, 3, Brush{Terrain: "stone", Height: 0, Reserve: true})
	b.Bridge(Pt{40, passoFissureZ - 3}, Pt{40, passoFissureZ + 3}, 2, Keep)
}

// passoLedge cuts a ledge into the mountain wall: a stone rim around a flat snow platform.
func passoLedge(b *Builder, cx, cz, h int) {
	b.FillCircle(cx, cz, 7, Brush{Terrain: "stone", Height: Keep, Reserve: true})
	b.FillCircle(cx, cz, 5, Brush{Terrain: "snow", Height: h, Reserve: true})
}

func passoLedges(b *Builder) {
	passoLedge(b, 25, 42, 5)
	passoLedge(b, 57, 66, 4)
}

func passoTrail(b *Builder) {
	b.Road([]Pt{{40, 78}, {40, 66}, {40, 56}, {40, 48}, {38, 38}, {36, 28}, {32, 18}, {31, 6}, {31, 0}}, 3, "dirt", Keep)
	b.Ramp(Pt{28, 41}, Pt{36, 44}, 2, "dirt")
	b.Ramp(Pt{54, 66}, Pt{44, 66}, 2, "dirt")
}

func passoCheckpoint(b *Builder) {
	for _, x := range []int{33, 34, 35, 45, 46, 47} {
		b.Place("barricade", x, 36, 0, nil)
	}
	for _, p := range []Pt{{37, 35}, {43, 35}, {37, 37}} {
		b.Place("sandbags", p.X, p.Z, 0, nil)
	}
	b.Place("car", 44, 40, 90, nil)
	b.Place("tent", 46, 43, 0, nil)
	b.Place("crate", 44, 43, 0, nil)
	b.Place("post", 41, 35, 0, nil)
	b.Place("sign", 39, 62, 0, nil)
}

func passoCabin(b *Builder) {
	info := b.Building(Rect{22, 39, 6, 5}, BuildingOpts{Door: East, Floor: "floor", Open: true})
	b.FurnishCorners(info, "crate", 2)
	b.Place("campfire", 29, 43, 0, nil)
	b.Place("barrel", 28, 38, 0, nil)
	b.Place("bench", 27, 45, 0, nil)
	b.Place("tent", 56, 64, 0, nil)
	b.Place("crate", 59, 63, 0, nil)
	b.Place("crate", 59, 67, 0, nil)
	b.Place("barrel", 55, 68, 0, nil)
}

func passoScatter(b *Builder) {
	all := Rect{0, 0, 80, 80}
	b.Scatter(all, Scatter{Prefab: "tree", Density: 0.2, Clump: 6, Variants: []string{"pine"}, Terrains: []string{"snow"}, MaxHeight: 6})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.05, Clump: 5, Terrains: []string{"snow", "dirt"}, MaxHeight: 7})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.015, Terrains: []string{"snow"}, MaxHeight: 4})
}
