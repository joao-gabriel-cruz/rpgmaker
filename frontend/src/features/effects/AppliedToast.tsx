import { useEffect, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { describe, type Suggestion } from './effects'

type Applied = Pick<Suggestion, 'characterName' | 'source' | 'effect' | 'amount' | 'message'>

/** Toast (slot: toasts) telling a player that the Mestre applied a terrain effect to their character. */
export function AppliedToast() {
  const { subscribe } = useRealtime()
  const [msg, setMsg] = useState('')

  useEffect(() => subscribe('effect.applied', (p) => {
    const a = p as Applied
    setMsg(`${a.source}: ${a.characterName} sofreu ${describe(a)}${a.message ? ` — ${a.message}` : ''}`)
  }), [subscribe])
  useEffect(() => {
    if (!msg) return
    const t = setTimeout(() => setMsg(''), 6000)
    return () => clearTimeout(t)
  }, [msg])

  if (!msg) return null
  return <div role="status" className="fixed left-1/2 top-16 z-50 -translate-x-1/2 rounded-xl bg-amber-600 px-4 py-2 text-sm text-black shadow-lg">{msg}</div>
}
