import { useEffect, useRef, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import type { ClaudeChange } from './mcp'

/** Toast for the Mestre only: what Claude just changed, with an undo button. */
export function ClaudeChangeToast() {
  const { subscribe, send } = useRealtime()
  const [change, setChange] = useState<ClaudeChange | null>(null)
  const [error, setError] = useState('')
  const currentId = useRef<string | null>(null)
  currentId.current = change?.id ?? null

  useEffect(() => subscribe('mcp.changed', (p) => {
    const { changes, latest } = p as { changes: ClaudeChange[]; latest?: ClaudeChange }
    if (latest) {
      setError('')
      setChange(latest)
    } else if (currentId.current && !changes.some((c) => c.id === currentId.current)) {
      setChange(null)
    }
  }), [subscribe])

  useEffect(() => {
    if (!change) return
    const t = setTimeout(() => setChange(null), 8000)
    return () => clearTimeout(t)
  }, [change])

  if (!change) return null
  const undo = () => send('mcp.undo', { changeId: change.id }).catch((e: Error) => setError(e.message))
  return (
    <div role="status" className="fixed left-1/2 top-28 z-50 flex -translate-x-1/2 items-center gap-3 rounded-xl bg-emerald-600 px-4 py-2 text-sm text-black shadow-lg">
      <span>{error || `Claude: ${change.summary}`}</span>
      <button type="button" onClick={undo} className="rounded bg-black/20 px-2 py-0.5 font-medium hover:bg-black/30">Desfazer</button>
    </div>
  )
}
