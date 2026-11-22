import { KIND_LABEL, STATUS_LABEL, STATUS_TONE, type Crisis, type CrisisStatus } from './crises'

type Run = (type: string, payload?: object) => void

const btn = 'rounded bg-zinc-800 px-2 py-1 text-xs hover:bg-zinc-700'

function Dice({ crisis, isGm, run }: { crisis: Crisis; isGm: boolean; run: Run }) {
  const roll = crisis.roll
  if (!roll) return null
  const waiting = roll.discarded === null
  return (
    <div className="flex flex-col gap-1">
      <div className="flex flex-wrap items-center gap-1" aria-label="Dados da gravidade">
        {roll.faces.map((face, i) => (
          <button key={i} disabled={!waiting || !isGm} aria-label={`Descartar dado ${i + 1} (face ${face})`}
            onClick={() => run('refuge.crisis.discard', { crisisId: crisis.id, index: i })}
            className={`h-8 w-8 rounded border text-sm font-bold ${roll.discarded === i ? 'border-zinc-700 text-zinc-600 line-through' : 'border-emerald-600 text-emerald-300'} ${waiting && isGm ? 'hover:bg-emerald-900' : ''}`}>
            {face}
          </button>
        ))}
        {roll.symbols !== null && <span className="text-xs text-zinc-300">{roll.symbols} símbolo(s) nos 3 dados mantidos</span>}
      </div>
      {waiting && <p className="text-xs text-amber-300">{isGm ? 'Escolha 1 dado para descartar.' : 'O Mestre vai descartar 1 dado.'}</p>}
    </div>
  )
}

function Suggestions({ crisis, run }: { crisis: Crisis; run: Run }) {
  const pending = crisis.suggestions.filter((s) => s.status === 'pending')
  return (
    <ul className="flex flex-col gap-1">
      {pending.map((s) => (
        <li key={s.id} className="flex flex-wrap items-center gap-2 rounded bg-zinc-800 px-2 py-1 text-xs">
          <span className="flex-1">Sugestão: {s.text}</span>
          <button className="rounded bg-emerald-600 px-2 py-1 font-medium hover:bg-emerald-500" aria-label={`Aplicar sugestão de ${crisis.name}`}
            onClick={() => run('refuge.crisis.apply', { crisisId: crisis.id, suggestionId: s.id })}>Aplicar</button>
          <button className={btn} onClick={() => run('refuge.crisis.dismiss', { crisisId: crisis.id, suggestionId: s.id })}>Ignorar</button>
        </li>
      ))}
    </ul>
  )
}

function GmActions({ crisis, run }: { crisis: Crisis; run: Run }) {
  const set = (status: CrisisStatus) => run('refuge.crisis.update', { id: crisis.id, status })
  const remove = () => run('refuge.crisis.delete', { id: crisis.id })
  if (crisis.status === 'suggested') {
    return (
      <div className="flex gap-2">
        <button className="rounded bg-emerald-600 px-2 py-1 text-xs font-medium hover:bg-emerald-500" onClick={() => run('refuge.crisis.confirm', { id: crisis.id })}>Confirmar crise</button>
        <button className={btn} onClick={remove}>Descartar</button>
      </div>
    )
  }
  return (
    <div className="flex flex-wrap gap-2">
      <button className={btn} onClick={() => run('refuge.crisis.roll', { crisisId: crisis.id })}>Rolar gravidade</button>
      {crisis.status !== 'contained' && <button className={btn} onClick={() => set('contained')}>Conter</button>}
      {crisis.status !== 'resolved' && <button className={btn} onClick={() => set('resolved')}>Resolver</button>}
      {crisis.status !== 'active' && <button className={btn} onClick={() => set('active')}>Reativar</button>}
      <button className={btn} onClick={remove}>Remover</button>
    </div>
  )
}

export function CrisisCard({ crisis, isGm, hasMedic, run }: { crisis: Crisis; isGm: boolean; hasMedic: boolean; run: Run }) {
  return (
    <li data-testid="crisis-card" data-status={crisis.status} className="flex flex-col gap-2 rounded bg-zinc-900 px-3 py-2">
      <div className="flex flex-wrap items-center gap-2">
        <span className="font-medium">{crisis.name}</span>
        <span className="text-xs text-zinc-400">{KIND_LABEL[crisis.kind]} · semana {crisis.week}</span>
        <span aria-label={`Gravidade de ${crisis.name}`} className="rounded bg-zinc-800 px-2 text-xs">Gravidade {crisis.grade}</span>
        <span className={`ml-auto rounded border px-2 text-xs ${STATUS_TONE[crisis.status]}`}>{STATUS_LABEL[crisis.status]}</span>
      </div>
      {crisis.description && <p className="whitespace-pre-line text-sm text-zinc-300">{crisis.description}</p>}
      {crisis.kind === 'disease' && crisis.status === 'active' && (
        <p className="text-xs text-lime-300">−1 dado em todos os testes até a cura.{hasMedic ? ' Um morador médico dá +1 dado para tratar a doença.' : ''}</p>
      )}
      <Dice crisis={crisis} isGm={isGm} run={run} />
      {isGm && <Suggestions crisis={crisis} run={run} />}
      {isGm && <GmActions crisis={crisis} run={run} />}
    </li>
  )
}
