import { CATALOG } from './catalog'
import { CATEGORY_LABELS, type Category } from './catalogData'

type Variant = 'pine' | 'round'

interface Props {
  value: string
  onChange: (id: string) => void
  variant: Variant
  onVariant: (v: Variant) => void
}

const CATEGORIES = Object.keys(CATEGORY_LABELS) as Category[]
const pick = (on: boolean) =>
  on ? 'bg-emerald-600 text-white' : 'bg-zinc-800 text-zinc-200 hover:bg-zinc-700'

function TreeVariants({ variant, onVariant }: Pick<Props, 'variant' | 'onVariant'>) {
  return (
    <div className="flex gap-1 text-xs">
      {(['pine', 'round'] as const).map((v) => (
        <button key={v} onClick={() => onVariant(v)} className={`flex-1 rounded px-1 py-1 ${pick(variant === v)}`}>
          {v === 'pine' ? 'Pinheiro' : 'Frondosa'}
        </button>
      ))}
    </div>
  )
}

function Category({ cat, value, onChange }: { cat: Category } & Pick<Props, 'value' | 'onChange'>) {
  return (
    <div className="flex flex-col gap-1">
      <h4 className="text-xs text-zinc-500">{CATEGORY_LABELS[cat]}</h4>
      {CATALOG.filter((p) => p.category === cat).map((p) => (
        <button key={p.id} onClick={() => onChange(p.id)} className={`rounded px-2 py-1 text-left text-sm ${pick(value === p.id)}`}>
          {p.name}
          <span className="float-right text-xs opacity-60">
            {p.width}×{p.depth}
          </span>
        </button>
      ))}
    </div>
  )
}

export function PrefabPalette({ value, onChange, variant, onVariant }: Props) {
  return (
    <div className="flex flex-col gap-2">
      {CATEGORIES.map((cat) => (
        <Category key={cat} cat={cat} value={value} onChange={onChange} />
      ))}
      {value === 'tree' && <TreeVariants variant={variant} onVariant={onVariant} />}
    </div>
  )
}
