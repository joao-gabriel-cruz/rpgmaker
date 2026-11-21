import { objectsAt, type MapObject } from '../map3d/objects/model'
import type { Token } from '../tokens/tokens'
import type { ActionTarget, GameAction } from './actions'

export interface AimResult {
  valid: boolean
  target?: ActionTarget
  /** Why the cell cannot be chosen (shown near the cursor). */
  reason?: string
}

interface Cell {
  x: number
  z: number
}

const dist = (a: Cell, b: Cell) => Math.max(Math.abs(a.x - b.x), Math.abs(a.z - b.z))

function targetAt(action: GameAction, cell: Cell, tokens: readonly Token[], objects: readonly MapObject[]): AimResult {
  if (action.target === 'token') {
    const t = tokens.find((k) => k.x === cell.x && k.z === cell.z)
    return t ? { valid: true, target: { kind: 'token', tokenId: t.id, x: t.x, z: t.z } } : { valid: false, reason: 'Escolha um personagem' }
  }
  if (action.target === 'object') {
    const o = objectsAt(objects, cell.x, cell.z)[0]
    if (o) return { valid: true, target: { kind: 'object', objectId: o.id, x: o.x, z: o.z } }
  }
  return { valid: true, target: { kind: 'cell', x: cell.x, z: cell.z } }
}

/** Resolves the cell under the cursor into a target for the action, honoring the range (Chebyshev, like the server). */
export function resolveAim(action: GameAction, cell: Cell, actor: Cell, tokens: readonly Token[], objects: readonly MapObject[]): AimResult {
  const r = targetAt(action, cell, tokens, objects)
  if (!r.valid || !r.target) return r
  if (action.outcome === 'damage' && r.target.tokenId && cell.x === actor.x && cell.z === actor.z) {
    return { valid: false, reason: 'Não dá para atacar a si mesmo' }
  }
  if (action.range && dist(cell, actor) > action.range) return { valid: false, reason: `Fora de alcance (${action.range} células)` }
  return r
}
