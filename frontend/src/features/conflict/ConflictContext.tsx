import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'
import type { Conflict, Triggered } from './conflict'

export interface ConflictState {
  campaignId: string
  isGm: boolean
  conflict: Conflict | null
  triggered: Triggered[]
  /** NPC character id of the threat that just rolled, for a few seconds. */
  acting: string | null
}

const ConflictContext = createContext<ConflictState | null>(null)

/** The conflict state, or null outside a ConflictProvider. */
export const useConflict = () => useContext(ConflictContext)

/** Loads the role-filtered conflict and keeps it live from conflict.* events. */
export function ConflictProvider({ campaignId, isGm, children }: { campaignId: string; isGm: boolean; children: ReactNode }) {
  const { status, subscribe } = useRealtime()
  const [conflict, setConflict] = useState<Conflict | null>(null)
  const [triggered, setTriggered] = useState<Triggered[]>([])
  const [acting, setActing] = useState<string | null>(null)
  useEffect(() => {
    if (status !== 'open') return
    api.get<{ conflict: Conflict | null }>(`/campaigns/${campaignId}/conflict`).then((r) => setConflict(r.conflict), () => {})
  }, [campaignId, status])
  useEffect(() => subscribe('conflict.updated', (p) => setConflict((p as { conflict: Conflict | null }).conflict)), [subscribe])
  useEffect(() => subscribe('conflict.activation.triggered', (p) => setTriggered((t) => [...t, p as Triggered])), [subscribe])
  useEffect(() => subscribe('conflict.threat.rolled', (p) => {
    setActing((p as { npcCharacterId?: string }).npcCharacterId ?? null)
    setTimeout(() => setActing(null), 4000)
  }), [subscribe])
  return <ConflictContext.Provider value={{ campaignId, isGm, conflict, triggered, acting }}>{children}</ConflictContext.Provider>
}
