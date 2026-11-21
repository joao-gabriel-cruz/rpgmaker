import { useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { SYMBOL_LABELS, type Conflict, type Threat, type Triggered } from './conflict'
import { useConflict } from './ConflictContext'
import { ConflictEditor } from './ConflictEditor'
import { ProgressBar } from './ProgressBar'

const btn = 'rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700'

function ThreatRow({ t, isGm, acting }: { t: Threat; isGm: boolean; acting: boolean }) {
  const { send } = useRealtime()
  const [error, setError] = useState('')
  const roll = () => send('conflict.threat.roll', { threatId: t.id }).then(() => setError(''), (e: Error) => setError(e.message))
  const pool = t.pool ? `${t.pool.d6}d6 + ${t.pool.d10}d10 + ${t.pool.d12}d12` : ''
  return (
    <li className={`flex flex-wrap items-center gap-2 rounded px-3 py-1 ${acting ? 'bg-red-900/60 ring-1 ring-red-400' : 'bg-zinc-900'}`}>
      <span className="font-medium">{t.name}</span>
      {isGm && <span className="text-xs text-zinc-400">{pool}</span>}
      {isGm && <button className={`${btn} ml-auto`} onClick={roll} aria-label={`Rolar ${t.name}`}>Rolar</button>}
      {error && <span role="alert" className="text-xs text-red-400">{error}</span>}
    </li>
  )
}

function Activations({ c, triggered }: { c: Conflict; triggered: Triggered[] }) {
  return (
    <div className="flex flex-col gap-1">
      <h4 className="text-sm font-semibold">Ativações (só Mestre)</h4>
      <ul className="flex flex-col gap-1 text-sm">
        {(c.activations ?? []).map((a) => (
          <li key={a.id}>
            A cada {a.threshold} {SYMBOL_LABELS[a.symbol]} ({a.source === 'players' ? 'jogadores' : 'ameaça'}): {a.text}
            <span className="text-zinc-400"> — {a.count ?? 0}/{a.threshold}</span>
          </li>
        ))}
      </ul>
      {triggered.map((t, i) => (
        <p key={i} role="status" className="rounded bg-amber-900/60 px-3 py-1 text-sm text-amber-100">
          Ativação disparada{t.times > 1 ? ` (x${t.times})` : ''}: {t.text || `${t.threshold} ${SYMBOL_LABELS[t.symbol]}`}
        </p>
      ))}
    </div>
  )
}

function Body({ c, isGm, acting, triggered }: { c: Conflict; isGm: boolean; acting: string | null; triggered: Triggered[] }) {
  return (
    <>
      {c.conditions.length > 0 && (
        <ul className="list-disc pl-5 text-sm text-zinc-300">{c.conditions.map((x, i) => <li key={i}>{x}</li>)}</ul>
      )}
      <ul className="flex flex-col gap-2">
        {c.objectives.map((o) => (
          <li key={o.id} className="flex flex-col gap-1 rounded bg-zinc-900 px-3 py-2">
            <span className="text-sm font-medium">
              {o.title} <span className="text-xs text-zinc-400">
                ({o.kind === 'main' ? 'principal' : 'secundário'}, {o.scope === 'collective' ? 'coletivo' : 'individual'})
              </span>
            </span>
            <ProgressBar objective={o} />
          </li>
        ))}
      </ul>
      {c.threats.length > 0 && (
        <ul className="flex flex-col gap-1" aria-label="Ameaças">
          {c.threats.map((t) => <ThreatRow key={t.id} t={t} isGm={isGm} acting={!!acting && t.npcCharacterId === acting} />)}
        </ul>
      )}
      {isGm && <Activations c={c} triggered={triggered} />}
    </>
  )
}

/** Partitura de Conflito: everyone sees objectives and progress; the Mestre edits and rolls threats. */
export function ConflictPanel() {
  const state = useConflict()
  const { send } = useRealtime()
  const [editing, setEditing] = useState(false)
  if (!state) return null
  const { conflict, isGm, acting, triggered } = state
  if (!conflict && !isGm) return null
  return (
    <section aria-label="Partitura de Conflito" className="flex flex-col gap-3">
      <div className="flex items-center gap-3">
        <h2 className="text-xl font-semibold">Partitura de Conflito{conflict ? `: ${conflict.title}` : ''}</h2>
        {isGm && (
          <div className="ml-auto flex gap-2">
            <button className={btn} onClick={() => setEditing((e) => !e)}>{editing ? 'Fechar editor' : conflict ? 'Editar' : 'Criar partitura'}</button>
            {conflict && <button className={btn} onClick={() => send('conflict.close').catch(() => {})}>Encerrar</button>}
          </div>
        )}
      </div>
      {isGm && editing && <ConflictEditor campaignId={state.campaignId} key={conflict?.id ?? 'new'} initial={conflict} onDone={() => setEditing(false)} />}
      {conflict && <Body c={conflict} isGm={isGm} acting={acting} triggered={triggered} />}
    </section>
  )
}
