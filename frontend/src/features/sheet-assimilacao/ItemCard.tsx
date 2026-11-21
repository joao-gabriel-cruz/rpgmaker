import { useState } from 'react'
import type { Item } from './types'
import { btnCls, inputCls } from './ui'

interface Props {
  item: Item
  readOnly: boolean
  moveLabel: string
  canMove: boolean
  onChange: (item: Item) => void
  onMove: () => void
  onRemove: () => void
}

const STATS = [['quality', 'Qualidade'], ['scarcity', 'Escassez'], ['pressure', 'Pressão']] as const

function Traits({ item, readOnly, onChange }: Pick<Props, 'item' | 'readOnly' | 'onChange'>) {
  const [draft, setDraft] = useState('')
  const add = () => {
    const t = draft.trim()
    if (t) onChange({ ...item, traits: [...item.traits, t] })
    setDraft('')
  }
  return (
    <div className="mt-2 flex flex-wrap items-center gap-1">
      {item.traits.map((t, i) => (
        <span key={`${t}-${i}`} className="inline-flex items-center gap-1 rounded-full bg-zinc-800 px-2 py-0.5 text-xs text-zinc-200">
          {t}
          {!readOnly && <button aria-label={`Remover característica ${t}`} onClick={() => onChange({ ...item, traits: item.traits.filter((_, j) => j !== i) })}>×</button>}
        </span>
      ))}
      {!readOnly && (
        <input className={`${inputCls} w-40`} aria-label="Nova característica do item" placeholder="+ característica" value={draft} onChange={(e) => setDraft(e.target.value)} onKeyDown={(e) => e.key === 'Enter' && add()} onBlur={add} />
      )}
    </div>
  )
}

export function ItemCard({ item, readOnly, moveLabel, canMove, onChange, onMove, onRemove }: Props) {
  return (
    <li className="rounded border border-zinc-700 bg-zinc-950 p-2">
      <div className="flex gap-2">
        <input className={`${inputCls} flex-1`} aria-label="Nome do item" placeholder="Nome do item" value={item.name} disabled={readOnly} onChange={(e) => onChange({ ...item, name: e.target.value })} />
        <button className={btnCls} disabled={readOnly || !canMove} title={canMove ? undefined : 'Destino cheio'} onClick={onMove}>{moveLabel}</button>
        <button className={btnCls} disabled={readOnly} aria-label={`Remover item ${item.name}`} onClick={onRemove}>Remover</button>
      </div>
      <div className="mt-2 flex gap-3 text-xs text-zinc-400">
        {STATS.map(([key, label]) => (
          <label key={key} className="flex items-center gap-1">{label}
            <input type="number" min={0} className={`${inputCls} w-14`} value={item[key]} disabled={readOnly} onChange={(e) => onChange({ ...item, [key]: Math.max(0, Number(e.target.value) || 0) })} />
          </label>
        ))}
      </div>
      <Traits item={item} readOnly={readOnly} onChange={onChange} />
    </li>
  )
}
