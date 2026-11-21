import { describe, expect, it } from 'vitest'
import { layoutBlocks, layoutSlabs } from './blockLayout'
import { createGrid, SLAB_THICKNESS, surfaceHeight } from './grid'

describe('height-0 floor', () => {
  const grid = createGrid(2, 1, { terrain: 'water', height: 0 })

  it('generates a thin slab with the cell terrain and no blocks', () => {
    expect(layoutBlocks(grid).water).toEqual([])
    expect(layoutSlabs(grid).water).toEqual([0.5, SLAB_THICKNESS / 2, 0.5, 1.5, SLAB_THICKNESS / 2, 0.5])
  })

  it('does not add slabs for raised cells', () => {
    expect(layoutSlabs(createGrid(2, 2)).floor).toEqual([])
  })

  it('supports surface at the slab top for height 0', () => {
    expect(surfaceHeight(0)).toBe(SLAB_THICKNESS)
    expect(surfaceHeight(3)).toBe(3)
  })
})
