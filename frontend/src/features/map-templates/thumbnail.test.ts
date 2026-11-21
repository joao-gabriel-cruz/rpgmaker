import { expect, it } from 'vitest'
import { createGrid } from '../map3d/grid'
import { drawThumbnail, shade, type Painter } from './thumbnail'

function recorder() {
  const calls: { color: string; rect: number[] }[] = []
  const p: Painter = { fillStyle: '', fillRect(x, y, w, h) { calls.push({ color: String(this.fillStyle), rect: [x, y, w, h] }) } }
  return { p, calls }
}

it('paints one square per cell, colored by terrain and darker when lower', () => {
  const grid = createGrid(2, 1, { terrain: 'grass', height: 1 })
  const cells = [{ terrain: 'water' as const, height: 0 }, { terrain: 'grass' as const, height: 5 }]
  const { p, calls } = recorder()
  drawThumbnail(p, { ...grid, cells }, [], 3)
  expect(calls.map((c) => c.rect)).toEqual([[0, 0, 3, 3], [3, 0, 3, 3]])
  expect(calls[0].color).toBe(shade('#3b82c4', 0))
  expect(calls[1].color).toBe(shade('#5a9e4b', 5))
  const low = shade('#5a9e4b', 0).match(/\d+/g)!.map(Number)
  const high = shade('#5a9e4b', 5).match(/\d+/g)!.map(Number)
  expect(low[1]).toBeLessThan(high[1])
})

it('paints objects over the terrain using their rotated footprint', () => {
  const { p, calls } = recorder()
  drawThumbnail(p, createGrid(4, 4), [
    { prefab: 'car', x: 1, z: 1, rotation: 0 },
    { prefab: 'car', x: 2, z: 2, rotation: 90 },
    { prefab: 'tree', x: 0, z: 3, rotation: 0 },
  ], 2)
  const objects = calls.slice(16)
  expect(objects.map((c) => c.rect)).toEqual([[2, 2, 4, 2], [4, 4, 2, 4], [0, 6, 2, 2]])
  expect(new Set(objects.map((c) => c.color)).size).toBe(2)
})
