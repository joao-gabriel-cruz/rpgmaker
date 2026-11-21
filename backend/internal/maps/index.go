package maps

// terrainWallBonus is the extra wall height of vision-blocking terrain (stone, ruin).
const terrainWallBonus = 3

// index caches per-cell blocking data so queries avoid rescanning objects.
type index struct {
	grid      *Grid
	moveBlock []bool
	wallTop   []float64
}

func buildIndex(g *Grid, objects []Object) *index {
	ix := &index{grid: g, moveBlock: make([]bool, len(g.Cells)), wallTop: make([]float64, len(g.Cells))}
	ix.addTerrain()
	ix.addObjects(objects)
	return ix
}

func (ix *index) addTerrain() {
	for i, c := range ix.grid.Cells {
		t, _ := TerrainByID(c.Terrain)
		ix.moveBlock[i] = t.BlocksMovement
		top := float64(c.Height)
		if t.BlocksVision {
			top += terrainWallBonus
		}
		ix.wallTop[i] = top
	}
}

func (ix *index) addObjects(objects []Object) {
	g := ix.grid
	for _, o := range objects {
		r := RuleFor(o)
		if !r.BlocksMovement && !r.BlocksVision {
			continue
		}
		for _, p := range Footprint(o) {
			if !g.InBounds(p.X, p.Z) {
				continue
			}
			i := p.Z*g.Width + p.X
			if r.BlocksMovement {
				ix.moveBlock[i] = true
			}
			if r.BlocksVision {
				ix.wallTop[i] = max(ix.wallTop[i], float64(g.Cells[i].Height+r.Height))
			}
		}
	}
}
