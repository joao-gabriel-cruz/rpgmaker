import { useCallback, useEffect, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import type { FinalizedEvent, RollEntry, RollResult } from './rolls'

export interface ActiveRoll {
  entry: RollEntry
  result?: RollResult
}

/** Rolls waiting for the tray, oldest first; the first one is on screen. */
export function useRollQueue() {
  const { subscribe } = useRealtime()
  const [queue, setQueue] = useState<ActiveRoll[]>([])
  useEffect(() => subscribe('roll.rolled', (p) => {
    const entry = p as RollEntry
    if (entry.kind !== 'assimilacao' && entry.kind !== 'threat') return
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
