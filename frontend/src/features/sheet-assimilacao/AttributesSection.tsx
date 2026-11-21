import { sumValues } from './derive'
import {
  APTITUDE_ADJECTIVES, INSTINCT_ADJECTIVES, INSTINCT_LABELS, KNOWLEDGE_LABELS, PRACTICE_LABELS,
} from './labels'
import type { AptitudeKey, InstinctKey, RollRequest } from './types'
import { INSTINCT_KEYS, KNOWLEDGE_KEYS, PRACTICE_KEYS } from './types'
import { IssueList, Section, Stepper, type SectionProps } from './ui'
import { APTITUDE_MAX, APTITUDE_TOTAL, INSTINCT_MAX, INSTINCT_TOTAL } from './validate'

interface RowProps {
  label: string
  value: number
  adjective: string
  min: number
  max: number
  disabled: boolean
  selected: boolean
  rollable: boolean
  onPick: () => void
  onChange: (v: number) => void
  issues: SectionProps['issues']
  field: string
}

function Row(p: RowProps) {
  return (
    <li className="py-1">
      <div className="flex items-center gap-2">
        <button
          disabled={!p.rollable}
          aria-pressed={p.selected}
          onClick={p.onPick}
          className={`flex-1 rounded px-2 py-1 text-left text-sm ${p.selected ? 'bg-emerald-700 text-white' : 'text-zinc-200 enabled:hover:bg-zinc-800'}`}
        >
          {p.label}
        </button>
        <span className="w-24 text-right text-xs text-zinc-400">{p.adjective}</span>
        <Stepper label={p.label} value={p.value} min={p.min} max={p.max} disabled={p.disabled} onChange={p.onChange} />
      </div>
      <IssueList issues={p.issues} field={p.field} />
    </li>
  )
}

function Counter({ used, limit, label }: { used: number; limit: number; label: string }) {
  const over = used > limit
  return (
    <span className={`text-xs ${over ? 'text-yellow-400' : 'text-zinc-400'}`}>
      {label}: {used}/{limit}
    </span>
  )
}

export interface AttributesProps extends SectionProps {
  selection: Partial<RollRequest>
  rollable: boolean
  onSelect: (s: Partial<RollRequest>) => void
}

function InstinctList({ sheet, onChange, readOnly, issues, selection, rollable, onSelect }: AttributesProps) {
  return (
    <ul>
      {INSTINCT_KEYS.map((key: InstinctKey) => (
        <Row
          key={key} disabled={readOnly} rollable={rollable} issues={issues} label={INSTINCT_LABELS[key]}
          value={sheet.instincts[key]} min={1} max={INSTINCT_MAX} field={`instincts.${key}`}
          adjective={INSTINCT_ADJECTIVES[sheet.instincts[key]] ?? ''}
          selected={selection.instinct === key} onPick={() => onSelect({ ...selection, instinct: key })}
          onChange={(v) => onChange({ ...sheet, instincts: { ...sheet.instincts, [key]: v } })}
        />
      ))}
    </ul>
  )
}

type Group = 'knowledge' | 'practices'

function AptitudeList({ group, keys, labels, props }: { group: Group; keys: readonly AptitudeKey[]; labels: Record<string, string>; props: AttributesProps }) {
  const { sheet, onChange, readOnly, issues, selection, rollable, onSelect } = props
  return (
    <ul>
      {keys.map((key) => {
        const value = (sheet[group] as Record<string, number>)[key]
        return (
          <Row
            key={key} disabled={readOnly} rollable={rollable} issues={issues} label={labels[key]} value={value}
            min={0} max={APTITUDE_MAX} field={`${group}.${key}`} adjective={APTITUDE_ADJECTIVES[value] ?? ''}
            selected={selection.aptitude === key} onPick={() => onSelect({ ...selection, aptitude: key })}
            onChange={(v) => onChange({ ...sheet, [group]: { ...sheet[group], [key]: v } })}
          />
        )
      })}
    </ul>
  )
}

export function AttributesSection(props: AttributesProps) {
  const { sheet, issues } = props
  const apt: Record<string, number> = { ...sheet.knowledge, ...sheet.practices }
  return (
    <Section title="Instintos e Aptidões" aside={<Counter used={sumValues(sheet.instincts)} limit={INSTINCT_TOTAL} label="Instintos" />}>
      <InstinctList {...props} />
      <IssueList issues={issues} field="instincts" />
      <div className="mt-4 flex justify-between border-t border-zinc-800 pt-3">
        <h3 className="text-xs uppercase text-zinc-500">Aptidões</h3>
        <Counter used={sumValues(apt)} limit={APTITUDE_TOTAL} label="Pontos de aptidão" />
      </div>
      <h4 className="mt-2 text-xs text-zinc-500">Conhecimentos</h4>
      <AptitudeList group="knowledge" keys={KNOWLEDGE_KEYS} labels={KNOWLEDGE_LABELS} props={props} />
      <h4 className="mt-2 text-xs text-zinc-500">Práticas</h4>
      <AptitudeList group="practices" keys={PRACTICE_KEYS} labels={PRACTICE_LABELS} props={props} />
      <IssueList issues={issues} field="aptitudes" />
    </Section>
  )
}
