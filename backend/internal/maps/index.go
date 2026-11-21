package maps

// terrainWallBonus is the extra wall height of vision-blocking terrain (stone, ruin).
const terrainWallBonus = 3

// index caches per-cell blocking data so queries avoid rescanning objects.
type index struct {
	grid      *Grid
	moveBlock []bool
	wallTop   []float64
	deck      []int // bridge deck height per cell, -1 where there is no bridge
}

func buildIndex(g *Grid, objects []Object) *index {
	ix := &index{grid: g, moveBlock: make([]bool, len(g.Cells)), wallTop: make([]float64, len(g.Cells)), deck: make([]int, len(g.Cells))}
	for i := range ix.deck {
		ix.deck[i] = -1
	}
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
	ix.addBridges(objects)
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

// addBridges sets the deck of every bridge cell. A chain of adjacent bridge cells shares one deck:
// the highest terrain touching it (or the prop deck of each cell, when set).
func (ix *index) addBridges(objects []Object) {
	g := ix.grid
	explicit := map[int]int{}
	for _, o := range objects {
		if o.Prefab != "bridge" {
			continue
		}
		d, has := propInt(o.Props, "deck")
		for _, p := range Footprint(o) {
			if g.InBounds(p.X, p.Z) {
				i := p.Z*g.Width + p.X
				explicit[i] = -1
				if has {
					explicit[i] = clampHeight(d)
				}
			}
		}
	}
	seen := map[int]bool{}
	for i := range explicit {
		if !seen[i] {
			ix.setChainDeck(i, explicit, seen)
		}
	}
}

func (ix *index) setChainDeck(start int, cells map[int]int, seen map[int]bool) {
	g := ix.grid
	chain, best := []int{start}, 0
	seen[start] = true
	for k := 0; k < len(chain); k++ {
		i := chain[k]
		best = max(best, g.Cells[i].Height)
		x, z := i%g.Width, i/g.Width
		for _, n := range []Coord{{x + 1, z}, {x - 1, z}, {x, z + 1}, {x, z - 1}} {
			if !g.InBounds(n.X, n.Z) {
				continue
			}
			ni := n.Z*g.Width + n.X
			if _, bridge := cells[ni]; !bridge {
				best = max(best, g.Cells[ni].Height)
			} else if !seen[ni] {
				seen[ni] = true
				chain = append(chain, ni)
			}
		}
	}
	for _, i := range chain {
		ix.deck[i] = best
		if d := cells[i]; d >= 0 {
			ix.deck[i] = d
		}
		ix.moveBlock[i] = false // the deck spans blocking terrain
	}
}

func propInt(props map[string]any, key string) (int, bool) {
	switch v := props[key].(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	}
	return 0, false
}
