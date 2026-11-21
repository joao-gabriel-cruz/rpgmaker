import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import {
  addLog, BUBBLE_MS, dropBubble, putBubble, REACTION_MS,
  type Bubble, type LogEntry, type Reaction, type SaidEvent,
} from './presence'

export interface WheelState {
  characterId: string
  /** Screen position (px) where the radial menu opens. */
  x: number
  y: number
}

export interface PresenceApi {
  bubbles: Bubble[]
  reactions: Reaction[]
  log: LogEntry[]
  speakOpen: boolean
  setSpeakOpen: (open: boolean) => void
  /** The reaction wheel opened from the dock button or the G key (for the active character). */
  reactOpen: boolean
  setReactOpen: (open: boolean) => void
  /** The reaction wheel opened by holding the own token. */
  wheel: WheelState | null
  openWheel: (w: WheelState) => void
  closeWheel: () => void
}

const Ctx = createContext<PresenceApi | null>(null)

/** Speech bubbles, emoji reactions and the speech log of the table; null outside the provider. */
export const usePresenceState = () => useContext(Ctx)

type Later = (fn: () => void, ms: number) => void

function useSaid(subscribe: ReturnType<typeof useRealtime>['subscribe'], later: Later, seq: { current: number }) {
  const [bubbles, setBubbles] = useState<Bubble[]>([])
  const [log, setLog] = useState<LogEntry[]>([])
  useEffect(() => subscribe('chat.said', (p) => {
    const e = p as SaidEvent
    if (!e.text) return
    const id = ++seq.current
    setBubbles((l) => putBubble(l, { id, characterId: e.characterId, name: e.name, text: e.text as string }))
    setLog((l) => addLog(l, { id, name: e.name, text: e.text as string, at: e.at }))
    later(() => setBubbles((l) => dropBubble(l, id)), BUBBLE_MS)
  }), [subscribe, later, seq])
  return { bubbles, log }
}

function useReacted(subscribe: ReturnType<typeof useRealtime>['subscribe'], later: Later, seq: { current: number }) {
  const [reactions, setReactions] = useState<Reaction[]>([])
  useEffect(() => subscribe('chat.reacted', (p) => {
    const e = p as SaidEvent
    if (!e.emoji) return
    const id = ++seq.current
    setReactions((l) => [...l, { id, characterId: e.characterId, emoji: e.emoji as string }])
    later(() => setReactions((l) => l.filter((r) => r.id !== id)), REACTION_MS)
  }), [subscribe, later, seq])
  return reactions
}

// Timeouts that expire with the provider so no state update lands after unmount.
function useLater(): Later {
  const timers = useRef(new Set<ReturnType<typeof setTimeout>>())
  useEffect(() => {
    const set = timers.current
    return () => set.forEach(clearTimeout)
  }, [])
  return useCallback((fn, ms) => {
    const t = setTimeout(() => { timers.current.delete(t); fn() }, ms)
    timers.current.add(t)
  }, [])
}

function useEvents() {
  const { subscribe } = useRealtime()
  const later = useLater()
  const seq = useRef(0)
  const said = useSaid(subscribe, later, seq)
  const reactions = useReacted(subscribe, later, seq)
  return { ...said, reactions }
}

export function PresenceProvider({ children }: { children: ReactNode }) {
  const events = useEvents()
  const [speakOpen, setSpeakOpen] = useState(false)
  const [reactOpen, setReactOpen] = useState(false)
  const [wheel, setWheel] = useState<WheelState | null>(null)
  const closeWheel = useCallback(() => setWheel(null), [])
  const value = useMemo<PresenceApi>(
    () => ({ ...events, speakOpen, setSpeakOpen, reactOpen, setReactOpen, wheel, openWheel: setWheel, closeWheel }),
    [events, speakOpen, reactOpen, wheel, closeWheel],
  )
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}
