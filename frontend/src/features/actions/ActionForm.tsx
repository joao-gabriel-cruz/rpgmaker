import { useState } from 'react'
import { APTITUDE_LABELS, INSTINCT_LABELS } from '../sheet-assimilacao/labels'
import { INSTINCT_KEYS, KNOWLEDGE_KEYS, PRACTICE_KEYS, type AptitudeKey, type InstinctKey } from '../sheet-assimilacao/types'
import {
  ACTION_ICONS, ANIMATIONS, ANIMATION_LABELS, CATEGORY_LABELS, OUTCOME_LABELS, TARGET_LABELS,
  type ActionCategory, type ActionOutcome, type ActionRoll, type ActionTargetKind, type Animation, type GameAction,
} from './actions'

const field = 'rounded bg-zinc-800 px-2 py-1 text-sm'
const APTITUDES: AptitudeKey[] = [...KNOWLEDGE_KEYS, ...PRACTICE_KEYS]

export const BLANK_ACTION: GameAction = {
  id: '', name: '', icon: 'zap', category: 'exploration', target: 'none', outcome: 'notice', animation: 'look_around',
  rolls: [{ instinct: 'perception', label: 'Percepção' }], enabled: true, builtin: false,
}

const labelOf = (r: Pick<ActionRoll, 'instinct' | 'aptitude'>) =>
  `${INSTINCT_LABELS[r.instinct]}${r.aptitude ? ` + ${APTITUDE_LABELS[r.aptitude]}` : ''}`

function Select<T extends string>({ label, value, options, onChange }: { label: string; value: T; options: [T, string][]; onChange: (v: T) => void }) {
  return (
    <label className="flex flex-col gap-0.5 text-xs text-zinc-400">{label}
      <select aria-label={label} className={field} value={value} onChange={(e) => onChange(e.target.value as T)}>
        {options.map(([v, l]) => <option key={v} value={v}>{l}</option>)}
      </select>
    </label>
  )
}

function RollRow({ roll, index, onChange, onRemove, canRemove }: { roll: ActionRoll; index: number; onChange: (r: ActionRoll) => void; onRemove: () => void; canRemove: boolean }) {
  const set = (p: Partial<Pick<ActionRoll, 'instinct' | 'aptitude'>>) => {
    const next = { ...roll, ...p }
    onChange({ ...next, label: labelOf(next) })
  }
  return (
    <div className="flex flex-wrap items-center gap-2">
      <select aria-label={`Instinto ${index + 1}`} className={field} value={roll.instinct} onChange={(e) => set({ instinct: e.target.value as InstinctKey })}>
        {INSTINCT_KEYS.map((k) => <option key={k} value={k}>{INSTINCT_LABELS[k]}</option>)}
      </select>
      <select aria-label={`Aptidão ${index + 1}`} className={field} value={roll.aptitude ?? ''} onChange={(e) => set({ aptitude: (e.target.value || undefined) as AptitudeKey | undefined })}>
        <option value="">Sem aptidão</option>
        {APTITUDES.map((k) => <option key={k} value={k}>{APTITUDE_LABELS[k]}</option>)}
      </select>
      {canRemove && <button type="button" aria-label={`Remover rolagem ${index + 1}`} className="rounded px-2 py-1 text-xs text-zinc-300 hover:bg-zinc-800" onClick={onRemove}>Remover</button>}
    </div>
  )
}

type Patch = (p: Partial<GameAction>) => void

function Fields({ a, patch }: { a: GameAction; patch: Patch }) {
  return (
    <div className="grid gap-2 sm:grid-cols-3">
        <label className="flex flex-col gap-0.5 text-xs text-zinc-400 sm:col-span-2">Nome
          <input aria-label="Nome da ação" className={field} maxLength={40} value={a.name} onChange={(e) => patch({ name: e.target.value })} />
        </label>
        <Select label="Ícone" value={a.icon} options={Object.keys(ACTION_ICONS).map((k) => [k, k])} onChange={(icon) => patch({ icon })} />
        <Select<ActionCategory> label="Categoria" value={a.category} options={Object.entries(CATEGORY_LABELS) as [ActionCategory, string][]} onChange={(category) => patch({ category })} />
        <Select<ActionTargetKind> label="Alvo" value={a.target} options={Object.entries(TARGET_LABELS) as [ActionTargetKind, string][]} onChange={(target) => patch({ target })} />
        <Select<ActionOutcome> label="Consequência" value={a.outcome} options={Object.entries(OUTCOME_LABELS) as [ActionOutcome, string][]} onChange={(outcome) => patch({ outcome })} />
        <Select<Animation> label="Animação" value={a.animation} options={ANIMATIONS.map((k) => [k, ANIMATION_LABELS[k]])} onChange={(animation) => patch({ animation })} />
        {a.target !== 'none' && (
          <label className="flex flex-col gap-0.5 text-xs text-zinc-400">Alcance (células)
            <input aria-label="Alcance" type="number" min={0} max={20} className={field} value={a.range ?? 0} onChange={(e) => patch({ range: Number(e.target.value) })} />
          </label>
        )}
      </div>
  )
}

function Rolls({ a, patch }: { a: GameAction; patch: Patch }) {
  return (
      <fieldset className="flex flex-col gap-1.5">
        <legend className="mb-1 text-xs text-zinc-400">Rolagens (a primeira é a padrão)</legend>
        {a.rolls.map((r, i) => (
          <RollRow key={i} roll={r} index={i} canRemove={a.rolls.length > 1}
            onChange={(n) => patch({ rolls: a.rolls.map((x, j) => (j === i ? n : x)) })} onRemove={() => patch({ rolls: a.rolls.filter((_, j) => j !== i) })} />
        ))}
        {a.rolls.length < 4 && (
          <button type="button" className="self-start rounded bg-zinc-800 px-2 py-1 text-xs hover:bg-zinc-700"
            onClick={() => patch({ rolls: [...a.rolls, { instinct: 'perception', label: 'Percepção' }] })}>Adicionar rolagem</button>
        )}
      </fieldset>
  )
}

/** Create or edit an action: name, icon, category, target, consequence, animation, range and the possible rolls. */
export function ActionForm({ initial, onSave, onCancel }: { initial: GameAction; onSave: (a: GameAction) => Promise<void>; onCancel: () => void }) {
  const [a, setA] = useState(initial)
  const [error, setError] = useState('')
  const patch: Patch = (p) => setA((cur) => ({ ...cur, ...p }))
  const save = () => onSave({ ...a, range: a.target === 'none' ? undefined : a.range || undefined }).then(onCancel, (e: Error) => setError(e.message))
  return (
    <form aria-label={initial.id ? `Editar ${initial.name}` : 'Nova ação'} className="flex flex-col gap-3 rounded-xl bg-zinc-800/40 p-3"
      onSubmit={(e) => { e.preventDefault(); save() }}>
      <Fields a={a} patch={patch} />
      <Rolls a={a} patch={patch} />
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      <div className="flex gap-2">
        <button type="submit" className="rounded bg-emerald-600 px-3 py-1 text-sm font-semibold text-white hover:bg-emerald-500">Salvar</button>
        <button type="button" onClick={onCancel} className="rounded px-3 py-1 text-sm text-zinc-300 hover:bg-zinc-800">Cancelar</button>
      </div>
    </form>
  )
}
