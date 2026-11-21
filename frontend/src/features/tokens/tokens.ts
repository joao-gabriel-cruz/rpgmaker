import { api } from '../../lib/api'
import type { Appearance } from '../miniature/appearance'
import { getCell, inBounds, type MapGrid } from '../map3d/grid'

export interface Token {
  id: string
  sceneId: string
  characterId: string
  x: number
  z: number
  name: string
  portraitUrl: string | null
  ownerId: string | null
  /** Miniature description; absent or null keeps the portrait standee. */
  appearance?: Appearance | null
  assimilationLevel?: number | null
}

export interface PingMarker {
  id: number
  x: number
  z: number
  userName: string
}

export interface CameraTarget {
  x: number
  z: number
  n: number
}

export const tokensUrl = (campaignId: string, sceneId: string) => `/campaigns/${campaignId}/scenes/${sceneId}/tokens`

export async function loadTokens(campaignId: string, sceneId: string): Promise<Token[]> {
  const data = await api.get<Token[]>(tokensUrl(campaignId, sceneId))
  return Array.isArray(data) ? data : []
}

export const applyPlaced = (list: Token[], sceneId: string, t: Token): Token[] =>
  t.sceneId !== sceneId || list.some((o) => o.id === t.id) ? list : [...list, t]

export const applyRemoved = (list: Token[], id: string): Token[] => list.filter((t) => t.id !== id)

export const applyMoved = (list: Token[], id: string, x: number, z: number): Token[] =>
  list.map((t) => (t.id === id ? { ...t, x, z } : t))

/** The sheet fields of a `character.updated` event that tokens mirror. */
export interface CharacterEvent {
  id: string
  name?: string
  portraitUrl?: string | null
  sheet?: { appearance?: Appearance | null; tugOfWar?: { assimilationLevel?: number } }
}

export const applyCharacter = (list: Token[], c: CharacterEvent): Token[] =>
  list.map((t) => t.characterId !== c.id ? t : {
    ...t,
    name: c.name ?? t.name,
    portraitUrl: c.portraitUrl === undefined ? t.portraitUrl : c.portraitUrl,
    appearance: c.sheet?.appearance ?? null,
    assimilationLevel: c.sheet?.tugOfWar?.assimilationLevel ?? null,
  })

/** Whether the viewer may drag this token: the Mestre moves all, players only their own. */
export const canMoveToken = (t: Token, userId: string, isGm: boolean) => isGm || t.ownerId === userId

const COLORS = ['#34d399', '#60a5fa', '#f59e0b', '#f472b6', '#a78bfa', '#f87171', '#2dd4bf', '#a3e635']

/** Stable color per player; characters without an owner (NPCs) are gray. */
export function playerColor(ownerId: string | null): string {
  if (!ownerId) return '#71717a'
  let h = 0
  for (const ch of ownerId) h = (h * 31 + ch.charCodeAt(0)) >>> 0
  return COLORS[h % COLORS.length]
}

export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return '?'
  return (parts.length === 1 ? parts[0].slice(0, 2) : parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}

/** Cells ordered by distance from the scene center, skipping occupied or terrain-blocked ones. */
export function spawnCandidates(grid: MapGrid, tokens: readonly Token[], limit = 12): { x: number; z: number }[] {
  const taken = new Set(tokens.map((t) => `${t.x},${t.z}`))
  const cx = Math.floor(grid.width / 2)
  const cz = Math.floor(grid.depth / 2)
  const all: { x: number; z: number; d: number }[] = []
  for (let z = 0; z < grid.depth; z++) {
    for (let x = 0; x < grid.width; x++) {
      if (!taken.has(`${x},${z}`) && inBounds(grid, x, z) && getCell(grid, x, z)) {
        all.push({ x, z, d: (x - cx) ** 2 + (z - cz) ** 2 })
      }
    }
  }
  return all.sort((a, b) => a.d - b.d).slice(0, limit).map(({ x, z }) => ({ x, z }))
}
