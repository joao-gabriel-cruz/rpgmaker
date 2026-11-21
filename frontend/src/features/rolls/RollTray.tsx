import { X } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useMe } from '../auth/auth'
import { AssignToObjective } from '../conflict/AssignToObjective'
import { DiceTray } from '../dice3d/DiceTray'
import type { DieKind, DieResult } from '../dice3d/faces'
import { useRealtime } from '../realtime/realtimeContext'
import { modifiersText } from './modifiers'
import { ResultBanner } from './ResultBanner'
import type { RollEntry } from './rolls'
import { useRollQueue, type ActiveRoll } from './useRollQueue'

/** Seconds the result stays up for people who did not roll. */
export const AUTO_CLOSE_MS = 4000

const toDice = (e: RollEntry): DieResult[] =>
  e.dice.map((d) => ({ kind: d.kind as DieKind, face: d.face, variant: d.difficulty ? 'difficulty' : 'normal' }))

function Picker({ entry, selected, onConfirm, busy }: { entry: RollEntry; selected: number[]; onConfirm: () => void; busy: boolean }) {
  const need = entry.keptCount ?? 1
  return (
    <div className="flex items-center justify-center gap-4 text-base">
      <span>Escolha {need} {need === 1 ? 'dado' : 'dados'} ({selected.length}/{need})</span>
      <button onClick={onConfirm} disabled={selected.length !== need || busy}
        className="rounded-lg bg-emerald-600 px-4 py-1.5 font-semibold text-white transition-[background-color,transform] hover:bg-emerald-500 active:scale-[0.96] disabled:opacity-40 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-300">Confirmar</button>
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

const TRAY_BOX = 'h-[min(48dvh,26rem)] w-full overflow-hidden rounded-xl bg-zinc-950/60'

function Body({ active, mine }: { active: ActiveRoll; mine: boolean }) {
  const { entry, result } = active
  const { send } = useRealtime()
  const [settled, setSettled] = useState(false)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const { selected, toggle } = useSelection(entry)
  const dice = useMemo(() => toDice(entry), [entry])
  const mods = modifiersText(entry.request)
  const confirm = () => {
    setBusy(true)
    send('roll.keep', { rollId: entry.rollId, kept: selected }).catch((e: Error) => setError(e.message)).finally(() => setBusy(false))
  }
  return (
    <>
      <DiceTray dice={dice} rollId={entry.rollId} seed={entry.seed} selectable={mine && !result} selected={result?.kept ?? selected}
        onToggle={toggle} onSettled={() => setSettled(true)} className={TRAY_BOX} />
      {mods && <p className="text-center text-xs text-amber-300">{mods}</p>}
      {result && <ResultBanner result={result} />}
      {result && entry.kind === 'assimilacao' && <AssignToObjective rollId={entry.rollId} ownerId={entry.userId} />}
      {!result && mine && settled && <Picker entry={entry} selected={selected} onConfirm={confirm} busy={busy} />}
      {!result && mine && !settled && <p className="text-center text-sm text-zinc-400">Rolando…</p>}
      {!result && !mine && <p className="text-center text-sm text-zinc-400">{entry.userName} está escolhendo…</p>}
      {error && <p role="alert" className="text-center text-sm text-red-400">{error}</p>}
    </>
  )
}

// Others' rolls close by themselves; the roller closes with the button or Esc once there is a result.
function useDismiss(active: ActiveRoll | undefined, mine: boolean, close: () => void) {
  const rollId = active?.entry.rollId
  const done = !!active?.result
  useEffect(() => {
    if (!rollId || !done || mine) return
    const t = setTimeout(close, AUTO_CLOSE_MS)
    return () => clearTimeout(t)
  }, [rollId, done, mine, close])
  useEffect(() => {
    if (!rollId || (mine && !done)) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') close() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [rollId, done, mine, close])
}

/** Translucent overlay (slot: overlay) with the dice of the roll in progress; extra rolls wait in a queue. */
export function RollTray() {
  const { current, waiting, close } = useRollQueue()
  const { data: me } = useMe()
  const mine = !!current && me?.id === current.entry.userId
  useDismiss(current, mine, close)
  if (!current) return null
  const { entry, result } = current
  return (
    <aside aria-label="Bandeja de dados" className="pointer-events-auto flex w-[min(52rem,calc(100%-2rem))] flex-col gap-3 rounded-3xl bg-zinc-900/75 p-2 shadow-2xl ring-1 ring-white/10 backdrop-blur-xl motion-safe:animate-[tray-in_160ms_ease-out]">
      <div className="flex flex-col gap-3 rounded-2xl p-3">
        <header className="flex items-center justify-between gap-3">
          <h2 className="text-base font-semibold">
            {entry.userName}{entry.characterName ? ` · ${entry.characterName}` : ''}{entry.visibility === 'gm' ? ' · Secreta' : ''}
          </h2>
          <span className="flex items-center gap-2">
            {waiting > 0 && <span className="rounded-full bg-zinc-800 px-2 py-0.5 text-xs text-zinc-300">+{waiting} na fila</span>}
            {(!mine || result) && (
              <button type="button" aria-label="Fechar bandeja" onClick={close}
                className="rounded-lg p-1.5 text-zinc-400 transition-[background-color,color,transform] hover:bg-zinc-800 hover:text-white active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-emerald-400">
                <X size={18} aria-hidden />
              </button>
            )}
          </span>
        </header>
        <Body key={entry.rollId} active={current} mine={mine} />
      </div>
    </aside>
  )
}
