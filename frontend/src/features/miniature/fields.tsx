import { btnCls, inputCls } from '../sheet-assimilacao/ui'

export function SelectField<T extends string>({ label, value, options, labels, onChange, disabled }: {
  label: string
  value: T
  options: readonly T[]
  labels: Record<T, string>
  onChange: (v: T) => void
  disabled?: boolean
}) {
  return (
    <label className="flex flex-col gap-1 text-xs text-zinc-400">
      {label}
      <select className={inputCls} value={value} disabled={disabled} onChange={(e) => onChange(e.target.value as T)}>
        {options.map((o) => <option key={o} value={o}>{labels[o]}</option>)}
      </select>
    </label>
  )
}

/** Palette swatches plus a free color input. */
export function ColorField({ label, value, palette, onChange, disabled }: {
  label: string
  value: string
  palette: readonly string[]
  onChange: (v: string) => void
  disabled?: boolean
}) {
  return (
    <div className="flex flex-col gap-1 text-xs text-zinc-400">
      <span>{label}</span>
      <div className="flex flex-wrap items-center gap-1">
        {palette.map((c) => (
          <button key={c} type="button" aria-label={`${label} ${c}`} aria-pressed={c === value} disabled={disabled}
            onClick={() => onChange(c)} style={{ backgroundColor: c }}
            className={`h-5 w-5 rounded border ${c === value ? 'border-emerald-400 ring-1 ring-emerald-400' : 'border-zinc-600'}`} />
        ))}
        <input type="color" aria-label={`${label} personalizada`} value={value} disabled={disabled}
          onChange={(e) => onChange(e.target.value)} className="h-6 w-8 cursor-pointer rounded border border-zinc-700 bg-transparent" />
      </div>
    </div>
  )
}

export function Toggle({ label, checked, onChange, disabled }: { label: string; checked: boolean; onChange: (v: boolean) => void; disabled?: boolean }) {
  return (
    <label className="flex items-center gap-2 text-sm text-zinc-300">
      <input type="checkbox" checked={checked} disabled={disabled} onChange={(e) => onChange(e.target.checked)} />
      {label}
    </label>
  )
}

export function Group({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <fieldset className="flex flex-col gap-2 rounded border border-zinc-800 p-3">
      <legend className="px-1 text-xs font-semibold uppercase tracking-wide text-zinc-400">{title}</legend>
      <div className="grid gap-3 sm:grid-cols-2">{children}</div>
    </fieldset>
  )
}

export { btnCls }
