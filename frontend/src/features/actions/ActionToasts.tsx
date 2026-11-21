import { useEffect, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'

interface Resolved {
  actionRunId: string
  actionName?: string
  actorName?: string
  outcome: { kind: string; text: string }
  result: { success: number; adaptation: number; pressure: number }
}

interface Toast {
  id: string
  text: string
  detail: string
}

const SHOW_MS = 5000
const MAX = 3

/** Everyone sees what an action achieved (the Mestre's suggestion details stay in the Efeitos tab). */
export function ActionToasts() {
  const { subscribe } = useRealtime()
  const [toasts, setToasts] = useState<Toast[]>([])
  useEffect(() => subscribe('action.resolved', (p) => {
    const e = p as Resolved
    if (!e.outcome?.text) return
    const detail = `${e.actionName ?? 'Ação'} · ${e.result.success} ✔ ${e.result.adaptation} ◆ ${e.result.pressure} ✖`
    setToasts((l) => [...l.slice(-(MAX - 1)), { id: e.actionRunId, text: e.outcome.text, detail }])
    setTimeout(() => setToasts((l) => l.filter((t) => t.id !== e.actionRunId)), SHOW_MS)
  }), [subscribe])
  if (toasts.length === 0) return null
  return (
    <div className="pointer-events-none fixed left-1/2 top-28 z-40 flex -translate-x-1/2 flex-col items-center gap-2">
      {toasts.map((t) => (
        <div key={t.id} role="status" className="rounded-xl bg-zinc-900/90 px-4 py-2 text-center shadow-lg ring-1 ring-white/10 backdrop-blur motion-safe:animate-[tray-in_160ms_ease-out]">
          <p className="text-sm font-medium text-zinc-100">{t.text}</p>
          <p className="text-xs text-zinc-400">{t.detail}</p>
        </div>
      ))}
    </div>
  )
}
