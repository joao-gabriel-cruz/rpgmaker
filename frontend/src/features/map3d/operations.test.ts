import { describe, expect, it } from 'vitest'
import {
  MAX_HEIGHT,
  applyPatch,
  createGrid,
  deserializeGrid,
  getCell,
  serializeGrid,
} from './grid'
import { brushCells, erase, fill, floodRegion, lower, paintTerrain, raise, setHeight } from './operations'
import { TERRAIN_IDS, TERRAIN_TYPES } from './terrainTypes'

describe('operations', () => {
  it('paints terrain and ignores out-of-bounds coords', () => {
    const g = createGrid(4, 4)
    const { grid, patch } = paintTerrain(g, [{ x: 1, z: 1 }, { x: -1, z: 0 }, { x: 4, z: 0 }], 'lava')
    expect(getCell(grid, 1, 1)?.terrain).toBe('lava')
    expect(patch.cells).toEqual([{ x: 1, z: 1, terrain: 'lava', height: 1 }])
    expect(getCell(g, 1, 1)?.terrain).toBe('floor')
  })

})

describe('heights', () => {
  it('clamps heights to 0..10', () => {
    const g = createGrid(2, 2)
    expect(getCell(lower(g, [{ x: 0, z: 0 }], 5).grid, 0, 0)?.height).toBe(0)
    expect(getCell(raise(g, [{ x: 0, z: 0 }], 50).grid, 0, 0)?.height).toBe(MAX_HEIGHT)
    expect(getCell(setHeight(g, [{ x: 0, z: 0 }], 99).grid, 0, 0)?.height).toBe(MAX_HEIGHT)
    expect(getCell(setHeight(g, [{ x: 0, z: 0 }], -3).grid, 0, 0)?.height).toBe(0)
  })

  it('omits unchanged cells from the patch', () => {
    const g = createGrid(2, 2)
    expect(setHeight(g, [{ x: 0, z: 0 }], 1).patch.cells).toHaveLength(0)
    const top = setHeight(g, [{ x: 0, z: 0 }], MAX_HEIGHT).grid
    expect(raise(top, [{ x: 0, z: 0 }]).patch.cells).toHaveLength(0)
  })

  it('erase restores the default cell', () => {
    let g = createGrid(3, 3)
    g = paintTerrain(g, [{ x: 0, z: 0 }], 'snow').grid
    g = raise(g, [{ x: 0, z: 0 }], 4).grid
    expect(getCell(erase(g, [{ x: 0, z: 0 }]).grid, 0, 0)).toEqual({ terrain: 'floor', height: 1 })
  })

})

describe('flood fill, brush and patches', () => {
  it('flood fill covers contiguous same terrain and height only', () => {
    let g = createGrid(5, 1)
    g = raise(g, [{ x: 2, z: 0 }]).grid
    expect(floodRegion(g, { x: 0, z: 0 })).toHaveLength(2)
    const { grid, patch } = fill(g, { x: 0, z: 0 }, 'water')
    expect(patch.cells).toHaveLength(2)
    expect(getCell(grid, 3, 0)?.terrain).toBe('floor')
  })

  it('flood fill returns nothing for an out-of-bounds start', () => {
    expect(fill(createGrid(2, 2), { x: 9, z: 9 }, 'sand').patch.cells).toEqual([])
  })

  it('brush is clamped at borders', () => {
    const g = createGrid(5, 5)
    expect(brushCells(g, { x: 2, z: 2 }, 3)).toHaveLength(9)
    expect(brushCells(g, { x: 0, z: 0 }, 3)).toHaveLength(4)
    expect(brushCells(g, { x: 2, z: 2 }, 1)).toHaveLength(1)
  })

})

describe('patches and serialization', () => {
  it('applying the patch to an equal grid reproduces the result', () => {
    const base = createGrid(6, 6)
    const r = raise(paintTerrain(base, brushCells(base, { x: 3, z: 3 }, 5), 'grass').grid, brushCells(base, { x: 3, z: 3 }, 3), 2)
    const replay = applyPatch(applyPatch(createGrid(6, 6), paintTerrain(base, brushCells(base, { x: 3, z: 3 }, 5), 'grass').patch), r.patch)
    expect(replay).toEqual(r.grid)
  })

  it('serializes and deserializes losslessly', () => {
    let g = createGrid(7, 3)
    g = paintTerrain(g, [{ x: 6, z: 2 }], 'ruin').grid
    g = setHeight(g, [{ x: 0, z: 0 }], MAX_HEIGHT).grid
    const json = JSON.parse(JSON.stringify(serializeGrid(g)))
    expect(deserializeGrid(json)).toEqual(g)
    expect(() => deserializeGrid({ width: 2, depth: 2, data: 'abc' })).toThrow()
  })
})

describe('terrain catalog', () => {
  it('has an entry with a PT-BR name for every id', () => {
    for (const id of TERRAIN_IDS) expect(TERRAIN_TYPES[id].name).toBeTruthy()
  })
})
