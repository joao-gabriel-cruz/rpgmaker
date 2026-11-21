import { api } from '../../lib/api'
import { getCell, type Coord, type MapGrid } from '../map3d/grid'

export interface Point {
  x: number
  z: number
}

export interface Drawing {
  id: string
  sceneId: string
  userId: string
  color: string
  width: number
  points: Point[]
}

export type DrawMode = 'none' | 'ruler' | 'draw'

export interface Measure {
  cells: number
  dh: number
}

// Server limit (drawings.MaxPoints); the client simplifies before sending.
export const MAX_POINTS = 500
const MIN_STEP = 0.2
const LIFT = 0.06
const SAMPLE = 0.25

export const DRAW_COLORS = ['#f87171', '#fbbf24', '#34d399', '#60a5fa', '#a78bfa', '#f5f5f5']
export const DRAW_WIDTHS = [2, 4, 6] as const

export const drawingsUrl = (campaignId: string, sceneId: string) => `/campaigns/${campaignId}/scenes/${sceneId}/drawings`

export async function loadDrawings(campaignId: string, sceneId: string): Promise<Drawing[]> {
  const data = await api.get<Drawing[]>(drawingsUrl(campaignId, sceneId))
  return Array.isArray(data) ? data : []
}

export const applyAdded = (list: Drawing[], sceneId: string, d: Drawing): Drawing[] =>
  d.sceneId !== sceneId || list.some((o) => o.id === d.id) ? list : [...list, d]

export const applyRemoved = (list: Drawing[], id: string): Drawing[] => list.filter((d) => d.id !== id)

const heightAt = (grid: MapGrid, x: number, z: number) => getCell(grid, Math.floor(x), Math.floor(z))?.height ?? 1

/**
 * Ruler rule: grid distance with diagonals (Chebyshev, a diagonal step costs 1 cell),
 * plus the height difference between the two cells.
 */
export function measure(grid: MapGrid, a: Coord, b: Coord): Measure {
  return {
    cells: Math.max(Math.abs(b.x - a.x), Math.abs(b.z - a.z)),
    dh: heightAt(grid, b.x, b.z) - heightAt(grid, a.x, a.z),
  }
}

export function formatMeasure({ cells, dh }: Measure): string {
  const dist = `${cells} ${cells === 1 ? 'célula' : 'células'}`
  return dh === 0 ? dist : `${dist} · ${dh > 0 ? '+' : '−'}${Math.abs(dh)} de altura`
}

/** Drops points closer than MIN_STEP to the previous one and decimates evenly down to MAX_POINTS. */
export function simplify(points: Point[]): Point[] {
  const kept: Point[] = []
  for (const p of points) {
    const q = kept[kept.length - 1]
    if (!q || Math.hypot(p.x - q.x, p.z - q.z) >= MIN_STEP) kept.push(p)
  }
  const last = points[points.length - 1]
  if (kept.length > 1 && kept[kept.length - 1] !== last) kept[kept.length - 1] = last
  if (kept.length <= MAX_POINTS) return kept
  const step = (kept.length - 1) / (MAX_POINTS - 1)
  return Array.from({ length: MAX_POINTS }, (_, i) => kept[Math.round(i * step)])
}

/** World-space polyline hugging the terrain: segments are subdivided so the height follows the cells. */
export function projectStroke(grid: MapGrid, points: readonly Point[], lift = LIFT): [number, number, number][] {
  const out: [number, number, number][] = []
  points.forEach((p, i) => {
    const prev = points[i - 1]
    const n = prev ? Math.ceil(Math.hypot(p.x - prev.x, p.z - prev.z) / SAMPLE) : 1
    for (let k = prev ? 1 : 0; k <= (prev ? n : 0); k++) {
      const x = prev ? prev.x + ((p.x - prev.x) * k) / n : p.x
      const z = prev ? prev.z + ((p.z - prev.z) * k) / n : p.z
      out.push([x, heightAt(grid, x, z) + lift, z])
    }
  })
  return out
}
