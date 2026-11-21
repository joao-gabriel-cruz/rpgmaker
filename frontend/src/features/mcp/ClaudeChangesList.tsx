import { useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { relativeTime, useClaudeChanges } from './mcp'

export function ClaudeChangesList({ campaignId }: { campaignId: string }) {
  const changes = useClaudeChanges(campaignId)
  const { send } = useRealtime()
  const [error, setError] = useState('')

  const undo = (changeId: string) => {
    setError('')
    send('mcp.undo', { changeId }).catch((e: Error) => setError(e.message))
  }

  return (
    <section className="flex flex-col gap-2">
      <h3 className="text-sm font-medium text-zinc-200">Mudanças recentes do Claude</h3>
      {changes.length === 0 && <p className="text-sm text-zinc-400">Nenhuma mudança ainda.</p>}
      <ul className="flex flex-col gap-1">
        {changes.map((c) => (
          <li key={c.id} className="flex items-center gap-3 rounded bg-zinc-800/60 px-3 py-2 text-sm">
            <span className="flex-1">{c.summary}</span>
            <span className="text-xs text-zinc-400">{relativeTime(c.at)}</span>
            <button onClick={() => undo(c.id)} className="rounded bg-zinc-800 px-2 py-1 text-xs hover:bg-zinc-700">
              Desfazer
            </button>
          </li>
        ))}
      </ul>
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
    </section>
  )
}
