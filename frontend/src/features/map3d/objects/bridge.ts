import { MAX_HEIGHT, MIN_HEIGHT, getCell, surfaceHeight, type Coord, type MapGrid } from '../grid'
import { objectsAt, occupiedCells, type MapObject } from './model'

const DIRS = [[1, 0], [-1, 0], [0, 1], [0, -1]]
const key = (c: Coord) => `${c.x},${c.z}`
const clamp = (v: number) => Math.min(MAX_HEIGHT, Math.max(MIN_HEIGHT, Math.round(v)))

function explicitDeck(o: MapObject): number | undefined {
  const raw = o.props.deck
  return raw !== undefined && Number.isFinite(Number(raw)) ? clamp(Number(raw)) : undefined
}

/**
 * Deck height of every bridge cell (mirrors the backend): a chain of adjacent bridge cells shares the
 * highest terrain touching it, unless the cell's own `deck` prop (0-10) says otherwise.
 */
export function bridgeDecks(grid: MapGrid, objects: readonly MapObject[]): Map<string, number> {
  const cells = new Map<string, number | undefined>()
  for (const o of objects) if (o.prefab === 'bridge') for (const c of occupiedCells(o)) cells.set(key(c), explicitDeck(o))
  const out = new Map<string, number>()
  for (const start of cells.keys()) {
    if (out.has(start)) continue
    const chain = [start]
    const seen = new Set(chain)
    let best = 0
    for (let k = 0; k < chain.length; k++) {
      const [x, z] = chain[k].split(',').map(Number)
      best = Math.max(best, getCell(grid, x, z)?.height ?? 0)
      for (const [dx, dz] of DIRS) {
        const n = key({ x: x + dx, z: z + dz })
        if (!cells.has(n)) best = Math.max(best, getCell(grid, x + dx, z + dz)?.height ?? 0)
        else if (!seen.has(n)) { seen.add(n); chain.push(n) }
      }
    }
    for (const c of chain) out.set(c, cells.get(c) ?? best)
  }
  return out
}

/** Deck height under a bridge object (its first cell); `objects` supplies the neighbouring bridge cells. */
export function bridgeDeck(grid: MapGrid, o: MapObject, objects: readonly MapObject[] = []): number {
  const all = objects.some((x) => x.id === o.id) ? objects : [...objects, o]
  return bridgeDecks(grid, all).get(key({ x: o.x, z: o.z })) ?? 0
}

/** Height a token stands at: the bridge deck when a bridge spans the cell, else the terrain surface. */
export function standHeight(grid: MapGrid, objects: readonly MapObject[] | undefined, x: number, z: number): number {
  const bridge = objects && objectsAt(objects, x, z).find((o) => o.prefab === 'bridge')
  const deck = bridge && bridgeDecks(grid, objects).get(key({ x, z }))
  return deck !== undefined ? surfaceHeight(deck) : surfaceHeight(getCell(grid, x, z)?.height ?? 1)
}
