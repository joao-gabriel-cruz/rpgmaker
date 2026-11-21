package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "vila-abandonada",
		Name:        "Vila Abandonada",
		Description: "Casas de muros caídos em volta de uma praça de terra, com rio a oeste e um morro com torre de vigia. Bom lugar para saquear, se esconder ou ser emboscado.",
		Tags:        []string{"urbano", "perigo", "social"},
		Width:       64, Depth: 64,
		Ambience: Ambience{TimeOfDay: "dusk", Weather: "clear"},
		Generate: generateVila,
	})
}

const vilaRoadZ = 32 // east-west main street

func generateVila(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(64, 64, seed, "grass", 1)
	vilaLandscape(b)
	vilaStreets(b)
	vilaHouses(b)
	vilaClutter(b)
	vilaVegetation(b)
	b.ClearEntry()
	return b.Build()
}

func vilaLandscape(b *Builder) {
	hx, hz := 55+b.Rng.Intn(4)-2, 9+b.Rng.Intn(3)
	b.Hill(hx, hz, 11, 4, "grass")
	b.Hill(20+b.Rng.Intn(6), 54+b.Rng.Intn(5), 7, 2, "")
	b.Hill(48+b.Rng.Intn(6), 54+b.Rng.Intn(4), 8, 2, "")
	b.River(Pt{9 + b.Rng.Intn(3), 0}, Pt{9 + b.Rng.Intn(3), 63}, 3, 2, "sand")
	top := b.Cell(hx, hz).Height
	b.FillCircle(hx, hz, 2, Brush{Terrain: "dirt", Height: top})
	b.Reserve(Rect{hx - 2, hz - 2, 5, 5})
	b.Place("watchtower", hx, hz, 0, nil)
	b.Scatter(Rect{hx - 11, hz - 9, 22, 20}, Scatter{Prefab: "rock", Density: 0.07, MaxHeight: 8})
	b.Place("barricade", hx-4, hz+5, 0, nil)
}

func vilaStreets(b *Builder) {
	b.Road([]Pt{{32, 63}, {32, 2}}, 3, "dirt", 1)
	b.Road([]Pt{{4, vilaRoadZ}, {63, vilaRoadZ}}, 3, "dirt", 1)
	plaza := Brush{Terrain: "dirt", Height: 1, Reserve: true}
	b.Blob(32, vilaRoadZ, 6.5, 0.12, plaza)
}

func vilaHouses(b *Builder) {
	// first ring: along the main street, doors toward it
	for _, x := range [][2]int{{17, 29}, {36, 47}, {49, 62}} {
		vilaRow(b, x[0], x[1], 27, North)
		vilaRow(b, x[0], x[1], 37, South)
	}
	// second ring: beside the north-south road, doors toward it
	for _, z := range []int{8, 47} {
		vilaSide(b, 18, 29, z, East)
		vilaSide(b, 36, 47, z, West)
	}
}

// vilaRow packs houses between x0 and x1; anchor is the z of the street-facing wall.
func vilaRow(b *Builder, x0, x1, anchor int, face Side) {
	for x := x0; x < x1-4; {
		w, d := 5+b.Rng.Intn(3), 5+b.Rng.Intn(2)
		if x+w > x1 {
			break
		}
		z, door := anchor-d+1, South
		if face == South {
			z, door = anchor, North
		}
		vilaHouse(b, Rect{x, z, w, d}, door)
		x += w + 1 + b.Rng.Intn(2)
	}
}

func vilaSide(b *Builder, x0, x1, z int, door Side) {
	w, d := 6+b.Rng.Intn(3), 5+b.Rng.Intn(3)
	x := x0 + b.Rng.Intn(max(1, x1-x0-w+1))
	if door == West {
		x = x1 - w + 1 - b.Rng.Intn(max(1, x1-x0-w+1))
	}
	vilaHouse(b, Rect{x, z + b.Rng.Intn(3), w, d}, door)
}

func vilaHouse(b *Builder, r Rect, door Side) {
	info := b.Building(r, BuildingOpts{Door: door, Ruin: b.Rng.Float64() * 0.3, Open: b.Rng.Intn(2) == 0, Floor: "floor"})
	b.FurnishCorners(info, "crate", 1+b.Rng.Intn(2))
	// a dirt path links the door to the nearest street
	target := Pt{32, info.Front.Z}
	if door == North || door == South {
		target = Pt{info.Front.X, vilaRoadZ}
	}
	b.Line(info.Front, target, 1, Brush{Terrain: "dirt", Height: 1, Reserve: true})
}

func vilaClutter(b *Builder) {
	b.Place("campfire", 32, vilaRoadZ, 0, nil)
	for _, p := range []Pt{{29, 29}, {35, 35}} {
		b.Place("crate", p.X, p.Z, 0, nil)
	}
	b.Place("barricade", 62, vilaRoadZ-1, 0, nil)
	b.Place("barricade", 62, vilaRoadZ+1, 0, nil)
	b.Place("barricade", 31, 3, 0, nil)
	b.Place("barricade", 33, 3, 0, nil)
	for i := 0; i < 9; i++ {
		vilaCar(b)
	}
}

func vilaCar(b *Builder) {
	if b.Rng.Intn(2) == 0 {
		b.Place("car", 8+b.Rng.Intn(50), vilaRoadZ-1+b.Rng.Intn(3), 0, nil)
		return
	}
	z := 8 + b.Rng.Intn(52)
	if z > 50 {
		z = 40 + b.Rng.Intn(10)
	}
	b.Place("car", 31+b.Rng.Intn(3), z, 90, nil)
}

func vilaVegetation(b *Builder) {
	all := Rect{0, 0, 64, 64}
	b.Scatter(all, Scatter{Prefab: "tree", Density: 0.14, Clump: 8, Variants: []string{"pine", "round"}, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.05, Terrains: []string{"grass", "sand"}})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.01, Terrains: []string{"grass", "sand"}})
}
