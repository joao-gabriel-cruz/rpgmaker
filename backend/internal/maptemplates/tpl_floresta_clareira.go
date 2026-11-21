package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "floresta-clareira",
		Name:        "Floresta Densa",
		Description: "Mata fechada cortada por trilhas sinuosas e um riacho com uma única ponte de tábuas. No fundo, uma clareira escondida com acampamento de sobreviventes. Ótimo para emboscadas, fugas e encontros às escondidas.",
		Tags:        []string{"natureza", "combate", "social"},
		Width:       64, Depth: 64,
		Ambience: Ambience{TimeOfDay: "day", Weather: "clear"},
		Generate: generateFloresta,
	})
}

func generateFloresta(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(64, 64, seed, "grass", 1)
	b.Hill(10+b.Rng.Intn(8), 12+b.Rng.Intn(8), 12, 3, "")
	b.Hill(48+b.Rng.Intn(8), 14+b.Rng.Intn(8), 10, 2, "")
	b.Hill(8+b.Rng.Intn(8), 48+b.Rng.Intn(6), 9, 2, "")
	cx := 28 + b.Rng.Intn(9)
	b.River(Pt{0, 36 + b.Rng.Intn(4)}, Pt{63, 34 + b.Rng.Intn(4)}, 3, 0.9, "sand")
	camp := Pt{16 + b.Rng.Intn(32), 11 + b.Rng.Intn(6)}
	florestaTrails(b, cx, camp)
	florestaCamp(b, camp)
	florestaCache(b, cx)
	florestaWoods(b)
	b.ClearEntry()
	return b.Build()
}

// florestaTrails links the entry to the camp, crossing the stream by a plank bridge at column cx.
func florestaTrails(b *Builder, cx int, camp Pt) {
	south := b.WindingPath(Pt{32, 60}, Pt{cx, 50}, 1.1)
	b.PaintTrail(south, 2, "dirt")
	var cross []Pt
	for z := 50; z >= 24; z-- {
		cross = append(cross, Pt{cx, z})
	}
	b.PaintTrail(cross, 2, "dirt")
	north := b.WindingPath(Pt{cx, 24}, camp, 1.2)
	b.PaintTrail(north, 2, "dirt")
	for z := 24; z <= 50; z++ {
		for _, x := range []int{cx, cx + 1} {
			if b.Cell(x, z).Terrain == "water" {
				b.Place("bridge", x, z, 0, map[string]any{"deck": 1})
			}
		}
	}
	b.PaintTrail(b.WindingPath(north[len(north)/2], Pt{4 + b.Rng.Intn(10), 4}, 1.2), 2, "dirt")
	b.PaintTrail(b.WindingPath(south[len(south)/2], Pt{58, 50 + b.Rng.Intn(8)}, 1.2), 2, "dirt")
}

// florestaCamp is the hidden clearing: tents around a fire, supplies and a lookout sign.
func florestaCamp(b *Builder, c Pt) {
	b.Blob(c.X, c.Z, 7, 0.2, Brush{Terrain: "grass", Height: Keep, Reserve: true})
	b.FillCircle(c.X, c.Z, 2, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
	b.Place("campfire", c.X, c.Z, 0, nil)
	for _, o := range []Pt{{-5, -3}, {4, -4}, {-5, 3}, {5, 2}} {
		b.Place("tent", c.X+o.X, c.Z+o.Z, 0, nil)
	}
	b.Place("bench", c.X-2, c.Z+2, 0, nil)
	b.Place("bench", c.X+1, c.Z-3, 0, nil)
	b.PlaceSpaced("barrel", Pt{c.X + 3, c.Z + 4}, 1, 0, 3, 0, nil)
	b.PlaceSpaced("crate", Pt{c.X - 3, c.Z - 5}, 1, 0, 2, 0, nil)
	b.Place("sign", c.X, c.Z+6, 0, nil)
}

// florestaCache is a loot nook off a side trail, hidden by bushes.
func florestaCache(b *Builder, cx int) {
	spot := Pt{54 + b.Rng.Intn(4), 54 + b.Rng.Intn(4)}
	b.FillCircle(spot.X, spot.Z, 2, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
	b.PaintTrail(b.WindingPath(Pt{cx + 6, 54}, spot, 0.6), 1, "dirt")
	b.Place("crate", spot.X, spot.Z, 0, nil)
	b.Place("barrel", spot.X+1, spot.Z, 0, nil)
	b.Place("rock", spot.X-1, spot.Z-1, 0, nil)
}

func florestaWoods(b *Builder) {
	all := Rect{0, 0, 64, 64}
	pine := map[string]any{"variant": "pine"}
	b.Scatter(all, Scatter{Prefab: "tree", Density: 0.6, Clump: 9, Variants: []string{"pine", "round"}, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "tree", Density: 0.16, Props: pine, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.14, Terrains: []string{"grass", "sand", "dirt"}})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.025, Clump: 6, Terrains: []string{"grass", "sand"}})
}
