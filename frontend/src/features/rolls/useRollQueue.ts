import { useCallback, useEffect, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import type { FinalizedEvent, RollEntry, RollResult } from './rolls'

export interface ActiveRoll {
  entry: RollEntry
  result?: RollResult
}

const TRAY_KINDS = new Set(['d4', 'd6', 'd8', 'd10', 'd12', 'd20'])

/** Whether the tray can show this roll: Assimilação and threat rolls, and generic ones with known dice. */
const trayable = (e: RollEntry) =>
  e.kind === 'generic' ? e.dice.length > 0 && e.dice.every((d) => TRAY_KINDS.has(d.kind)) : e.kind === 'assimilacao' || e.kind === 'threat'

/** Generic and threat rolls are final from the start; Assimilação rolls wait for the kept dice. */
export const isSettled = (a: ActiveRoll) => !!a.result || a.entry.kind === 'generic'

/** Rolls waiting for the tray, oldest first; the first one is on screen. */
export function useRollQueue() {
  const { subscribe } = useRealtime()
  const [queue, setQueue] = useState<ActiveRoll[]>([])
  useEffect(() => subscribe('roll.rolled', (p) => {
    const entry = p as RollEntry
    if (!trayable(entry)) return
    const item: ActiveRoll = entry.kind === 'threat' ? { entry, result: entry.result } : { entry }
    setQueue((q) => (q.some((a) => a.entry.rollId === entry.rollId) ? q : [...q, item]))
  }), [subscribe])
  useEffect(() => subscribe('roll.finalized', (p) => {
    const ev = p as FinalizedEvent
    setQueue((q) => q.map((a) => (a.entry.rollId === ev.rollId ? { ...a, result: ev.result } : a)))
  }), [subscribe])
  const close = useCallback(() => setQueue((q) => q.slice(1)), [])
  return { current: queue[0] as ActiveRoll | undefined, waiting: Math.max(0, queue.length - 1), close }
}
