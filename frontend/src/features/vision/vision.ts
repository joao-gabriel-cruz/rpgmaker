import { api } from '../../lib/api'
import { getCell, type Coord, type MapGrid } from '../map3d/grid'
import type { Token } from '../tokens/tokens'
import { cellsWhere, decodeBits, hasBit } from './bits'

export type FogMode = 'off' | 'manual' | 'dynamic'

export const FOG_MODES: { id: FogMode; label: string }[] = [
  { id: 'off', label: 'Sem névoa' },
  { id: 'manual', label: 'Manual' },
  { id: 'dynamic', label: 'Visão dinâmica' },
]

export const DEFAULT_RADIUS = 12

/** Decoded fog of the scene: players get `visible`, the Mestre gets `revealed` and the radius. */
export interface VisionState {
  mode: FogMode
  visionRadius: number
  visible: Uint8Array | null
  revealed: Uint8Array | null
}

export const NO_FOG: VisionState = { mode: 'off', visionRadius: DEFAULT_RADIUS, visible: null, revealed: null }

export interface Wire {
  mode?: FogMode
  visionRadius?: number
  visible?: string
  revealed?: string
}

export const visionUrl = (campaignId: string, sceneId: string) => `/campaigns/${campaignId}/scenes/${sceneId}/vision`

export function toState(w: Wire, g: MapGrid, isGm: boolean): VisionState {
  const mode = w.mode ?? 'off'
  return {
    mode,
    visionRadius: w.visionRadius || DEFAULT_RADIUS,
    visible: !isGm && mode !== 'off' ? decodeBits(w.visible, g.width, g.depth) : null,
    revealed: isGm ? decodeBits(w.revealed, g.width, g.depth) : null,
  }
}

export async function loadVision(campaignId: string, sceneId: string, g: MapGrid, isGm: boolean): Promise<VisionState> {
  return toState(await api.get<Wire>(visionUrl(campaignId, sceneId)), g, isGm)
}

/** Merges a vision.updated (player) or fog.updated (Mestre) event into the state. */
export function applyEvent(prev: VisionState, w: Wire, g: MapGrid, isGm: boolean): VisionState {
  const next = toState(w, g, isGm)
  return isGm ? { ...next, visionRadius: w.visionRadius || prev.visionRadius } : next
}

/** Cells the players cannot see, as the viewer should draw them (null = nothing to cover). */
export function fogCover(s: VisionState, g: MapGrid, isGm: boolean): { cells: Coord[]; gm: boolean } | null {
  if (s.mode === 'off') return null
  if (isGm) {
    if (!s.revealed) return null
    // Manual: show what stays hidden; dynamic: the vision is per player, so mark the manual overrides.
    return { cells: cellsWhere(s.revealed, g.width, g.depth, s.mode === 'dynamic'), gm: true }
  }
  return s.visible ? { cells: cellsWhere(s.visible, g.width, g.depth, false), gm: false } : null
}

/** Players only draw tokens in sight (their own always); the server already filters, this is defense in depth. */
export function visibleTokens(list: Token[], s: VisionState, g: MapGrid, userId: string, isGm: boolean): Token[] {
  if (isGm || s.mode === 'off' || !s.visible) return list
  const vis = s.visible
  return list.filter((t) => t.ownerId === userId || hasBit(vis, g.width, t.x, t.z))
}

export interface FogInstance {
  x: number
  y: number
  z: number
  h: number
}

const COVER_ABOVE = 4

/** Box per covered cell: players get an opaque column above any terrain/object, the Mestre a thin film on top. */
export function fogInstances(g: MapGrid, cells: Coord[], gm: boolean): FogInstance[] {
  return cells.map((c) => {
    const top = getCell(g, c.x, c.z)?.height ?? 0
    const h = gm ? 0.06 : top + COVER_ABOVE
    return { x: c.x + 0.5, y: gm ? top + 0.1 : h / 2, z: c.z + 0.5, h }
  })
}
