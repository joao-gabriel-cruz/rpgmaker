import { ItemCard } from './ItemCard'
import { BACKPACK_LIMIT, BODY_LIMIT, newItem, type Item, type Sheet } from './types'
import { btnCls, IssueList, Section, type SectionProps } from './ui'

type Place = 'body' | 'backpack'
type Inv = Sheet['inventory']
const PLACES: { key: Place; title: string; limit: number; other: Place; moveLabel: string }[] = [
  { key: 'body', title: 'Corpo', limit: BODY_LIMIT, other: 'backpack', moveLabel: 'Mover para mochila' },
  { key: 'backpack', title: 'Mochila', limit: BACKPACK_LIMIT, other: 'body', moveLabel: 'Mover para corpo' },
]
const LIMITS = { body: BODY_LIMIT, backpack: BACKPACK_LIMIT }

function moveItem(inv: Inv, from: Place, to: Place, i: number): Inv {
  if (inv[to].length >= LIMITS[to]) return inv
  return { ...inv, [from]: inv[from].filter((_, j) => j !== i), [to]: [...inv[to], inv[from][i]] }
}

interface ColProps extends SectionProps {
  place: (typeof PLACES)[number]
  set: (next: Inv) => void
}

function PlaceColumn({ sheet, readOnly, issues, place: p, set }: ColProps) {
  const inv = sheet.inventory
  const update = (i: number, item: Item) => set({ ...inv, [p.key]: inv[p.key].map((x, j) => (j === i ? item : x)) })
  const remove = (i: number) => set({ ...inv, [p.key]: inv[p.key].filter((_, j) => j !== i) })
  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-sm text-zinc-300">{p.title} ({inv[p.key].length}/{p.limit})</h3>
        <button className={btnCls} disabled={readOnly || inv[p.key].length >= p.limit} aria-label={`Adicionar item ao ${p.title.toLowerCase()}`} onClick={() => set({ ...inv, [p.key]: [...inv[p.key], newItem()] })}>
          + Item
        </button>
      </div>
      <ul className="space-y-2">
        {inv[p.key].map((item, i) => (
          <ItemCard key={i} item={item} readOnly={readOnly} moveLabel={p.moveLabel} canMove={inv[p.other].length < LIMITS[p.other]} onChange={(it) => update(i, it)} onMove={() => set(moveItem(inv, p.key, p.other, i))} onRemove={() => remove(i)} />
        ))}
      </ul>
      <IssueList issues={issues} field={`inventory.${p.key}`} />
    </div>
  )
}

export function InventorySection(props: SectionProps) {
  const set = (next: Inv) => props.onChange({ ...props.sheet, inventory: next })
  return (
    <Section title="Inventário">
      <div className="grid grid-cols-2 gap-4">
        {PLACES.map((p) => <PlaceColumn key={p.key} {...props} place={p} set={set} />)}
      </div>
    </Section>
  )
}
