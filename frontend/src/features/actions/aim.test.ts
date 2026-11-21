import { PerspectiveCamera, Vector3 } from 'three'
import { expect, it } from 'vitest'
import type { MapObject } from '../map3d/objects/model'
import type { Token } from '../tokens/tokens'
import type { GameAction } from './actions'
import { resolveAim } from './aim'
import { nearestSnap } from './AimLayer'

const base: GameAction = {
  id: 'a', name: 'A', icon: 'zap', category: 'combat', target: 'token', outcome: 'damage', animation: 'melee_strike', range: 2,
  rolls: [{ instinct: 'potency', label: 'P' }], enabled: true, builtin: true,
}
const tok = (id: string, x: number, z: number): Token => ({ id, sceneId: 's', characterId: id, x, z, name: id, portraitUrl: null, ownerId: null })
const tokens = [tok('me', 3, 3), tok('foe', 4, 3), tok('far', 8, 8)]
const crate = { id: 'crate1', prefab: 'crate', x: 5, z: 5, rotation: 0, props: {} } as MapObject

it('picks the token under the cursor within range', () => {
  expect(resolveAim(base, { x: 4, z: 3 }, tokens[0], tokens, [])).toEqual({ valid: true, target: { kind: 'token', tokenId: 'foe', x: 4, z: 3 } })
})

it('rejects empty cells, far targets and attacking yourself', () => {
  expect(resolveAim(base, { x: 5, z: 3 }, tokens[0], tokens, [])).toMatchObject({ valid: false, reason: 'Escolha um personagem' })
  expect(resolveAim(base, { x: 8, z: 8 }, tokens[0], tokens, [])).toMatchObject({ valid: false, reason: 'Fora de alcance (2 células)' })
  expect(resolveAim(base, { x: 3, z: 3 }, tokens[0], tokens, [])).toMatchObject({ valid: false, reason: 'Não dá para atacar a si mesmo' })
})

it('allows healing yourself and targets objects or cells for search actions', () => {
  const aid = { ...base, outcome: 'heal' as const, range: 1 }
  expect(resolveAim(aid, { x: 3, z: 3 }, tokens[0], tokens, []).valid).toBe(true)
  const search = { ...base, target: 'object' as const, outcome: 'loot' as const, range: 3 }
  expect(resolveAim(search, { x: 5, z: 5 }, tokens[0], tokens, [crate]).target).toEqual({ kind: 'object', objectId: 'crate1', x: 5, z: 5 })
  expect(resolveAim(search, { x: 4, z: 4 }, tokens[0], tokens, [crate]).target).toEqual({ kind: 'cell', x: 4, z: 4 })
})

it('snaps the pointer to the figure it is over even though the ground ray lands behind it', () => {
  const cam = new PerspectiveCamera(50, 1, 0.1, 100)
  cam.position.set(4, 8, 12)
  cam.lookAt(4, 0, 4)
  cam.updateMatrixWorld()
  cam.updateProjectionMatrix()
  const dom = { getBoundingClientRect: () => ({ left: 0, top: 0, width: 800, height: 800 }) } as HTMLElement
  const at = (px: number, py: number) => ({ clientX: px, clientY: py }) as PointerEvent
  const snaps = [{ cell: { x: 3, z: 3 }, height: 0.55 }, { cell: { x: 4, z: 3 }, height: 0.55 }]
  const p = new Vector3(4.5, 0.55, 3.5).project(cam)
  const px = ((p.x + 1) / 2) * 800
  const py = ((1 - p.y) / 2) * 800
  expect(nearestSnap(snaps, at(px + 5, py - 4), dom, cam)).toEqual({ x: 4, z: 3 })
  expect(nearestSnap(snaps, at(px + 400, py), dom, cam)).toBeNull()
})
