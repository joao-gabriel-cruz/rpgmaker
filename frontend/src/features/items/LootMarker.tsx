import { ItemIcon } from './ItemIcon'
import type { Loot } from './loot'

interface Props {
  loot: Loot
  isGm: boolean
  /** The viewer can pick this up now (visible and next to one of their tokens). */
  canPick: boolean
  onPick: () => void
  onRemove: () => void
}

/** DOM marker of one loot: floating glowing icon for players; translucent with difficulty for the Mestre when hidden. */
export function LootMarker({ loot, isGm, canPick, onPick, onRemove }: Props) {
  const color = loot.item.color ?? '#fbbf24'
  const ghost = loot.hidden
  return (
    <div data-testid={`loot-${loot.id}`} data-hidden={ghost} className="group flex flex-col items-center">
      <button type="button" disabled={!canPick} onClick={onPick}
        aria-label={canPick ? `Pegar ${loot.item.name}` : `${loot.item.name}${ghost ? ' (escondido)' : ''}`}
        title={canPick ? `Pegar ${loot.item.name}` : loot.item.name}
        style={{ boxShadow: ghost ? undefined : `0 0 14px 3px ${color}`, borderColor: color }}
        className={`flex h-9 w-9 items-center justify-center rounded-full border bg-zinc-950/80 motion-safe:animate-bounce ${ghost ? 'opacity-50' : ''} ${canPick ? 'cursor-pointer ring-2 ring-white' : 'cursor-default'}`}>
        <ItemIcon name={loot.item.icon} color={color} size={20} />
      </button>
      {isGm && <span className="mt-0.5 rounded bg-zinc-900/90 px-1 text-[10px] text-amber-300">Dif. {loot.difficulty}{loot.qty > 1 ? ` · ×${loot.qty}` : ''}</span>}
      {isGm && <button type="button" aria-label={`Remover ${loot.item.name} do mapa`} onClick={onRemove}
        className="hidden rounded bg-red-900/90 px-1 text-[10px] text-red-100 group-hover:block">remover</button>}
    </div>
  )
}
