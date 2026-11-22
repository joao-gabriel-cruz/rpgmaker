const stepCls = 'rounded bg-zinc-800 px-2 text-sm hover:bg-zinc-700 disabled:opacity-40'

interface Props {
  label: string
  value: number
  min?: number
  max?: number
  editable: boolean
  onChange: (value: number) => void
}

/** −/+ counter; players see only the value. */
export function Stepper({ label, value, min = 0, max = Infinity, editable, onChange }: Props) {
  return (
    <span className="flex items-center gap-2">
      {editable && <button className={stepCls} aria-label={`Diminuir ${label}`} disabled={value <= min} onClick={() => onChange(value - 1)}>−</button>}
      <span aria-label={label} className="min-w-6 text-center font-semibold text-emerald-400">{value}</span>
      {editable && <button className={stepCls} aria-label={`Aumentar ${label}`} disabled={value >= max} onClick={() => onChange(value + 1)}>+</button>}
    </span>
  )
}
