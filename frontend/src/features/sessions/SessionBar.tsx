import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'

interface SessionState {
  active: boolean
  startedAt: string | null
}

const btn = 'rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700'

/** Active-session indicator for everyone; the Mestre also gets start/end buttons. */
export function SessionBar({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const { status, send, subscribe } = useRealtime()
  const [state, setState] = useState<SessionState>({ active: false, startedAt: null })
  const [error, setError] = useState('')
  useEffect(() => subscribe('session.updated', (p) => setState(p as SessionState)), [subscribe])
  useEffect(() => {
    if (status !== 'open') return
    api.get<SessionState>(`/campaigns/${campaignId}/session`).then(setState, () => {})
  }, [campaignId, status])
  const run = (type: string) => {
    setError('')
    send(type).catch((e: Error) => setError(e.message))
  }
  return (
    <div className="flex flex-wrap items-center gap-3">
      <span aria-label="Estado da sessão" className={`rounded px-3 py-1 text-sm ${state.active ? 'bg-emerald-600 font-semibold text-white' : 'bg-zinc-800 text-zinc-400'}`}>
        {state.active ? 'Sessão ativa' : 'Sem sessão ativa'}
      </span>
      {isGm && !state.active && <button className={btn} onClick={() => run('session.start')}>Iniciar sessão</button>}
      {isGm && state.active && <button className={btn} onClick={() => run('session.end')}>Encerrar sessão</button>}
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
    </div>
  )
}
