import {
  DEFAULT_CELL,
  applyPatch,
  clampHeight,
  getCell,
  inBounds,
  type Cell,
  type Coord,
  type MapGrid,
  type MapPatch,
} from './grid'
import type { TerrainTypeId } from './terrainTypes'

export interface OpResult {
  grid: MapGrid
  patch: MapPatch
}

type CellFn = (cell: Cell) => Cell

// Applies fn to each in-bounds, unique coord and keeps only real changes in the patch.
function mapCells(grid: MapGrid, coords: Coord[], fn: CellFn): OpResult {
  const seen = new Set<number>()
  const patch: MapPatch = { cells: [] }
  for (const { x, z } of coords) {
    const cell = getCell(grid, x, z)
    const key = z * grid.width + x
    if (!cell || seen.has(key)) continue
    seen.add(key)
    const next = fn(cell)
    const height = clampHeight(next.height)
    if (next.terrain !== cell.terrain || height !== cell.height) {
      patch.cells.push({ x, z, terrain: next.terrain, height })
    }
  }
  return { grid: applyPatch(grid, patch), patch }
}

export const paintTerrain = (g: MapGrid, coords: Coord[], type: TerrainTypeId) =>
  mapCells(g, coords, (c) => ({ ...c, terrain: type }))

export const raise = (g: MapGrid, coords: Coord[], delta = 1) =>
  mapCells(g, coords, (c) => ({ ...c, height: c.height + delta }))

export const lower = (g: MapGrid, coords: Coord[], delta = 1) => raise(g, coords, -delta)

export const setHeight = (g: MapGrid, coords: Coord[], height: number) =>
  mapCells(g, coords, (c) => ({ ...c, height }))

export const erase = (g: MapGrid, coords: Coord[]) => mapCells(g, coords, () => ({ ...DEFAULT_CELL }))

const NEIGHBORS = [
  [1, 0],
  [-1, 0],
  [0, 1],
  [0, -1],
]

// Flood fill over 4-connected cells sharing the start cell's terrain and height.
export function floodRegion(grid: MapGrid, start: Coord): Coord[] {
  const origin = getCell(grid, start.x, start.z)
  if (!origin) return []
  const region: Coord[] = []
  const visited = new Set([start.z * grid.width + start.x])
  const stack = [start]
  while (stack.length) {
    const cur = stack.pop()!
    region.push(cur)
    for (const [dx, dz] of NEIGHBORS) {
      const x = cur.x + dx
      const z = cur.z + dz
      const key = z * grid.width + x
      if (!inBounds(grid, x, z) || visited.has(key)) continue
      const c = getCell(grid, x, z)!
      if (c.terrain !== origin.terrain || c.height !== origin.height) continue
      visited.add(key)
      stack.push({ x, z })
    }
  }
  return region
}

export const fill = (g: MapGrid, start: Coord, type: TerrainTypeId) =>
  paintTerrain(g, floodRegion(g, start), type)

export function brushCells(grid: MapGrid, center: Coord, size: number): Coord[] {
  const r = Math.floor(size / 2)
  const out: Coord[] = []
  for (let dz = -r; dz <= r; dz++) {
    for (let dx = -r; dx <= r; dx++) {
      const x = center.x + dx
      const z = center.z + dz
      if (inBounds(grid, x, z)) out.push({ x, z })
    }
  }
  return out
}
