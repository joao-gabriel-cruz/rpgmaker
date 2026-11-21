import { useState } from 'react'
import { ActionBar } from '../actions/ActionBar'
import { useRealtime } from '../realtime/realtimeContext'
import { PoolBuilder } from './PoolBuilder'

const GENERIC = ['d6', '2d6', 'd10', 'd12', 'd20'] as const

function GenericShortcuts() {
  const { send } = useRealtime()
  const [error, setError] = useState('')
  const roll = (expression: string) => {
    setError('')
    send('roll.generic', { expression }).catch((e: Error) => setError(e.message))
  }
  return (
    <section aria-label="Atalhos genéricos" className="flex flex-col gap-1.5">
      <h3 className="text-xs font-medium uppercase tracking-wide text-zinc-400">Dados soltos</h3>
      <div className="flex flex-wrap gap-1.5">
        {GENERIC.map((x) => (
          <button key={x} type="button" aria-label={`Rolar ${x}`} onClick={() => roll(x)}
            className="rounded-lg bg-zinc-800 px-3 py-1.5 text-sm font-medium tabular-nums text-zinc-100 transition-[background-color,transform] hover:bg-zinc-700 active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-emerald-400">{x}</button>
        ))}
      </div>
      {error && <p role="alert" className="text-xs text-red-400">{error}</p>}
    </section>
  )
}

/** Rolls tab controls without a text box: the actions for everyone; pool builder and loose dice for the Mestre. */
export function QuickRolls({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  return (
    <div className="flex flex-col gap-3">
      <section aria-label="Ações" className="flex flex-col gap-1.5">
        <h3 className="text-xs font-medium uppercase tracking-wide text-zinc-400">Ações</h3>
        <ActionBar campaignId={campaignId} isGm={isGm} layout="list" />
      </section>
      {isGm && <PoolBuilder />}
      {isGm && <GenericShortcuts />}
    </div>
  )
}
