import { INSTINCT_LABELS } from '../sheet-assimilacao/labels'
import type { InstinctKey } from '../sheet-assimilacao/types'
import type { Modifiers, RollState } from './modifiers'

interface Props {
  value: Modifiers
  onChange: (m: Modifiers) => void
  state?: RollState
  isGm: boolean
  onDifficulty: (d12: number) => void
}

function Flag({ label, checked, disabled, hint, onChange }: {
  label: string; checked: boolean; disabled?: boolean; hint?: string; onChange: (v: boolean) => void
}) {
  return (
    <label className={`flex items-center gap-2 text-sm ${disabled ? 'text-zinc-500' : ''}`}>
      <input type="checkbox" checked={checked} disabled={disabled} onChange={(e) => onChange(e.target.checked)} />
      {label}
      {hint && <span className="text-xs text-zinc-500">({hint})</span>}
    </label>
  )
}

function Assimilated({ value, onChange }: Pick<Props, 'value' | 'onChange'>) {
  const options = (Object.keys(INSTINCT_LABELS) as InstinctKey[])
  return (
    <div className="flex items-center gap-2">
      <Flag label="Assimilada" checked={value.assimilated} onChange={(assimilated) => onChange({ ...value, assimilated })} />
      {value.assimilated && (
        <select aria-label="Segundo instinto" value={value.secondInstinct ?? ''}
          onChange={(e) => onChange({ ...value, secondInstinct: (e.target.value || undefined) as InstinctKey | undefined })}
          className="rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm">
          <option value="">Segundo instinto…</option>
          {options.map((k) => <option key={k} value={k}>{INSTINCT_LABELS[k]}</option>)}
        </select>
      )}
    </div>
  )
}

function Difficulty({ value, onChange, state, onDifficulty }: Props) {
  const num = 'w-16 rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm'
  return (
    <div className="flex flex-wrap items-center gap-3 text-sm">
      <label className="flex items-center gap-2">Dificuldade (d12)
        <input type="number" min={0} max={3} aria-label="Dificuldade" className={num} value={value.difficultyD12}
          onChange={(e) => onChange({ ...value, difficultyD12: Math.min(3, Math.max(0, Number(e.target.value) || 0)) })} />
      </label>
      <button type="button" className="rounded bg-zinc-800 px-2 py-1 hover:bg-zinc-700"
        onClick={() => onDifficulty(value.difficultyD12)}>Definir para a próxima rolagem</button>
      {!!state?.difficultyD12 && <span className="text-amber-300">Próxima rolagem: +{state.difficultyD12} d12</span>}
    </div>
  )
}

/** Modifiers applied to the next instinct + aptitude roll. */
export function ModifierPanel(props: Props) {
  const { value, onChange, state, isGm } = props
  const noSession = state ? !state.sessionActive : false
  const once = (used: boolean | undefined) => (used ? 'usado nesta sessão' : noSession ? 'requer sessão ativa' : undefined)
  return (
    <fieldset aria-label="Modificadores" className="flex flex-col gap-2 rounded border border-zinc-800 p-3">
      <legend className="px-1 text-sm text-zinc-400">Modificadores</legend>
      <Flag label="Empenho" hint="−1 Determinação" checked={value.effort} onChange={(effort) => onChange({ ...value, effort })} />
      <Flag label="Ocupação" checked={value.useOccupation && !state?.occupationUsed} hint={once(state?.occupationUsed)}
        disabled={!!state?.occupationUsed || noSession} onChange={(useOccupation) => onChange({ ...value, useOccupation })} />
      <Flag label="Evento Marcante" checked={value.useMarkingEvent && !state?.markingEventUsed} hint={once(state?.markingEventUsed)}
        disabled={!!state?.markingEventUsed || noSession} onChange={(useMarkingEvent) => onChange({ ...value, useMarkingEvent })} />
      <Assimilated value={value} onChange={onChange} />
      {isGm && <Difficulty {...props} />}
    </fieldset>
  )
}
