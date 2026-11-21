import { useState } from 'react'
import { useMe } from '../auth/auth'
import { useRealtime } from '../realtime/realtimeContext'
import { useConflict } from './ConflictContext'

/** "Atribuir a objetivo…" for a finalized roll of the viewer (or any, for the Mestre). */
export function AssignToObjective({ rollId, ownerId }: { rollId: string; ownerId: string }) {
  const state = useConflict()
  const { data: me } = useMe()
  const { send } = useRealtime()
  const [open, setOpen] = useState(false)
  const [objectiveId, setObjectiveId] = useState('')
  const [done, setDone] = useState(false)
  const [error, setError] = useState('')
  const objectives = state?.conflict?.objectives ?? []
  if (!state || objectives.length === 0 || (me?.id !== ownerId && !state.isGm)) return null
  if (done) return <p className="text-xs text-emerald-400">Atribuída ao objetivo.</p>
  if (!open) {
    return <button onClick={() => setOpen(true)} className="self-start text-xs text-emerald-400 underline">Atribuir a objetivo…</button>
  }
  const confirm = () => {
    setError('')
    send('conflict.assign', { rollId, objectiveId: objectiveId || objectives[0].id }).then(() => setDone(true), (e: Error) => setError(e.message))
  }
  return (
    <div className="flex flex-wrap items-center gap-2 text-xs">
      <select aria-label="Objetivo" value={objectiveId || objectives[0].id} onChange={(e) => setObjectiveId(e.target.value)}
        className="rounded border border-zinc-700 bg-zinc-900 px-1 py-0.5">
        {objectives.map((o) => <option key={o.id} value={o.id}>{o.title}</option>)}
      </select>
      <button onClick={confirm} className="rounded bg-emerald-600 px-2 py-0.5 text-white">Atribuir</button>
      {error && <span role="alert" className="text-red-400">{error}</span>}
    </div>
  )
}
