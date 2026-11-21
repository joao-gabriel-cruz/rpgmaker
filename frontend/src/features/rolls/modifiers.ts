import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'
import type { InstinctKey } from '../sheet-assimilacao/types'

/** What the roll panel can still use for a character (from `GET /rolls/state`). */
export interface RollState {
  occupationUsed: boolean
  markingEventUsed: boolean
  sessionActive: boolean
  difficultyD12: number
}

export interface Modifiers {
  effort: boolean
  useOccupation: boolean
  useMarkingEvent: boolean
  assimilated: boolean
  secondInstinct?: InstinctKey
  difficultyD12: number
}

export const NO_MODIFIERS: Modifiers = {
  effort: false, useOccupation: false, useMarkingEvent: false, assimilated: false, difficultyD12: 0,
}

/** Payload fields of `roll.request` for the chosen modifiers. */
export function modifiersPayload(m: Modifiers, isGm: boolean) {
  return {
    effort: m.effort,
    useOccupation: m.useOccupation,
    useMarkingEvent: m.useMarkingEvent,
    ...(m.assimilated && m.secondInstinct ? { secondInstinct: m.secondInstinct } : {}),
    ...(isGm && m.difficultyD12 > 0 ? { difficultyD12: m.difficultyD12 } : {}),
  }
}

/** Loads the character's roll state and refreshes it when rolls, sessions or difficulty change. */
export function useRollState(campaignId: string, characterId: string) {
  const { subscribe } = useRealtime()
  const qc = useQueryClient()
  const key = ['campaigns', campaignId, 'roll-state', characterId]
  const query = useQuery({
    queryKey: key,
    queryFn: () => api.get<RollState>(`/campaigns/${campaignId}/rolls/state?characterId=${characterId}`),
  })
  useEffect(() => {
    const refresh = () => void qc.invalidateQueries({ queryKey: key })
    const offs = ['roll.rolled', 'session.updated', 'roll.difficulty.updated'].map((t) => subscribe(t, refresh))
    return () => offs.forEach((off) => off())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [subscribe, qc, campaignId, characterId])
  return query.data
}

const FLAG_LABELS: [keyof Pick<Modifiers, 'effort' | 'useOccupation' | 'useMarkingEvent'>, string][] = [
  ['effort', 'Empenho'], ['useOccupation', 'Ocupação'], ['useMarkingEvent', 'Evento Marcante'],
]

/** "Empenho, Ocupação, +2 d12 dificuldade" for a roll request; empty when no modifier was used. */
export function modifiersText(r?: {
  effort?: boolean; useOccupation?: boolean; useMarkingEvent?: boolean; secondInstinct?: string; difficultyD12?: number
}): string {
  if (!r) return ''
  const parts = FLAG_LABELS.filter(([k]) => r[k]).map(([, l]) => l)
  if (r.secondInstinct) parts.push('Assimilada')
  if (r.difficultyD12) parts.push(`+${r.difficultyD12} d12 dificuldade`)
  return parts.join(', ')
}
