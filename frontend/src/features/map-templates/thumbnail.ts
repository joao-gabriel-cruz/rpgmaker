import { cellIndex, type MapGrid } from '../map3d/grid'
import { PREFAB_META_BY_ID } from '../map3d/objects/catalogData'
import { TERRAIN_TYPES } from '../map3d/terrainTypes'

export interface ThumbObject {
  prefab: string
  x: number
  z: number
  rotation: number
}

/** The slice of CanvasRenderingContext2D the thumbnail needs (easy to fake in tests). */
export interface Painter {
  fillStyle: string | CanvasGradient | CanvasPattern
  fillRect(x: number, y: number, w: number, h: number): void
}

const OBJECT_COLORS: Record<string, string> = {
  wall: '#3f3f46', door: '#a16207', barricade: '#a8a29e', fence: '#a8a29e', watchtower: '#78350f', post: '#d4d4d8',
  tree: '#14532d', bush: '#3f7d3a', rock: '#9ca3af', car: '#dc2626', crate: '#b45309', campfire: '#fb923c', volcano: '#7f1d1d',
}
const FALLBACK_OBJECT = '#e4e4e7'

/** Terrain color darkened for low cells and brightened a little for tall ones. */
export function shade(hex: string, height: number): string {
  const k = Math.min(1.25, 0.6 + height * 0.12)
  const channel = (i: number) => Math.min(255, Math.round(parseInt(hex.slice(1 + i * 2, 3 + i * 2), 16) * k))
  return `rgb(${channel(0)},${channel(1)},${channel(2)})`
}

function footprintOf(o: ThumbObject): { w: number; d: number } {
  const meta = PREFAB_META_BY_ID[o.prefab]
  const w = meta?.width ?? 1
  const d = meta?.depth ?? 1
  return o.rotation === 90 || o.rotation === 270 ? { w: d, d: w } : { w, d }
}

/** Top-down thumbnail: one `scale`-pixel square per cell (terrain shaded by height), objects painted over it. */
export function drawThumbnail(ctx: Painter, grid: MapGrid, objects: readonly ThumbObject[], scale: number) {
  for (let z = 0; z < grid.depth; z++) {
    for (let x = 0; x < grid.width; x++) {
      const c = grid.cells[cellIndex(grid, x, z)]
      ctx.fillStyle = shade(TERRAIN_TYPES[c.terrain].color, c.height)
      ctx.fillRect(x * scale, z * scale, scale, scale)
    }
  }
  for (const o of objects) {
    const { w, d } = footprintOf(o)
    ctx.fillStyle = OBJECT_COLORS[o.prefab] ?? FALLBACK_OBJECT
    ctx.fillRect(o.x * scale, o.z * scale, w * scale, d * scale)
  }
}
