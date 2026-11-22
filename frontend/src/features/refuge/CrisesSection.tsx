import { useState } from 'react'
import { CrisisCard } from './CrisisCard'
import { KIND_LABEL, KINDS, useRefugeCommand, type CrisisKind } from './crises'
import type { Refuge } from './refuge'

const field = 'rounded bg-zinc-800 px-2 py-1 text-sm'

function NewCrisis({ run }: { run: (type: string, payload?: object) => void }) {
  const [name, setName] = useState('')
  const [kind, setKind] = useState<CrisisKind>('custom')
  const [grade, setGrade] = useState(1)
  return (
    <form className="flex flex-wrap items-end gap-2" onSubmit={(e) => {
      e.preventDefault()
      if (!name.trim()) return
      run('refuge.crisis.create', { name: name.trim(), kind, grade, description: '' })
      setName('')
    }}>
      <input aria-label="Nome da crise" placeholder="Nova crise" value={name} onChange={(e) => setName(e.target.value)} className={field} />
      <select aria-label="Tipo da crise" value={kind} onChange={(e) => setKind(e.target.value as CrisisKind)} className={field}>
        {KINDS.map((k) => <option key={k} value={k}>{KIND_LABEL[k]}</option>)}
      </select>
      <input aria-label="Gravidade da nova crise" type="number" min={0} max={10} value={grade} onChange={(e) => setGrade(Number(e.target.value))} className={`${field} w-16`} />
      <button className="rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500">Criar crise</button>
      <button type="button" className="rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700" onClick={() => run('refuge.crisis.roll', {})}>Rolar gravidade (nova crise)</button>
    </form>
  )
}

/** Crises of the Refúgio: status cards, the Mestre's severity roll and the Moral penalty option. */
export function CrisesSection({ refuge, isGm }: { refuge: Refuge; isGm: boolean }) {
  const { run, error } = useRefugeCommand()
  const crises = (refuge.crises ?? []).filter((c) => isGm || c.status !== 'suggested')
  const hasMedic = (refuge.residents ?? []).some((r) => r.role === 'medic')
  const penalty = refuge.crisisSettings?.moralePenalty ?? true
  return (
    <section aria-label="Crises" className="flex max-w-3xl flex-col gap-2">
      <h3 className="text-sm text-zinc-400">Crises ({crises.length})</h3>
      <ul className="flex flex-col gap-2">
        {crises.map((c) => <CrisisCard key={c.id} crisis={c} isGm={isGm} hasMedic={hasMedic} run={run} />)}
        {crises.length === 0 && <li className="text-sm text-zinc-400">Nenhuma crise. O Refúgio respira.</li>}
      </ul>
      {isGm && <NewCrisis run={run} />}
      {isGm && (
        <label className="flex items-center gap-2 text-xs text-zinc-400">
          <input type="checkbox" checked={penalty} onChange={(e) => run('refuge.crisis.settings', { moralePenalty: e.target.checked })} />
          Cada crise confirmada tira 1 de Moral
        </label>
      )}
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
    </section>
  )
}
