package maptemplates

import "github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"

func init() {
	Register(Template{
		ID:          "rodovia-engarrafada",
		Name:        "Rodovia Engarrafada",
		Description: "Uma rodovia travada de carros abandonados, com posto de gasolina, barreiras e mato alto dos dois lados. Cobertura de sobra para emboscadas e fugas.",
		Tags:        []string{"combate", "perigo", "natureza"},
		Width:       96, Depth: 64,
		Ambience: Ambience{TimeOfDay: "day", Weather: "fog"},
		Generate: generateRodovia,
	})
}

// Highway lanes occupy z 25..34; shoulders (25, 34) stay car-free so there is always a way through.
const (
	hwTop    = 25
	hwBottom = 34
)

var hwLaneRows = []int{26, 27, 28, 31, 32, 33}

func generateRodovia(seed int64) (maps.Grid, []maps.Object) {
	b := NewBuilder(96, 64, seed, "grass", 1)
	rodoviaLandscape(b)
	hx := rodoviaHighway(b)
	rodoviaGasStation(b, hx)
	rodoviaTrails(b, hx)
	rodoviaJam(b)
	rodoviaWoods(b)
	b.ClearEntry()
	return b.Build()
}

func rodoviaLandscape(b *Builder) {
	b.Hill(14+b.Rng.Intn(14), 9+b.Rng.Intn(4), 10, 4, "grass")
	b.Hill(66+b.Rng.Intn(10), 50+b.Rng.Intn(5), 8, 2, "")
	b.Hill(8+b.Rng.Intn(10), 50+b.Rng.Intn(6), 7, 2, "")
}

// rodoviaHighway paints the asphalt and median and returns the x of the gas station.
func rodoviaHighway(b *Builder) int {
	road := Rect{0, hwTop, 96, hwBottom - hwTop + 1}
	b.FillRect(road.Grow(1), Brush{Terrain: "dirt", Height: 2, Reserve: true})
	b.FillRect(road, Brush{Terrain: "floor", Height: 2, Reserve: true})
	b.FillRect(Rect{0, 29, 96, 2}, Brush{Terrain: "dirt", Height: 2, Reserve: true})
	for x := 2; x < 94; x += 3 {
		b.Place("barricade", x, 29, 0, nil)
	}
	for _, x := range []int{20, 76} {
		rodoviaCheckpoint(b, x+b.Rng.Intn(5)-2)
	}
	return 56 + b.Rng.Intn(10)
}

// rodoviaCheckpoint blocks the lanes with barricades and crates, leaving a gap on the shoulder side.
func rodoviaCheckpoint(b *Builder, x int) {
	for z := hwTop + 1; z <= hwBottom-1; z++ {
		if z == 29 || z == 30 || z == hwTop+1+b.Rng.Intn(2) {
			continue
		}
		b.Place("barricade", x, z, 0, nil)
	}
	b.Place("crate", x+1, 27, 0, nil)
	b.Place("crate", x-1, 32, 0, nil)
}

func rodoviaGasStation(b *Builder, x int) {
	b.FillRect(Rect{x - 3, hwBottom + 1, 20, 7}, Brush{Terrain: "floor", Height: 1, Reserve: true})
	info := b.Building(Rect{x, hwBottom + 8, 10, 7}, BuildingOpts{Door: North, DoorAt: 4, Ruin: 0.12})
	b.FurnishCorners(info, "crate", 2)
	for _, px := range []int{x - 1, x + 3, x + 7, x + 11} {
		b.Place("post", px, hwBottom+4, 0, nil)
	}
	b.Place("car", x+13, hwBottom+3, 0, nil)
	b.Place("campfire", x+14, hwBottom+9, 0, nil)
	b.Place("barricade", x+12, hwBottom+9, 0, nil)
}

func rodoviaTrails(b *Builder, gx int) {
	dirt := Brush{Terrain: "dirt", Height: Keep, Reserve: true}
	b.Line(Pt{48, 63}, Pt{48, hwBottom + 1}, 3, dirt)
	b.Line(Pt{48, hwBottom + 6}, Pt{gx - 3, hwBottom + 6}, 2, dirt)
	// hideout in the north woods: trail from the highway to a clearing with a campfire
	cx := 66 + b.Rng.Intn(14)
	b.Line(Pt{cx, hwTop - 1}, Pt{cx, 11}, 2, dirt)
	b.Blob(cx, 10, 5, 0.2, Brush{Terrain: "dirt", Height: Keep, Reserve: true})
	b.Place("campfire", cx, 10, 0, nil)
	b.Place("crate", cx+2, 9, 0, nil)
	b.Place("crate", cx-3, 12, 0, nil)
	b.Place("bush", cx+3, 12, 0, nil)
	b.ClearPath(Pt{48, 63}, Pt{48, hwBottom + 1}, 3)
}

func rodoviaJam(b *Builder) {
	for _, z := range hwLaneRows {
		for x := 1; x < 94; {
			if b.Rng.Float64() < 0.55 {
				b.Place("car", x, z, 0, nil)
			}
			x += 2 + b.Rng.Intn(3)
		}
	}
}

func rodoviaWoods(b *Builder) {
	rodoviaShack(b)
	north, south := Rect{0, 0, 96, hwTop - 1}, Rect{0, hwBottom + 2, 96, 64 - hwBottom - 2}
	for _, area := range []Rect{north, south} {
		b.Scatter(area, Scatter{Prefab: "tree", Density: 0.2, Clump: 7, Variants: []string{"pine", "round"}, Terrains: []string{"grass"}})
		b.Scatter(area, Scatter{Prefab: "bush", Density: 0.09, Terrains: []string{"grass"}})
		b.Scatter(area, Scatter{Prefab: "rock", Density: 0.015, Terrains: []string{"grass"}})
	}
}

// rodoviaShack is a roadside vendor's shack, half collapsed.
func rodoviaShack(b *Builder) {
	info := b.Building(Rect{10 + b.Rng.Intn(8), 44, 7, 5}, BuildingOpts{Door: North, Ruin: 0.35})
	b.FurnishCorners(info, "crate", 1)
}
