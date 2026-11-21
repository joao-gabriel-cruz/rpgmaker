import { useCallback, useEffect, useRef, useState, type Dispatch, type SetStateAction } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'
import { occupiedCells, type MapObject } from '../map3d/objects/model'
import type { Token } from '../tokens/tokens'
import type { ItemEffects } from './items'

export interface SheetItemView {
  name: string
  icon?: string
  color?: string
  kind?: string
  description?: string
  effects?: ItemEffects
}

export interface Loot {
  id: string
  sceneId: string
  x: number
  z: number
  objectId: string | null
  itemId: string | null
  item: SheetItemView
  qty: number
  difficulty: number
  hidden: boolean
  foundBy: string | null
}

export interface FoundFlash {
  id: string
  x: number
  z: number
  who: string
  itemName: string
}

const FLASH_MS = 3500

/** What a viewer may see: the Mestre everything, players only loot that is not hidden. */
export const visibleLoot = (list: Loot[], isGm: boolean) => (isGm ? list : list.filter((l) => !l.hidden))

const adjacent = (t: Pick<Token, 'x' | 'z'>, l: Pick<Loot, 'x' | 'z'>) => Math.abs(t.x - l.x) <= 1 && Math.abs(t.z - l.z) <= 1

/** The viewer's own token next to the loot (so it can be picked up), if any. */
export function pickerFor(l: Loot, tokens: readonly Token[], userId: string): Token | undefined {
  return l.hidden ? undefined : tokens.find((t) => t.ownerId === userId && adjacent(t, l))
}

/** The object whose footprint covers the cell, if any (so loot goes "inside" crates, cars, ruins). */
export const objectAt = (objects: readonly MapObject[], x: number, z: number) =>
  objects.find((o) => occupiedCells(o).some((c) => c.x === x && c.z === z))

const upsert = (list: Loot[], l: Loot) => (list.some((x) => x.id === l.id) ? list.map((x) => (x.id === l.id ? l : x)) : [...list, l])

interface FoundEvent {
  loot: Loot
  characterName: string
  left: number
  item: SheetItemView
}

function flash(set: Dispatch<SetStateAction<FoundFlash[]>>, id: string, e: FoundEvent) {
  set((cur) => [...cur, { id, x: e.loot.x, z: e.loot.z, who: e.characterName, itemName: e.item?.name ?? e.loot.item.name }])
  setTimeout(() => set((cur) => cur.filter((f) => f.id !== id)), FLASH_MS)
}

/** Scene loot kept live by loot.placed / loot.removed / loot.found; the server never sends players hidden loot. */
export function useSceneLoot(campaignId: string, sceneId: string) {
  const { subscribe, send, status } = useRealtime()
  const [loot, setLoot] = useState<Loot[]>([])
  const [flashes, setFlashes] = useState<FoundFlash[]>([])
  const seq = useRef(0)
  useEffect(() => {
    if (status !== 'open') return
    api.get<{ loot: Loot[] }>(`/campaigns/${campaignId}/scenes/${sceneId}/loot`).then((r) => setLoot(Array.isArray(r?.loot) ? r.loot : []), () => {})
  }, [campaignId, sceneId, status])
  useEffect(() => subscribe('loot.placed', (p) => {
    const l = (p as { loot: Loot }).loot
    if (l.sceneId === sceneId) setLoot((cur) => upsert(cur, l))
  }), [subscribe, sceneId])
  useEffect(() => subscribe('loot.removed', (p) => {
    const e = p as { id: string }
    setLoot((cur) => cur.filter((l) => l.id !== e.id))
  }), [subscribe])
  useEffect(() => subscribe('loot.found', (p) => {
    const e = p as FoundEvent
    if (e.loot.sceneId !== sceneId) return
    setLoot((cur) => (e.left > 0 ? upsert(cur, e.loot) : cur.filter((l) => l.id !== e.loot.id)))
    flash(setFlashes, `${e.loot.id}-${++seq.current}`, e)
  }), [subscribe, sceneId])
  const pickup = useCallback((lootId: string, characterId: string) => send('loot.pickup', { lootId, characterId }), [send])
  const remove = useCallback((lootId: string) => send('loot.remove', { lootId }), [send])
  const place = useCallback((p: { x: number; z: number; objectId?: string; itemId: string; difficulty: number }) =>
    send('loot.place', { sceneId, qty: 1, ...p }), [send, sceneId])
  return { loot, flashes, pickup, remove, place }
}

export type SceneLoot = ReturnType<typeof useSceneLoot>
