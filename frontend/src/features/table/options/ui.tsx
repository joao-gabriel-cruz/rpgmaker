import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { Kbd, Tip } from '../Tip'

const FOCUS = 'focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-400'
const MOTION = 'transition-[background-color,color,box-shadow,transform] motion-reduce:transition-none'

export const BTN = `rounded-md bg-zinc-800 px-2 py-1 text-xs text-zinc-200 hover:bg-zinc-700 active:scale-[0.96] disabled:opacity-40 disabled:active:scale-100 ${MOTION} ${FOCUS}`
export const BTN_DANGER = `${BTN} text-red-300 hover:bg-red-950`
export const INPUT = `rounded-md bg-zinc-800 px-2 py-1 text-sm text-zinc-100 placeholder:text-zinc-500 ${FOCUS}`

export function Section({ title, children, aside }: { title: string; children: ReactNode; aside?: ReactNode }) {
  return (
    <section className="flex flex-col gap-1.5">
      <div className="flex items-center justify-between">
        <h3 className="text-[11px] font-medium uppercase tracking-wide text-zinc-400">{title}</h3>
        {aside}
      </div>
      {children}
    </section>
  )
}

export interface SegOption<T extends string | number> {
  id: T
  label: string
  icon?: LucideIcon
  /** Visible text instead of an icon (e.g. brush sizes). */
  text?: string
  keys?: string
}

interface SegProps<T extends string | number> {
  label: string
  value: T | null
  options: SegOption<T>[]
  onChange: (id: T) => void
  /** Icon-only buttons are squares; text ones share the row. */
  className?: string
}

/** Segmented control: one pressed option at a time, icon buttons with the name and key in a tooltip. */
export function Segmented<T extends string | number>({ label, value, options, onChange, className = '' }: SegProps<T>) {
  return (
    <div role="group" aria-label={label} className={`flex gap-0.5 rounded-lg bg-zinc-950/60 p-0.5 ${className}`}>
      {options.map((o) => {
        const on = o.id === value
        const Icon = o.icon
        return (
          <button key={o.id} type="button" aria-label={o.label} aria-pressed={on} aria-keyshortcuts={o.keys}
            onClick={() => onChange(o.id)}
            className={`group relative flex h-8 min-w-8 flex-1 items-center justify-center rounded-md px-1.5 text-xs ${MOTION} ${FOCUS} active:scale-[0.96] ${
              on ? 'bg-emerald-600 text-white shadow-sm' : 'text-zinc-300 hover:bg-zinc-800 hover:text-white'}`}>
            {Icon && <Icon size={16} strokeWidth={on ? 2.25 : 1.5} aria-hidden />}
            {o.text}
            <Tip side="bottom">{o.label}{o.keys && <Kbd>{o.keys}</Kbd>}</Tip>
          </button>
        )
      })}
    </div>
  )
}

interface StepperProps {
  label: string
  value: number
  min: number
  max: number
  onChange: (v: number) => void
}

/** Minus / value / plus, clamped to the range. */
export function Stepper({ label, value, min, max, onChange }: StepperProps) {
  const step = (d: number) => onChange(Math.min(max, Math.max(min, value + d)))
  const b = `flex h-8 w-8 items-center justify-center rounded-md bg-zinc-800 text-zinc-200 hover:bg-zinc-700 active:scale-[0.96] disabled:opacity-40 ${MOTION} ${FOCUS}`
  return (
    <div role="group" aria-label={label} className="flex items-center gap-1">
      <button type="button" aria-label={`Diminuir ${label.toLowerCase()}`} disabled={value <= min} onClick={() => step(-1)} className={b}>−</button>
      <output className="w-8 text-center text-sm tabular-nums text-zinc-100">{value}</output>
      <button type="button" aria-label={`Aumentar ${label.toLowerCase()}`} disabled={value >= max} onClick={() => step(1)} className={b}>+</button>
    </div>
  )
}

/** On/off row with an icon, used for "Mostrar grade" and the draft switch. */
export function Switch({ label, on, onChange, icon: Icon }: { label: string; on: boolean; onChange: () => void; icon?: LucideIcon }) {
  return (
    <button type="button" role="switch" aria-checked={on} aria-label={label} onClick={onChange}
      className={`flex items-center gap-2 rounded-md px-1 py-1 text-sm text-zinc-200 hover:bg-zinc-800 active:scale-[0.98] ${MOTION} ${FOCUS}`}>
      {Icon && <Icon size={16} strokeWidth={1.5} aria-hidden className="text-zinc-400" />}
      <span className="flex-1 text-left">{label}</span>
      <span aria-hidden className={`flex h-4 w-7 items-center rounded-full p-0.5 ${MOTION} ${on ? 'bg-emerald-600' : 'bg-zinc-700'}`}>
        <span className={`h-3 w-3 rounded-full bg-white shadow transition-transform motion-reduce:transition-none ${on ? 'translate-x-3' : ''}`} />
      </span>
    </button>
  )
}

export function Hint({ children }: { children: ReactNode }) {
  return <p className="text-xs leading-relaxed text-zinc-400">{children}</p>
}
