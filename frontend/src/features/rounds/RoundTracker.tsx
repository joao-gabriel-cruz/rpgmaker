import { ChevronLeft, ChevronRight, Play, Square } from 'lucide-react'
import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'
import { IDLE, PHASES, type RoundState } from './rounds'

const iconBtn = 'flex h-7 w-7 items-center justify-center rounded-lg bg-zinc-800 text-zinc-200 transition-[background-color,transform] hover:bg-zinc-700 active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-400 disabled:opacity-40'

function Phases({ state }: { state: RoundState }) {
  return (
    <ol className="hidden items-center gap-1 lg:flex">
      {PHASES.map((p) => {
        const current = state.active && state.phase === p.id
        return (
          <li key={p.id} aria-current={current ? 'step' : undefined}
            className={`whitespace-nowrap rounded-md px-2 py-0.5 text-xs ${current ? 'bg-emerald-600 font-semibold text-white' : 'bg-zinc-800/80 text-zinc-400 compact:hidden'}`}>
            {p.label}
          </li>
        )
      })}
    </ol>
  )
}

function Controls({ active, run }: { active: boolean; run: (type: string) => void }) {
  if (!active) {
    return <button type="button" title="Iniciar rodada" className={`${iconBtn} w-auto shrink-0 gap-1 px-2 text-xs`} onClick={() => run('round.start')}><Play size={14} aria-hidden /><span className="compact:sr-only">Iniciar</span></button>
  }
  return (
    <>
      <button type="button" className={iconBtn} aria-label="Voltar" title="Fase anterior" onClick={() => run('round.prev')}><ChevronLeft size={16} aria-hidden /></button>
      <button type="button" className={iconBtn} aria-label="Avançar" title="Próxima fase" onClick={() => run('round.next')}><ChevronRight size={16} aria-hidden /></button>
      <button type="button" className={iconBtn} aria-label="Encerrar" title="Encerrar a rodada" onClick={() => run('round.end')}><Square size={14} aria-hidden /></button>
    </>
  )
}

/** Compact round and phase bar shared live with the table; the Mestre also gets ◀ ▶ and start/end. */
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
    <div className="flex shrink-0 items-center gap-2" aria-label="Controle de rodada">
      <span className="whitespace-nowrap text-sm font-semibold">{state.active ? `Rodada ${state.round}` : 'Sem conflito'}</span>
      <Phases state={state} />
      {isGm && <div className="flex items-center gap-1"><Controls active={state.active} run={run} /></div>}
      {error && <span role="alert" className="text-xs text-red-400">{error}</span>}
    </div>
  )
}
