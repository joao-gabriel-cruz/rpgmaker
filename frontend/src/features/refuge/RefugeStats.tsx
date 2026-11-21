import { ATTRIBUTES, RESOURCES, stepValue, type NumericField, type Refuge } from './refuge'

interface Props {
  refuge: Refuge
  isGm: boolean
  onChange: (field: NumericField, value: number) => void
}

const stepCls = 'rounded bg-zinc-800 px-2 text-sm hover:bg-zinc-700 disabled:opacity-40'

function Stat({ label, field, value, isGm, onChange }: { label: string; field: NumericField; value: number; isGm: boolean; onChange: Props['onChange'] }) {
  return (
    <li className="flex items-center gap-2 rounded bg-zinc-900 px-3 py-2">
      <span className="text-sm text-zinc-300">{label}</span>
      <span className="ml-auto flex items-center gap-2">
        {isGm && <button className={stepCls} aria-label={`Diminuir ${label}`} onClick={() => onChange(field, stepValue(field, value, -1))}>−</button>}
        <span aria-label={label} className="min-w-6 text-center font-semibold text-emerald-400">{value}</span>
        {isGm && <button className={stepCls} aria-label={`Aumentar ${label}`} onClick={() => onChange(field, stepValue(field, value, 1))}>+</button>}
      </span>
    </li>
  )
}

export function RefugeStats({ refuge, isGm, onChange }: Props) {
  return (
    <div className="grid max-w-3xl grid-cols-1 gap-4 sm:grid-cols-2">
      {[['Atributos', ATTRIBUTES], ['Recursos', RESOURCES]].map(([title, list]) => (
        <div key={title as string}>
          <h3 className="mb-1 text-sm text-zinc-400">{title as string}</h3>
          <ul className="flex flex-col gap-1">
            {(list as typeof ATTRIBUTES).map((s) => (
              <Stat key={s.key} label={s.label} field={s.key} value={refuge[s.key]} isGm={isGm} onChange={onChange} />
            ))}
          </ul>
        </div>
      ))}
    </div>
  )
}
