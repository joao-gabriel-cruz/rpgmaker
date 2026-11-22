import { Hammer } from 'lucide-react'
import { useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { STATE_LABEL, costAt, defOf, type Building, type Proposal } from './buildings'
import { NewBuildingDialog } from './NewBuildingDialog'
import type { Refuge } from './refuge'

function Progress({ b, cost }: { b: Building; cost: number }) {
  const pct = b.state === 'done' ? 100 : Math.min(100, (b.progress / cost) * 100)
  return (
    <div className="flex items-center gap-2 text-xs text-zinc-400">
      <div role="progressbar" aria-label="Pontos de Obra" aria-valuenow={b.progress} aria-valuemax={cost} className="h-2 flex-1 overflow-hidden rounded-full bg-zinc-700">
        <div className="h-full bg-emerald-500" style={{ width: `${pct}%` }} />
      </div>
      <span>{b.state === 'done' ? cost : b.progress}/{cost} PO</span>
    </div>
  )
}

function BuildingCard({ b, isGm, run }: { b: Building; isGm: boolean; run: (type: string, p: object) => void }) {
  const d = defOf(b.kind)
  if (!d) return null
  const btn = 'rounded bg-zinc-700 px-2 py-0.5 text-xs hover:bg-zinc-600'
  return (
    <li className="flex flex-col gap-1 rounded-lg bg-zinc-900 p-3" aria-label={d.name}>
      <div className="flex items-center gap-2">
        <span className="font-semibold text-zinc-100">{d.name}</span>
        {d.leveled && <span className="text-xs text-zinc-400">nível {b.built}{b.level > b.built ? ` → ${b.level}` : ''}</span>}
        <span className="ml-auto rounded bg-zinc-800 px-2 py-0.5 text-xs text-emerald-300">{STATE_LABEL[b.state]}</span>
      </div>
      <p className="text-sm text-zinc-400">{d.effect}</p>
      {b.state !== 'done' && <Progress b={b} cost={costAt(d, b.level)} />}
      <p className="text-xs text-zinc-400">Trabalhadores alocados: <span aria-label={`Trabalhadores em ${d.name}`}>{b.workers}</span></p>
      {isGm && (
        <div className="flex gap-2">
          {d.leveled && b.state === 'done' && <button type="button" className={btn} onClick={() => run('refuge.building.upgrade', { id: b.id })}>Evoluir</button>}
          {b.built === 0 && <button type="button" className={btn} onClick={() => run('refuge.building.cancel', { id: b.id })}>Cancelar obra</button>}
        </div>
      )}
    </li>
  )
}

function Proposals({ proposals, isGm, run }: { proposals: Proposal[]; isGm: boolean; run: (type: string, p: object) => void }) {
  if (!proposals.length) return null
  return (
    <ul aria-label="Propostas" className="flex flex-col gap-1">
      {proposals.map((p) => (
        <li key={p.id} className="flex items-center gap-2 rounded bg-zinc-900 px-3 py-2 text-sm text-zinc-300">
          <span>{p.by} propôs <strong>{defOf(p.kind)?.name ?? p.kind}</strong></span>
          {isGm && (
            <span className="ml-auto flex gap-2">
              <button type="button" aria-label={`Aprovar ${defOf(p.kind)?.name}`} className="rounded bg-emerald-700 px-2 py-0.5 text-xs" onClick={() => run('refuge.building.approve', { id: p.id })}>Aprovar</button>
              <button type="button" aria-label={`Recusar ${defOf(p.kind)?.name}`} className="rounded bg-zinc-700 px-2 py-0.5 text-xs" onClick={() => run('refuge.building.reject', { id: p.id })}>Recusar</button>
            </span>
          )}
        </li>
      ))}
    </ul>
  )
}

export function BuildingsSection({ refuge, isGm }: { refuge: Refuge; isGm: boolean }) {
  const { send } = useRealtime()
  const [open, setOpen] = useState(false)
  const [error, setError] = useState('')
  const run = (type: string, p: object) => { setError(''); send(type, p).catch((e: Error) => setError(e.message)) }
  return (
    <section className="flex max-w-3xl flex-col gap-2">
      <div className="flex items-center gap-3">
        <h3 className="text-sm text-zinc-400">Construções</h3>
        <button type="button" onClick={() => setOpen(true)} className="flex items-center gap-1 rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700">
          <Hammer size={14} />Nova construção</button>
      </div>
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      {refuge.buildings.length === 0 && <p className="text-sm text-zinc-500">Nenhuma construção ainda.</p>}
      <ul className="grid grid-cols-1 gap-2">
        {refuge.buildings.map((b) => <BuildingCard key={b.id} b={b} isGm={isGm} run={run} />)}
      </ul>
      <Proposals proposals={refuge.proposals} isGm={isGm} run={run} />
      {open && <NewBuildingDialog isGm={isGm} buildings={refuge.buildings} proposals={refuge.proposals}
        onPick={(kind) => run(isGm ? 'refuge.building.add' : 'refuge.building.propose', { kind })} onClose={() => setOpen(false)} />}
    </section>
  )
}
