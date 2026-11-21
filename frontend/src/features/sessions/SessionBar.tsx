import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'

interface SessionState {
  active: boolean
  startedAt: string | null
}

const pill = 'flex h-8 items-center gap-2 rounded-lg px-2.5 text-sm'

function Indicator({ active }: { active: boolean }) {
  return (
    <>
      <span aria-hidden className={`h-2 w-2 rounded-full ${active ? 'bg-emerald-400' : 'bg-zinc-500'}`} />
      <span className={active ? 'font-medium text-emerald-300' : 'text-zinc-400'}>{active ? 'Sessão ativa' : 'Sem sessão ativa'}</span>
    </>
  )
}

/** Compact session indicator for everyone; the Mestre clicks it to start or end the session. */
export function SessionBar({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const { status, send, subscribe } = useRealtime()
  const [state, setState] = useState<SessionState>({ active: false, startedAt: null })
  const [error, setError] = useState('')
  useEffect(() => subscribe('session.updated', (p) => setState(p as SessionState)), [subscribe])
  useEffect(() => {
    if (status !== 'open') return
    api.get<SessionState>(`/campaigns/${campaignId}/session`).then(setState, () => {})
  }, [campaignId, status])
  const toggle = () => {
    setError('')
    send(state.active ? 'session.end' : 'session.start').catch((e: Error) => setError(e.message))
  }
  return (
    <div className="flex items-center gap-2">
      {isGm ? (
        <button type="button" aria-label={state.active ? 'Encerrar sessão' : 'Iniciar sessão'} title={state.active ? 'Encerrar sessão' : 'Iniciar sessão'}
          onClick={toggle} className={`${pill} transition-[background-color,transform] hover:bg-zinc-800 active:scale-[0.96]`}>
          <Indicator active={state.active} />
        </button>
      ) : (
        <span aria-label="Estado da sessão" className={pill}><Indicator active={state.active} /></span>
      )}
      {error && <p role="alert" className="text-xs text-red-400">{error}</p>}
    </div>
  )
}
