import { APTITUDE_LABELS, INSTINCT_LABELS } from '../sheet-assimilacao/labels'
import { INSTINCT_KEYS, KNOWLEDGE_KEYS, PRACTICE_KEYS, type AptitudeKey, type InstinctKey, type Sheet } from '../sheet-assimilacao/types'
import { aptitudeValue, type PoolSelection } from './pool'

function Pick({ label, value, dice, active, onClick }: {
  label: string; value: number; dice: string; active: boolean; onClick: () => void
}) {
  return (
    <button type="button" data-pick aria-pressed={active} aria-label={`${label} ${value}, ${dice}`} onClick={onClick}
      className={`flex flex-col items-start rounded-lg px-2.5 py-1.5 text-left transition-[background-color,color,box-shadow,transform] active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-400 ${
        active ? 'bg-emerald-600 text-white shadow-md shadow-emerald-950/50' : 'bg-zinc-800/80 text-zinc-200 hover:bg-zinc-700'}`}>
      <span className="text-xs leading-tight">{label}</span>
      <span className="flex items-baseline gap-1.5">
        <span className="text-lg font-semibold leading-tight tabular-nums">{value}</span>
        <span className={`text-[11px] ${active ? 'text-emerald-100' : 'text-zinc-400'}`}>{dice}</span>
      </span>
    </button>
  )
}

interface Props {
  sheet: Sheet
  selection: PoolSelection
  instinctDie: string
  onChange: (s: PoolSelection) => void
}

function Group({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <fieldset className="min-w-0">
      <legend className="mb-1 text-xs font-medium uppercase tracking-wide text-zinc-400">{title}</legend>
      <div className="grid grid-cols-3 gap-1.5">{children}</div>
    </fieldset>
  )
}

function Aptitudes({ title, keys, sheet, selection, onChange }: Omit<Props, 'instinctDie'> & { title: string; keys: readonly AptitudeKey[] }) {
  return (
    <Group title={title}>
      {keys.map((k) => {
        const v = aptitudeValue(sheet, k)
        const active = selection.aptitude === k
        return <Pick key={k} label={APTITUDE_LABELS[k]} value={v} dice={`${v} d10`} active={active}
          onClick={() => onChange({ ...selection, aptitude: active ? undefined : k })} />
      })}
    </Group>
  )
}

/** Instinct and aptitude buttons showing each value and how many dice it adds; click again to clear an aptitude. */
export function ComposerGrids({ sheet, selection, instinctDie, onChange }: Props) {
  return (
    <div className="flex flex-col gap-3">
      <Group title="Instinto">
        {INSTINCT_KEYS.map((k: InstinctKey) => (
          <Pick key={k} label={INSTINCT_LABELS[k]} value={sheet.instincts[k]} dice={`${sheet.instincts[k]} ${instinctDie}`}
            active={selection.instinct === k} onClick={() => onChange({ ...selection, instinct: k })} />
        ))}
      </Group>
      <Aptitudes title="Conhecimentos" keys={KNOWLEDGE_KEYS} {...{ sheet, selection, onChange }} />
      <Aptitudes title="Práticas" keys={PRACTICE_KEYS} {...{ sheet, selection, onChange }} />
    </div>
  )
}
