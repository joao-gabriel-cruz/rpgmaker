import { useEffect } from 'react'
import { useActiveCharacter } from '../hotbar/useActiveCharacter'
import { useTableToolOptional } from '../table/useTableTool'
import { useActionFlowOptional } from './ActionFlow'
import { CATEGORY_LABELS, CATEGORY_ORDER, iconOf, useActions, type GameAction } from './actions'

interface Props {
  campaignId: string
  isGm: boolean
  /** `bar` is the compact icon row above the hotbar; `list` shows names (Rolagens tab). */
  layout?: 'bar' | 'list'
}

const BTN = 'group relative flex items-center gap-1.5 rounded-lg transition-[background-color,color,transform] active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-400'

const hint = (a: GameAction) => `${a.rolls[0]?.label ?? ''}${a.target !== 'none' ? ' · escolha o alvo no mapa' : ''}`

function ActionButton({ action, number, armed, disabled, layout, onPick }: {
  action: GameAction; number?: number; armed: boolean; disabled: boolean; layout: 'bar' | 'list'; onPick: () => void
}) {
  const Icon = iconOf(action.icon)
  const tone = armed ? 'bg-emerald-600 text-white' : 'bg-zinc-800 text-zinc-100 hover:bg-zinc-700'
  return (
    <button type="button" aria-label={action.name} aria-pressed={armed} disabled={disabled} onClick={onPick}
      className={`${BTN} ${tone} disabled:opacity-40 ${layout === 'bar' ? 'h-9 w-9 justify-center' : 'px-2.5 py-1.5 text-sm'}`}>
      <Icon size={layout === 'bar' ? 18 : 16} aria-hidden />
      {layout === 'list' && <span className="truncate">{action.name}</span>}
      {layout === 'bar' && number !== undefined && number < 10 && (
        <span aria-hidden className="absolute -right-1 -top-1 rounded bg-zinc-950 px-1 font-mono text-[9px] text-zinc-400">{number}</span>
      )}
      {layout === 'bar' && <ActionTip action={action} number={number} />}
    </button>
  )
}

function ActionTip({ action, number }: { action: GameAction; number?: number }) {
  return (
    <span role="tooltip" aria-hidden
      className="pointer-events-none absolute bottom-full left-1/2 z-40 mb-2 -translate-x-1/2 whitespace-nowrap rounded-md bg-zinc-950 px-2 py-1 text-left text-xs text-zinc-100 opacity-0 shadow-lg ring-1 ring-white/10 transition-opacity duration-100 group-hover:opacity-100 group-focus-visible:opacity-100 motion-reduce:transition-none">
      <strong className="block">{action.name}{number ? ` (${number})` : ''}</strong>
      <span className="text-zinc-400">{hint(action)}</span>
    </span>
  )
}

// Digits 1-9 start the n-th action while the Mover tool is active (the editors use digits for their own tools).
function useActionDigits(enabled: boolean, list: GameAction[], start: (a: GameAction) => void) {
  const tool = useTableToolOptional()?.tool
  useEffect(() => {
    if (!enabled || (tool && tool !== 'move')) return
    const onKey = (e: KeyboardEvent) => {
      const t = e.target
      const typing = t instanceof HTMLElement && (t.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(t.tagName))
      if (typing || e.metaKey || e.ctrlKey || e.altKey || e.shiftKey || !/^[1-9]$/.test(e.key)) return
      const a = list[Number(e.key) - 1]
      if (a) start(a)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [enabled, tool, list, start])
}

/** Action buttons grouped by category for the active character; picking one starts the action flow. */
export function ActionBar({ campaignId, isGm, layout = 'bar' }: Props) {
  const { actions } = useActions(campaignId)
  const { active } = useActiveCharacter(campaignId, isGm)
  const flow = useActionFlowOptional()
  const usable = !!active && active.kind !== 'npc' && !!flow
  const list = CATEGORY_ORDER.flatMap((c) => actions.filter((a) => a.enabled && a.category === c))
  const start = (a: GameAction) => active && flow?.begin(a, active)
  useActionDigits(usable && layout === 'bar', list, start)
  if (list.length === 0) return null
  if (!usable) return layout === 'list' ? <p className="text-sm text-zinc-400">Escolha um personagem para agir.</p> : null
  return (
    <div role="toolbar" aria-label="Ações" className={layout === 'bar' ? 'flex flex-wrap items-center gap-x-3 gap-y-1' : 'flex flex-col gap-2'}>
      {CATEGORY_ORDER.map((c) => {
        const group = list.filter((a) => a.category === c)
        return group.length > 0 && (
          <div key={c} role="group" aria-label={CATEGORY_LABELS[c]} className={layout === 'bar' ? 'flex items-center gap-1' : 'flex flex-wrap gap-1.5'}>
            {layout === 'list' && <span className="basis-full text-[11px] text-zinc-500">{CATEGORY_LABELS[c]}</span>}
            {group.map((a) => (
              <ActionButton key={a.id} action={a} number={list.indexOf(a) + 1} armed={flow?.pending?.action.id === a.id} disabled={false}
                layout={layout} onPick={() => start(a)} />
            ))}
          </div>
        )
      })}
    </div>
  )
}
