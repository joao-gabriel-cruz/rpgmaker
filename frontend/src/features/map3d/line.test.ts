import { describe, expect, it } from 'vitest'
import { lineCells } from './line'

describe('lineCells', () => {
  it('returns one cell for the same point', () => {
    expect(lineCells({ x: 2, z: 2 }, { x: 2, z: 2 })).toEqual([{ x: 2, z: 2 }])
  })

  it('has no gaps on a fast diagonal-ish drag', () => {
    const cells = lineCells({ x: 0, z: 0 }, { x: 9, z: 3 })
    expect(cells).toHaveLength(10)
    cells.slice(1).forEach((c, i) => {
      expect(Math.abs(c.x - cells[i].x)).toBeLessThanOrEqual(1)
      expect(Math.abs(c.z - cells[i].z)).toBeLessThanOrEqual(1)
    })
    expect(cells.at(-1)).toEqual({ x: 9, z: 3 })
  })

  it('works backwards', () => {
    expect(lineCells({ x: 3, z: 0 }, { x: 0, z: 0 }).map((c) => c.x)).toEqual([3, 2, 1, 0])
  })
})
