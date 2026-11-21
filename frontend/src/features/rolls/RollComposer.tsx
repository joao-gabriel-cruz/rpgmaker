import { Dices } from 'lucide-react'
import { useEffect, useRef, useState, type KeyboardEvent } from 'react'
import type { CharacterView } from '../characters/characters'
import { vitalsOf, penaltyLabel } from '../hotbar/vitals'
import { useRealtime } from '../realtime/realtimeContext'
import { APTITUDE_LABELS, INSTINCT_LABELS } from '../sheet-assimilacao/labels'
import type { AptitudeKey, InstinctKey } from '../sheet-assimilacao/types'
import { SURFACE } from '../table/surface'
import { ComposerGrids } from './ComposerGrids'
import { ModifierPanel } from './ModifierPanel'
import { modifiersPayload, NO_MODIFIERS, useRollState, type Modifiers } from './modifiers'
import { instinctDieKind, poolPreview, type PoolSelection } from './pool'

export interface RollOption {
  instinct: InstinctKey
  aptitude?: AptitudeKey
  label?: string
}

/** The `roll.request` body the composer sends. */
export interface RollPayload {
  characterId: string
  instinct: InstinctKey
  aptitude?: AptitudeKey
  effort: boolean
  useOccupation: boolean
  useMarkingEvent: boolean
  secondInstinct?: InstinctKey
  difficultyD12?: number
  secret: boolean
}

export interface RollComposerProps {
  campaignId: string
  character: CharacterView
  isGm: boolean
  /** Pre-selected combination (e.g. an action's roll or the last roll). */
  preset?: { instinct: InstinctKey; aptitude?: AptitudeKey }
  /** Quick alternatives shown as chips above the grids. */
  alternatives?: RollOption[]
  title?: string
  /** Called with the sent payload once the server accepted the request. */
  onRoll?: (payload: RollPayload) => void
  onClose: () => void
}

const optionLabel = (o: RollOption) =>
  o.label ?? `${INSTINCT_LABELS[o.instinct]}${o.aptitude ? ` + ${APTITUDE_LABELS[o.aptitude]}` : ''}`

function Alternatives({ options, selection, onPick }: { options: RollOption[]; selection: PoolSelection; onPick: (o: RollOption) => void }) {
  return (
    <div role="group" aria-label="Alternativas" className="flex flex-wrap gap-1.5">
      {options.map((o) => {
        const active = o.instinct === selection.instinct && o.aptitude === selection.aptitude
        return (
          <button key={optionLabel(o)} type="button" data-pick aria-pressed={active} onClick={() => onPick(o)}
            className={`rounded-full px-3 py-1 text-xs transition-[background-color,color,transform] active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-emerald-400 ${
              active ? 'bg-emerald-600 text-white' : 'bg-zinc-800 text-zinc-300 hover:bg-zinc-700'}`}>{optionLabel(o)}</button>
        )
      })}
    </div>
  )
}

// Arrow keys walk through every pick button, Tab still works as usual.
function arrowNav(e: KeyboardEvent<HTMLElement>) {
  const step = ({ ArrowRight: 1, ArrowDown: 1, ArrowLeft: -1, ArrowUp: -1 } as Record<string, number>)[e.key]
  if (!step) return
  const picks = [...e.currentTarget.querySelectorAll<HTMLButtonElement>('button[data-pick]')]
  const at = picks.indexOf(document.activeElement as HTMLButtonElement)
  if (at < 0) return
  e.preventDefault()
  picks[(at + step + picks.length) % picks.length]?.focus()
}

function useComposer(p: RollComposerProps) {
  const { send } = useRealtime()
  const [selection, setSelection] = useState<PoolSelection>({ instinct: p.preset?.instinct, aptitude: p.preset?.aptitude })
  const [mods, setMods] = useState<Modifiers>(NO_MODIFIERS)
  const [secret, setSecret] = useState(false)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const state = useRollState(p.campaignId, p.character.id)
  const needsSecond = mods.assimilated && !mods.secondInstinct
  const ready = !!selection.instinct && !needsSecond && !busy
  const roll = () => {
    if (!ready || !selection.instinct) return
    const body: RollPayload = {
      characterId: p.character.id, instinct: selection.instinct, ...(selection.aptitude ? { aptitude: selection.aptitude } : {}),
      ...modifiersPayload(mods, p.isGm), secret: p.isGm && secret,
    } as RollPayload
    setError('')
    setBusy(true)
    send('roll.request', body).then(() => { p.onRoll?.(body); p.onClose() }, (e: Error) => { setError(e.message); setBusy(false) })
  }
  const setDifficulty = (d12: number) => {
    send('roll.difficulty.set', { characterId: p.character.id, d12 }).catch((e: Error) => setError(e.message))
  }
  return { selection, setSelection, mods, setMods, secret, setSecret, error, state, needsSecond, ready, roll, setDifficulty }
}

