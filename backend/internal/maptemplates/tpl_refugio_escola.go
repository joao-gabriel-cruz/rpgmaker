package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "refugio-escola",
		Name:        "Refúgio Fortificado",
		Description: "Uma escola murada, com torres de vigia nos cantos, sacos de areia no portão e uma fogueira no pátio onde os sobreviventes se aquecem. Negociar a entrada, defender o muro ou invadir à noite?",
		Tags:        []string{"refúgio", "social", "combate"},
		Width:       64, Depth: 64,
		Ambience: Ambience{TimeOfDay: "night", Weather: "clear"},
		Generate: generateRefugio,
	})
}

var (
	refugioYard   = Rect{6, 6, 52, 48}
	refugioSchool = Rect{16, 12, 32, 18}
	refugioGate   = []Pt{{31, 53}, {32, 53}}
)

func generateRefugio(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(64, 64, seed, "grass", 1)
	b.FillRect(refugioYard, Brush{Terrain: "dirt", Height: 1})
	b.Line(Pt{32, 62}, Pt{32, 30}, 3, Brush{Terrain: "dirt", Height: 1, Reserve: true})
	refugioWalls(b)
	refugioSchoolBuilding(b)
	refugioGateDefences(b)
	refugioCourtyard(b)
	refugioCamp(b)
	refugioGarden(b)
	b.Reserve(refugioYard)
	refugioOutside(b)
	b.ClearEntry()
	return b.Build()
}

// refugioWalls closes the compound: corner watchtowers, a wall ring and one gate of open doors.
func refugioWalls(b *Builder) {
	r := refugioYard
	for _, c := range []Pt{{r.X, r.Z}, {r.X + r.W - 1, r.Z}, {r.X, r.Z + r.D - 1}, {r.X + r.W - 1, r.Z + r.D - 1}, {r.X + r.W/2, r.Z}} {
		b.Place("watchtower", c.X, c.Z, 0, nil)
	}
	b.WallRing(r, "wall", refugioGate...)
	for _, g := range refugioGate {
		b.Place("door", g.X, g.Z, 0, map[string]any{"open": true})
	}
}

// refugioSchoolBuilding is the school: a corridor/hall to the south and four classrooms to the north.
func refugioSchoolBuilding(b *Builder) {
	r := refugioSchool
	info := b.Building(r, BuildingOpts{Door: South, DoorAt: 16, Open: true, Floor: "floor"})
	wallZ := r.Z + 6
	for x := r.X + 1; x < r.X+r.W-1; x++ {
		if (x-r.X)%7 == 3 {
			b.Place("door", x, wallZ, 0, map[string]any{"open": true})
		} else {
			b.Place("wall", x, wallZ, 0, nil)
		}
	}
	for _, off := range []int{7, 14, 21} {
		b.PlaceSpaced("wall", Pt{r.X + off, r.Z + 1}, 0, 1, 5, 0, nil)
	}
	for _, off := range []int{2, 9, 16, 23} {
		b.Place("crate", r.X+off, r.Z+1, 0, nil)
		b.Place("barrel", r.X+off+1, r.Z+1, 0, nil)
	}
	b.PlaceSpaced("bench", Pt{r.X + 3, r.Z + 9}, 8, 0, 3, 0, nil)
	b.PlaceSpaced("bench", Pt{r.X + 3, r.Z + 13}, 8, 0, 3, 0, nil)
	b.Place("campfire", r.X+r.W/2, r.Z+11, 0, nil)
	b.Place("lamp_post", info.Front.X-2, info.Front.Z, 0, nil)
	b.Place("lamp_post", info.Front.X+2, info.Front.Z, 0, nil)
}

func refugioGateDefences(b *Builder) {
	b.PlaceSpaced("sandbags", Pt{27, 56}, 1, 0, 3, 0, nil)
	b.PlaceSpaced("sandbags", Pt{34, 56}, 1, 0, 3, 0, nil)
	b.Place("barricade", 28, 58, 0, nil)
	b.Place("barricade", 35, 58, 0, nil)
	b.PlaceSpaced("sandbags", Pt{28, 51}, 0, -1, 2, 0, nil)
	b.PlaceSpaced("sandbags", Pt{35, 51}, 0, -1, 2, 0, nil)
	b.Place("lamp_post", 29, 54, 0, nil)
	b.Place("lamp_post", 34, 54, 0, nil)
	b.Place("sign", 30, 59, 0, nil)
}

// refugioCourtyard has the shared bonfire with benches and light posts.
func refugioCourtyard(b *Builder) {
	b.FillCircle(32, 42, 4, Brush{Terrain: "stone", Height: 1, Reserve: true})
	b.Place("campfire", 32, 42, 0, nil)
	b.Place("bench", 28, 41, 0, nil)
	b.Place("bench", 35, 43, 0, nil)
	b.Place("bench", 31, 38, 0, nil)
	b.Place("bench", 32, 46, 0, nil)
	for _, p := range []Pt{{26, 38}, {38, 38}, {26, 47}, {38, 47}, {20, 34}, {44, 34}} {
		b.Place("lamp_post", p.X, p.Z, 0, nil)
	}
}

// refugioCamp is the survivors' tent village on the west side, with supplies and a lookout nook.
func refugioCamp(b *Builder) {
	for i, p := range []Pt{{9, 36}, {13, 36}, {9, 41}, {13, 41}, {9, 46}, {13, 46}} {
		b.Place("tent", p.X, p.Z, 0, nil)
		b.Place("barrel", p.X+2+i%2, p.Z+2, 0, nil)
	}
	b.PlaceSpaced("crate", Pt{18, 38}, 0, 2, 4, 0, nil)
	b.Place("campfire", 16, 49, 0, nil)
	b.Place("lamp_post", 16, 43, 0, nil)
	b.Place("lamp_post", 8, 33, 0, nil)
}

// refugioGarden is the vegetable garden: rows of dirt with bushes between grass paths.
func refugioGarden(b *Builder) {
	g := Rect{46, 34, 10, 16}
	for z := g.Z; z < g.Z+g.D; z++ {
		for x := g.X; x < g.X+g.W; x++ {
			t := "dirt"
			if (z-g.Z)%3 == 2 {
				t = "grass"
			}
			b.Set(x, z, t, 1)
		}
	}
	b.Scatter(g, Scatter{Prefab: "bush", Density: 0.6, Terrains: []string{"dirt"}})
	b.FenceRect(g.Grow(1), West)
	b.Place("barrel", 44, 50, 0, nil)
	b.Place("well", 44, 36, 0, nil)
}

func refugioOutside(b *Builder) {
	all := Rect{0, 0, 64, 64}
	b.Scatter(all, Scatter{Prefab: "tree", Density: 0.2, Clump: 7, Variants: []string{"pine", "round"}, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "bush", Density: 0.05, Terrains: []string{"grass"}})
	b.Scatter(all, Scatter{Prefab: "rock", Density: 0.02, Terrains: []string{"grass"}})
	b.Scatter(Rect{0, 0, 64, 6}, Scatter{Prefab: "car", Density: 0.03, Terrains: []string{"grass"}})
}
