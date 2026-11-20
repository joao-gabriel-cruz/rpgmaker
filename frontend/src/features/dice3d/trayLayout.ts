import { Vector3 } from 'three'
import type { Rng } from './prng'

export const SPACING = 2.1
export const TRAY_W = 16
export const TRAY_D = 10

// Grid slots with small jitter; spacing exceeds die diameter so dice never overlap.
export function slotFor(index: number, total: number, rng: Rng): Vector3 {
  const cols = Math.min(total, 6)
  const rows = Math.ceil(total / cols)
  const row = Math.floor(index / cols)
  const inRow = row === rows - 1 ? total - row * cols : cols
  const col = index % cols
  const x = (col - (inRow - 1) / 2) * SPACING + (rng() - 0.5) * 0.2
  const z = (row - (rows - 1) / 2) * SPACING + (rng() - 0.5) * 0.2
  return new Vector3(x, 0, z)
}

// Half extents (x, z) of the dice grid, including a margin for the dice themselves.
export function gridHalfExtents(total: number): { x: number; z: number } {
  const cols = Math.min(Math.max(total, 1), 6)
  const rows = Math.ceil(Math.max(total, 1) / cols)
  return { x: ((cols - 1) * SPACING) / 2 + 1.5, z: ((rows - 1) * SPACING) / 2 + 1.5 }
}