function Footer({ c, p, preview }: { c: ReturnType<typeof useComposer>; p: RollComposerProps; preview: string }) {
  const penalty = vitalsOf(p.character).penalty
  return (
    <div className="flex flex-wrap items-center gap-3 border-t border-white/10 pt-3">
      <p aria-live="polite" data-testid="pool-preview" className="min-w-0 flex-1 text-sm font-medium text-emerald-300">
        {c.selection.instinct ? preview : 'Escolha um Instinto'}
        {penalty > 0 && <span className="ml-2 text-xs font-normal text-amber-400">Saúde: {penaltyLabel(penalty)}</span>}
        {c.needsSecond && <span className="ml-2 text-xs font-normal text-amber-400">Escolha o segundo instinto</span>}
      </p>
      <button type="button" onClick={p.onClose}
        className="rounded-lg px-3 py-1.5 text-sm text-zinc-300 transition-[background-color,transform] hover:bg-zinc-800 active:scale-[0.96]">Cancelar</button>
      <button type="button" disabled={!c.ready} onClick={c.roll}
        className="flex items-center gap-1.5 rounded-lg bg-emerald-600 px-4 py-1.5 text-sm font-semibold text-white transition-[background-color,transform] hover:bg-emerald-500 active:scale-[0.96] disabled:opacity-40 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-300">
        <Dices size={16} aria-hidden /> Rolar
      </button>
    </div>
  )
}

function Options({ c, p }: { c: ReturnType<typeof useComposer>; p: RollComposerProps }) {
  return (
    <div className="flex flex-col gap-2">
      <ModifierPanel value={c.mods} onChange={c.setMods} state={c.state} isGm={p.isGm} onDifficulty={c.setDifficulty} />
      {p.isGm && (
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={c.secret} onChange={(e) => c.setSecret(e.target.checked)} /> Rolagem secreta
        </label>
      )}
    </div>
  )
}

// Esc closes, arrows move between picks, Enter (on a pick, checkbox or the box itself) rolls.
function useKeys(c: ReturnType<typeof useComposer>, onClose: () => void) {
  return (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key === 'Escape') { e.stopPropagation(); onClose(); return }
    arrowNav(e)
    const t = e.target as HTMLElement
    const rollable = t.hasAttribute('data-pick') || t === e.currentTarget || t.getAttribute('type') === 'checkbox'
    if (e.key === 'Enter' && !e.defaultPrevented && rollable) {
      e.preventDefault()
      c.roll()
    }
  }
}

/** Roll composer popover: pick instinct/aptitude, tweak modifiers, see the pool and roll. Enter rolls, Esc closes. */
export function RollComposer(p: RollComposerProps) {
  const c = useComposer(p)
  const root = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const picks = root.current?.querySelectorAll<HTMLButtonElement>('button[data-pick]')
    ;(root.current?.querySelector<HTMLButtonElement>('button[data-pick][aria-pressed="true"]') ?? picks?.[0])?.focus()
  }, [])
  const preview = poolPreview(p.character.sheet, c.selection, c.mods, p.isGm)
  return (
    <div ref={root} role="dialog" aria-label={p.title ?? 'Compositor de rolagem'} tabIndex={-1} onKeyDown={useKeys(c, p.onClose)}
      className={`pointer-events-auto max-h-[min(36rem,calc(100dvh-14rem))] w-[min(46rem,calc(100vw-2rem))] overflow-y-auto rounded-2xl p-1 outline-none ${SURFACE}`}>
      <div className="flex flex-col gap-3 rounded-xl p-3">
        <h2 className="text-sm font-semibold">{p.title ?? `Rolar com ${p.character.name}`}</h2>
        {p.alternatives && p.alternatives.length > 0 && (
          <Alternatives options={p.alternatives} selection={c.selection} onPick={(o) => c.setSelection({ instinct: o.instinct, aptitude: o.aptitude })} />
        )}
        <div className="grid gap-4 sm:grid-cols-[1.25fr_1fr]">
          <ComposerGrids sheet={p.character.sheet} selection={c.selection} instinctDie={instinctDieKind(c.mods)} onChange={c.setSelection} />
          <Options c={c} p={p} />
        </div>
        <Footer c={c} p={p} preview={preview.text} />
        {c.error && <p role="alert" className="rounded bg-red-950 px-3 py-2 text-sm text-red-300">{c.error}</p>}
      </div>
    </div>
  )
}
