import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'
import { IDLE, PHASES, type RoundState } from './rounds'

const btn = 'rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700 disabled:opacity-50'

function Phases({ state }: { state: RoundState }) {
  return (
    <ol className="flex gap-2">
      {PHASES.map((p) => {
        const current = state.active && state.phase === p.id
        return (
          <li key={p.id} aria-current={current ? 'step' : undefined}
            className={`rounded px-3 py-1 text-sm ${current ? 'bg-emerald-600 font-semibold text-white' : 'bg-zinc-800 text-zinc-400'}`}>
            {p.label}
          </li>
        )
      })}
    </ol>
  )
}

function Controls({ active, run }: { active: boolean; run: (type: string) => void }) {
  if (!active) return <button className={btn} onClick={() => run('round.start')}>Iniciar</button>
  return (
    <>
      <button className={btn} onClick={() => run('round.prev')}>Voltar</button>
      <button className={btn} onClick={() => run('round.next')}>Avançar</button>
      <button className={btn} onClick={() => run('round.end')}>Encerrar</button>
    </>
  )
}

/** Round and phase bar shared live with the whole table; the Mestre also gets the controls. */
export function RoundTracker({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const { status, send, subscribe } = useRealtime()
  const [state, setState] = useState<RoundState>(IDLE)
  const [error, setError] = useState('')

  useEffect(() => subscribe('round.updated', (p) => setState((p as { state: RoundState }).state)), [subscribe])
  useEffect(() => {
    if (status !== 'open') return
    api.get<RoundState>(`/campaigns/${campaignId}/round`).then(setState, () => {})
  }, [campaignId, status])

  const run = (type: string) => {
    setError('')
    send(type).catch((e: Error) => setError(e.message))
  }

  return (
    <div className="flex flex-wrap items-center gap-4 rounded bg-zinc-900 px-4 py-3" aria-label="Controle de rodada">
      <span className="font-semibold">{state.active ? `Rodada ${state.round}` : 'Sem conflito'}</span>
      <Phases state={state} />
      {isGm && <div className="ml-auto flex gap-2"><Controls active={state.active} run={run} /></div>}
      {error && <span role="alert" className="text-sm text-red-400">{error}</span>}
    </div>
  )
}
