import { TERRAIN_IDS, type TerrainTypeId } from './terrainTypes'

export const MIN_HEIGHT = 0
export const MAX_HEIGHT = 10
export const DEFAULT_CELL: Cell = { terrain: 'floor', height: 1 }

export interface Cell {
  terrain: TerrainTypeId
  height: number
}

export interface Coord {
  x: number
  z: number
}

export interface CellPatch extends Coord, Cell {}

export interface MapPatch {
  cells: CellPatch[]
}

export interface MapGrid {
  width: number
  depth: number
  cells: readonly Cell[]
}

export interface SerializedGrid {
  width: number
  depth: number
  data: string
}

// Height-0 cells render as a thin slab of their terrain instead of a hole.
export const SLAB_THICKNESS = 0.15

// Top of the walkable surface: standing height for tokens, objects, lines and fog.
export const surfaceHeight = (height: number) => Math.max(height, SLAB_THICKNESS)

export const clampHeight = (h: number) =>
  Math.min(MAX_HEIGHT, Math.max(MIN_HEIGHT, Math.round(h)))

export const inBounds = (g: MapGrid, x: number, z: number) =>
  Number.isInteger(x) && Number.isInteger(z) && x >= 0 && z >= 0 && x < g.width && z < g.depth

export const cellIndex = (g: MapGrid, x: number, z: number) => z * g.width + x

export function createGrid(width: number, depth: number, fill: Cell = DEFAULT_CELL): MapGrid {
  if (width < 1 || depth < 1) throw new Error('grid size must be positive')
  return { width, depth, cells: Array.from({ length: width * depth }, () => ({ ...fill })) }
}

export function getCell(g: MapGrid, x: number, z: number): Cell | undefined {
  return inBounds(g, x, z) ? g.cells[cellIndex(g, x, z)] : undefined
}

export function applyPatch(g: MapGrid, patch: MapPatch): MapGrid {
  const cells = g.cells.slice()
  for (const p of patch.cells) {
    if (!inBounds(g, p.x, p.z)) continue
    cells[cellIndex(g, p.x, p.z)] = { terrain: p.terrain, height: clampHeight(p.height) }
  }
  return { ...g, cells }
}

export function serializeGrid(g: MapGrid): SerializedGrid {
  const data = g.cells
    .map((c) => TERRAIN_IDS.indexOf(c.terrain).toString(36) + c.height.toString(36))
    .join('')
  return { width: g.width, depth: g.depth, data }
}

export function deserializeGrid(s: SerializedGrid): MapGrid {
  if (s.data.length !== s.width * s.depth * 2) throw new Error('invalid serialized grid')
  const cells: Cell[] = []
  for (let i = 0; i < s.data.length; i += 2) {
    const terrain = TERRAIN_IDS[parseInt(s.data[i], 36)]
    const height = parseInt(s.data[i + 1], 36)
    if (!terrain || Number.isNaN(height)) throw new Error('invalid serialized grid')
    cells.push({ terrain, height: clampHeight(height) })
  }
  return { width: s.width, depth: s.depth, cells }
}
