import { ICON_LABEL } from './iconLabels'
import { ItemIcon } from './ItemIcon'
import { ICON_NAMES } from './items'

/** Grid of the curated icons; the selected one is marked with aria-pressed. */
export function IconPicker({ value, color, onChange }: { value: string; color: string; onChange: (icon: string) => void }) {
  return (
    <div role="group" aria-label="Ícone do item" className="grid grid-cols-8 gap-1">
      {ICON_NAMES.map((n) => (
        <button key={n} type="button" aria-label={ICON_LABEL[n]} title={ICON_LABEL[n]} aria-pressed={value === n} onClick={() => onChange(n)}
          className={`flex h-8 w-8 items-center justify-center rounded ${value === n ? 'bg-zinc-700 ring-2 ring-emerald-400' : 'bg-zinc-900 hover:bg-zinc-800'}`}>
          <ItemIcon name={n} color={value === n ? color : undefined} size={18} />
        </button>
      ))}
    </div>
  )
}
