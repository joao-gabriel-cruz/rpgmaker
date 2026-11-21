import { effectLines, KIND_LABEL, type ItemDraft } from './items'
import { ItemIcon } from './ItemIcon'

/** The item as a card: icon on its color, name, kind, stats, traits, effects and description. */
export function ItemPreview({ item, label }: { item: ItemDraft; label?: string }) {
  const lines = effectLines(item.effects)
  return (
    <article aria-label={label ?? `Cartão de ${item.name || 'item sem nome'}`} className="flex gap-3 rounded-lg border border-zinc-700 bg-zinc-950 p-3">
      <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg" style={{ background: `${item.color}33`, boxShadow: `inset 0 0 0 1px ${item.color}` }}>
        <ItemIcon name={item.icon} color={item.color} size={26} />
      </span>
      <div className="min-w-0 flex-1 text-sm">
        <p className="truncate font-semibold text-zinc-100">{item.name || 'Item sem nome'}</p>
        <p className="text-xs text-zinc-400">{KIND_LABEL[item.kind]} · Qualidade {item.quality} · Escassez {item.scarcity} · Pressão {item.pressure}</p>
        {item.traits.length > 0 && <p className="mt-1 flex flex-wrap gap-1">{item.traits.map((t) => <span key={t} className="rounded-full bg-zinc-800 px-2 text-xs">{t}</span>)}</p>}
        {lines.map((l) => <p key={l} className="text-xs text-emerald-300">{l}</p>)}
        {item.description && <p className="mt-1 text-xs text-zinc-300">{item.description}</p>}
      </div>
    </article>
  )
}
