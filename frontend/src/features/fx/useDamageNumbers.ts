import { useEffect, useRef } from 'react'
import { fxBus } from '../../lib/fxBus'
import { useCharacters } from '../characters/characters'
import { useRealtime } from '../realtime/realtimeContext'
import type { Token } from '../tokens/tokens'

const damageOf = (sheet: unknown): number | undefined => (sheet as { health?: { damage?: number } } | undefined)?.health?.damage

/** Emits `fx:number` with the delta whenever a character's damage changes (positive = hurt, negative = healed). */
export function useDamageNumbers(campaignId: string, tokens: readonly Token[]) {
  const { subscribe } = useRealtime()
  const { data } = useCharacters(campaignId)
  const known = useRef(new Map<string, number>())
  const tokensRef = useRef(tokens)
  useEffect(() => { tokensRef.current = tokens }, [tokens])
  useEffect(() => {
    for (const c of Array.isArray(data) ? data : []) {
      const d = damageOf(c.sheet)
      if (d !== undefined && !known.current.has(c.id)) known.current.set(c.id, d)
    }
  }, [data])
  useEffect(() => subscribe('character.updated', (p) => {
    const c = p as { id: string; sheet?: unknown }
    const now = damageOf(c.sheet)
    if (now === undefined) return
    const before = known.current.get(c.id)
    known.current.set(c.id, now)
    const token = tokensRef.current.find((t) => t.characterId === c.id)
    if (before !== undefined && before !== now && token) fxBus.emit('fx:number', { x: token.x, z: token.z, amount: now - before })
  }), [subscribe])
}
