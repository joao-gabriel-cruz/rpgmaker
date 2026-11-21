import type { ReactNode } from 'react'
import type { Sheet, SheetIssue } from './types'

export interface SectionProps {
  sheet: Sheet
  onChange: (next: Sheet) => void
  readOnly: boolean
  issues: SheetIssue[]
}

export const inputCls =
  'rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm text-zinc-100 disabled:opacity-60 focus:border-emerald-500 focus:outline-none'
export const btnCls =
  'rounded border border-zinc-700 bg-zinc-800 px-2 py-1 text-sm text-zinc-200 hover:bg-zinc-700 disabled:cursor-not-allowed disabled:opacity-40'

export function Section({ title, children, aside }: { title: string; children: ReactNode; aside?: ReactNode }) {
  return (
    <section className="rounded-lg border border-zinc-800 bg-zinc-900/60 p-4">
      <header className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-emerald-400">{title}</h2>
        {aside}
      </header>
      {children}
    </section>
  )
}

export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="flex flex-col gap-1 text-xs text-zinc-400">
      {label}
      {children}
    </label>
  )
}

function issuesFor(issues: SheetIssue[], field: string): SheetIssue[] {
  return issues.filter((i) => i.field === field || i.field.startsWith(`${field}.`))
}

export function IssueList({ issues, field }: { issues: SheetIssue[]; field: string }) {
  const list = issuesFor(issues, field)
  if (list.length === 0) return null
  return (
    <ul className="mt-1 space-y-1">
      {list.map((i) => (
        <li
          key={`${i.code}:${i.field}`}
          role={i.blocking ? 'alert' : 'status'}
          className={`rounded px-2 py-1 text-xs ${i.blocking ? 'bg-red-950 text-red-300' : 'bg-yellow-950 text-yellow-300'}`}
        >
          {i.message}
        </li>
      ))}
    </ul>
  )
}

interface StepperProps {
  label: string
  value: number
  min: number
  max: number
  disabled?: boolean
  onChange: (v: number) => void
}

export function Stepper({ label, value, min, max, disabled, onChange }: StepperProps) {
  return (
    <span className="inline-flex items-center gap-1">
      <button className={btnCls} aria-label={`Diminuir ${label}`} disabled={disabled || value <= min} onClick={() => onChange(value - 1)}>
        −
      </button>
      <span className="w-5 text-center font-mono text-sm" aria-label={`Valor de ${label}`}>
        {value}
      </span>
      <button className={btnCls} aria-label={`Aumentar ${label}`} disabled={disabled || value >= max} onClick={() => onChange(value + 1)}>
        +
      </button>
    </span>
  )
}

export function TextList({ values, onChange, readOnly, label }: { values: string[]; onChange: (v: string[]) => void; readOnly: boolean; label: string }) {
  return (
    <>
      {values.map((v, i) => (
        <Field key={i} label={`${label} ${i + 1}`}>
          <input
            className={inputCls}
            value={v}
            disabled={readOnly}
            onChange={(e) => onChange(values.map((x, j) => (j === i ? e.target.value : x)))}
          />
        </Field>
      ))}
    </>
  )
}
