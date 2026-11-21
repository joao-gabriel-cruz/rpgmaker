import { useEffect, useMemo, useState } from 'react'
import { useMe } from '../auth/auth'
import { DiceTray } from '../dice3d/DiceTray'
import type { DieKind, DieResult } from '../dice3d/faces'
import { useRealtime } from '../realtime/realtimeContext'
import type { FinalizedEvent, RollEntry, RollResult } from './rolls'
import { SymbolCounts } from './SymbolCounts'

const toDice = (e: RollEntry): DieResult[] =>
  e.dice.map((d) => ({ kind: d.kind as DieKind, face: d.face, variant: d.difficulty ? 'difficulty' : 'normal' }))

interface Active {
  entry: RollEntry
  result?: RollResult
}

function useActiveRoll() {
  const { subscribe } = useRealtime()
  const [active, setActive] = useState<Active | null>(null)
  useEffect(() => subscribe('roll.rolled', (p) => {
    const entry = p as RollEntry
    if (entry.kind === 'assimilacao') setActive({ entry })
  }), [subscribe])
  useEffect(() => subscribe('roll.finalized', (p) => {
    const ev = p as FinalizedEvent
    setActive((a) => (a && a.entry.rollId === ev.rollId ? { ...a, result: ev.result } : a))
  }), [subscribe])
  return { active, close: () => setActive(null) }
}

function Picker({ entry, selected, onConfirm, busy }: { entry: RollEntry; selected: number[]; onConfirm: () => void; busy: boolean }) {
  const need = entry.keptCount ?? 1
  return (
    <div className="flex items-center justify-between gap-2 text-sm">
      <span>Escolha {need} {need === 1 ? 'dado' : 'dados'} ({selected.length}/{need})</span>
      <button onClick={onConfirm} disabled={selected.length !== need || busy}
        className="rounded bg-emerald-600 px-3 py-1 font-medium text-white disabled:opacity-40">Confirmar</button>
    </div>
  )
}

function useSelection(entry: RollEntry) {
  const [selected, setSelected] = useState<number[]>(entry.suggested ?? [])
  const toggle = (i: number) => {
    if (entry.dice[i]?.difficulty) return
    setSelected((s) => (s.includes(i) ? s.filter((x) => x !== i) : [...s, i]))
  }
  return { selected, toggle }
}

function Body({ active, onClose }: { active: Active; onClose: () => void }) {
  const { entry, result } = active
  const { data: me } = useMe()
  const { send } = useRealtime()
  const [settled, setSettled] = useState(false)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const { selected, toggle } = useSelection(entry)
  const dice = useMemo(() => toDice(entry), [entry])
  const mine = me?.id === entry.userId
  const confirm = () => {
    setBusy(true)
    send('roll.keep', { rollId: entry.rollId, kept: selected }).catch((e: Error) => setError(e.message)).finally(() => setBusy(false))
  }
  return (
    <>
      <DiceTray dice={dice} rollId={entry.rollId} seed={entry.seed} selectable={mine && !result} selected={result?.kept ?? selected}
        onToggle={toggle} onSettled={() => setSettled(true)} />
      {result && <SymbolCounts result={result} />}
      {!result && mine && settled && <Picker entry={entry} selected={selected} onConfirm={confirm} busy={busy} />}
      {!result && mine && !settled && <p className="text-sm text-zinc-400">Rolando…</p>}
      {!result && !mine && <p className="text-sm text-zinc-400">{entry.userName} está escolhendo…</p>}
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      {result && <button onClick={onClose} className="self-end text-sm text-zinc-400 underline">Fechar</button>}
    </>
  )
}

/** Fixed panel showing the dice of the roll in progress, for everyone who received it. */
export function RollTray() {
  const { active, close } = useActiveRoll()
  if (!active) return null
  const { entry } = active
  return (
    <aside aria-label="Bandeja de dados" className="fixed bottom-4 right-4 z-50 flex w-[26rem] max-w-[calc(100vw-2rem)] flex-col gap-2 rounded-lg border border-zinc-700 bg-zinc-900/95 p-3 shadow-xl">
      <h2 className="text-sm font-semibold">
        {entry.userName}{entry.characterName ? ` · ${entry.characterName}` : ''}{entry.visibility === 'gm' ? ' · Secreta' : ''}
      </h2>
      <Body key={entry.rollId} active={active} onClose={close} />
    </aside>
  )
}
