package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "fazenda",
		Name:        "Sítio Isolado",
		Description: "Casa, celeiro e plantações em fileiras cercadas, com poço no terreiro e um pomar ao fundo. Uma estrada de terra liga tudo ao portão. Quem vive aqui? O que o celeiro esconde?",
		Tags:        []string{"natureza", "refúgio", "social"},
		Width:       80, Depth: 80,
		Ambience: Ambience{TimeOfDay: "dawn", Weather: "clear"},
		Generate: generateFazenda,
	})
}

var (
	fazendaField1 = Rect{8, 8, 28, 22}
	fazendaField2 = Rect{46, 8, 26, 20}
)

func generateFazenda(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(80, 80, seed, "grass", 1)
	b.Hill(8, 70, 9, 2, "")
	b.Hill(72, 68, 9, 2, "")
	fazendaRoads(b)
	fazendaFarmyard(b)
	fazendaField(b, fazendaField1, true)
	fazendaField(b, fazendaField2, false)
	fazendaOrchard(b)
	fazendaJunk(b)
	b.FenceRect(Rect{2, 2, 76, 70}, South)
	fazendaWilds(b)
	b.ClearEntry()
	return b.Build()
}

func fazendaRoads(b *Builder) {
	dirt := func(pts ...Pt) { b.Road(pts, 3, "dirt", 1) }
	dirt(Pt{40, 79}, Pt{40, 38})
	b.Road([]Pt{{40, 36}, {22, 36}, {22, 32}}, 2, "dirt", 1)
	b.Road([]Pt{{40, 36}, {58, 36}, {58, 30}}, 2, "dirt", 1)
	b.Blob(40, 44, 9, 0.12, Brush{Terrain: "dirt", Height: 1, Reserve: true})
}

func fazendaFarmyard(b *Builder) {
	house := b.Building(Rect{22, 40, 10, 8}, BuildingOpts{Door: East, Open: true, Floor: "floor"})
	b.FurnishCorners(house, "crate", 2)
	b.Place("barrel", 24, 44, 0, nil)
	barn := b.Building(Rect{50, 38, 14, 11}, BuildingOpts{Door: West, DoorAt: 5, Open: true, Floor: "dirt"})
	b.FurnishCorners(barn, "barrel", 2)
	b.PlaceSpaced("crate", Pt{59, 40}, 1, 0, 3, 0, nil)
	b.PlaceSpaced("crate", Pt{59, 46}, 1, 0, 3, 0, nil)
	b.Place("campfire", 56, 43, 0, nil)
	b.Place("well", 40, 44, 0, nil)
	b.Place("bench", 33, 42, 0, nil)
	b.Place("sign", 38, 52, 0, nil)
	b.Place("sign", 42, 71, 0, nil)
	b.Place("post", 35, 51, 0, nil)
	b.Place("lamp_post", 44, 50, 0, nil)
}

// fazendaField paints alternating dirt and grass rows (dirt rows hold crops), fences it and leaves a gate gap.
func fazendaField(b *Builder, r Rect, horizontal bool) {
	for z := r.Z; z < r.Z+r.D; z++ {
		for x := r.X; x < r.X+r.W; x++ {
			i := z - r.Z
			if !horizontal {
				i = x - r.X
			}
			t := "dirt"
			if i%3 == 2 {
				t = "grass"
			}
			b.Set(x, z, t, 1)
		}
	}
	b.Scatter(r, Scatter{Prefab: "bush", Density: 0.55, Terrains: []string{"dirt"}})
	b.Place("post", r.X+r.W/2, r.Z+r.D/2-b.Rng.Intn(2)*3, 0, nil)
	b.Place("sign", r.X+r.W/2+1, r.Z+r.D/2, 0, nil)
	b.FenceRect(r.Grow(1), South)
	b.Reserve(r.Grow(1))
}

func fazendaOrchard(b *Builder) {
	round := map[string]any{"variant": "round"}
	for z := 52; z <= 68; z += 4 {
		for x := 8; x <= 22; x += 4 {
			b.Place("tree", x+b.Rng.Intn(2), z+b.Rng.Intn(2), 0, round)
		}
	}
	b.Reserve(Rect{6, 50, 20, 20})
}

// fazendaJunk is the abandoned tractor and scrap pile in the south-east corner.
func fazendaJunk(b *Builder) {
	b.Place("car", 58+b.Rng.Intn(3), 58, 0, nil)
	b.Place("car", 64, 62+b.Rng.Intn(3), 90, nil)
	b.Place("barrel", 62, 58, 0, nil)
	b.Place("crate", 60, 60, 0, nil)
	b.Place("sandbags", 56, 62, 0, nil)
	b.Reserve(Rect{54, 54, 16, 16})
}

func fazendaWilds(b *Builder) {
	all := Rect{0, 0, 80, 80}
	b.Scatter(Rect{0, 0, 80, 80}, Scatter{Prefab: "tree", Density: 0.06, Clump: 8, Variants: []string{"pine", "round"}, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.03, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.008, Terrains: []string{"grass"}})
}
