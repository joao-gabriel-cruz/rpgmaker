import { Dices, Eraser } from 'lucide-react'
import { useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'

const KINDS = ['d6', 'd10', 'd12'] as const
type Kind = (typeof KINDS)[number]
type Pool = Record<Kind, number>

const EMPTY: Pool = { d6: 0, d10: 0, d12: 0 }
const MAX_PER_KIND = 20
const BTN = 'rounded-lg px-2.5 py-1.5 text-sm transition-[background-color,transform] active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-emerald-400'

const poolText = (p: Pool) => KINDS.filter((k) => p[k] > 0).map((k) => `${p[k]}${k}`).join(' + ')

/** Mestre's dice pool assembled with buttons (+d6 +d10 +d12), rolled through `roll.pool` without a character. */
export function PoolBuilder() {
  const { send } = useRealtime()
  const [pool, setPool] = useState<Pool>(EMPTY)
  const [error, setError] = useState('')
  const empty = poolText(pool) === ''
  const roll = () => {
    setError('')
    send('roll.pool', pool).then(() => setPool(EMPTY), (e: Error) => setError(e.message))
  }
  return (
    <section aria-label="Montador de parada" className="flex flex-col gap-2 rounded-xl bg-zinc-800/50 p-2">
      <h3 className="text-xs font-medium uppercase tracking-wide text-zinc-400">Parada do Mestre</h3>
      <div className="flex gap-1.5">
        {KINDS.map((k) => (
          <button key={k} type="button" aria-label={`Adicionar ${k}`} disabled={pool[k] >= MAX_PER_KIND}
            onClick={() => setPool((p) => ({ ...p, [k]: p[k] + 1 }))} className={`${BTN} bg-zinc-700 text-zinc-100 hover:bg-zinc-600`}>+{k}</button>
        ))}
      </div>
      <p aria-live="polite" data-testid="pool-builder-text" className="min-h-5 text-sm text-emerald-300">{empty ? 'Nenhum dado ainda' : poolText(pool)}</p>
      <div className="flex gap-1.5">
        <button type="button" disabled={empty} onClick={() => setPool(EMPTY)} className={`${BTN} flex items-center gap-1 bg-zinc-800 text-zinc-300 hover:bg-zinc-700 disabled:opacity-40`}>
          <Eraser size={14} aria-hidden /> Limpar</button>
        <button type="button" disabled={empty} onClick={roll} className={`${BTN} flex items-center gap-1 bg-emerald-600 font-semibold text-white hover:bg-emerald-500 disabled:opacity-40`}>
          <Dices size={14} aria-hidden /> Rolar parada</button>
      </div>
      {error && <p role="alert" className="text-xs text-red-400">{error}</p>}
    </section>
  )
}
