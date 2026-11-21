import type { Coord } from './grid'

// Bresenham line, both ends included.
export function lineCells(a: Coord, b: Coord): Coord[] {
  const out: Coord[] = []
  const dx = Math.abs(b.x - a.x)
  const dz = Math.abs(b.z - a.z)
  const sx = a.x < b.x ? 1 : -1
  const sz = a.z < b.z ? 1 : -1
  let err = dx - dz
  let { x, z } = a
  for (;;) {
    out.push({ x, z })
    if (x === b.x && z === b.z) return out
    const e2 = 2 * err
    if (e2 > -dz) {
      err -= dz
      x += sx
    }
    if (e2 < dx) {
      err += dx
      z += sz
    }
  }
}
