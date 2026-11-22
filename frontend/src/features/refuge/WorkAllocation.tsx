import { useEffect, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { allocationOf, defOf, totalOf, type Building } from './buildings'
import type { Refuge } from './refuge'
import { Stepper } from './Stepper'

function Row({ b, value, room, isGm, onChange }: { b: Building; value: number; room: number; isGm: boolean; onChange: (n: number) => void }) {
  const name = defOf(b.kind)?.name ?? b.kind
  return (
    <div className="flex items-center justify-between rounded bg-zinc-900 px-3 py-2 text-sm text-zinc-300">
      {name}{b.kind === 'fortification' && <span className="text-xs text-zinc-500">(guarda)</span>}
      <Stepper label={`Alocar em ${name}`} value={value} max={value + room} editable={isGm} onChange={onChange} />
    </div>
  )
}

/** Distributes the available workers among works, productive buildings and guard posts. */
export function WorkAllocation({ refuge, isGm }: { refuge: Refuge; isGm: boolean }) {
  const { send } = useRealtime()
  const saved = allocationOf(refuge.buildings)
  const [draft, setDraft] = useState(saved)
  const [error, setError] = useState('')
  const key = JSON.stringify(saved)
  useEffect(() => setDraft(JSON.parse(key)), [key])
  if (!refuge.buildings.length) return null
  const available = refuge.derived.workers
  const used = totalOf(draft)
  const dirty = JSON.stringify(draft) !== key
  const save = () => { setError(''); send('refuge.work.allocate', { allocation: draft }).catch((e: Error) => setError(e.message)) }
  return (
    <section className="flex max-w-md flex-col gap-2">
      <h3 className="text-sm text-zinc-400">Alocação de trabalho</h3>
      <p className="text-sm text-zinc-300">Alocados <span aria-label="Trabalhadores alocados" className="font-semibold text-emerald-400">{used}</span> de {available} trabalhadores</p>
      {refuge.buildings.map((b) => (
        <Row key={b.id} b={b} value={draft[b.id] ?? 0} room={Math.max(0, available - used)} isGm={isGm} onChange={(n) => setDraft({ ...draft, [b.id]: n })} />
      ))}
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      {isGm && <button type="button" disabled={!dirty} onClick={save} className="self-start rounded bg-emerald-600 px-3 py-1 text-sm disabled:opacity-50">Salvar alocação</button>}
    </section>
  )
}
