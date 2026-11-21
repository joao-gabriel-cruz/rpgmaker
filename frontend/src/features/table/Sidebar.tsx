import { PanelRightClose, PanelRightOpen } from 'lucide-react'
import { useEffect, useState, type ReactNode } from 'react'
import { CharactersSection } from '../characters/CharactersSection'
import { ConflictPanel } from '../conflict/ConflictPanel'
import { useConflict } from '../conflict/ConflictContext'
import { PendingEffectsView, usePendingEffects } from '../effects/PendingEffects'
import { ItemsPanel } from '../items/ItemsPanel'
import { useRealtime } from '../realtime/realtimeContext'
import { RefugeSection } from '../refuge/RefugeSection'
import { RollLog } from '../rolls/RollLog'
import { tabsFor, type SidebarState, type SidebarTabDef, type SidebarTabId } from './sidebarTabs'
import { ICON_BTN } from './surface'
import { Tip } from './Tip'

interface Props {
  campaignId: string
  isGm: boolean
  sidebar: SidebarState
}

/** Count of rolls that arrived while the Rolagens tab was not on screen. */
function useUnseenRolls(visible: boolean) {
  const { subscribe } = useRealtime()
  const [count, setCount] = useState(0)
  useEffect(() => subscribe('roll.rolled', () => setCount((c) => (visible ? 0 : c + 1))), [subscribe, visible])
  if (visible && count) setCount(0)
  return count
}

function ConflictTab() {
  const state = useConflict()
  if (!state?.conflict && !state?.isGm) return <p className="text-sm text-zinc-400">Nenhuma partitura de conflito ativa.</p>
  return <ConflictPanel />
}

function TabButton({ def, active, badge, onClick }: { def: SidebarTabDef; active: boolean; badge: string; onClick: () => void }) {
  const Icon = def.icon
  return (
    <button type="button" role="tab" id={`tab-${def.id}`} aria-selected={active} aria-controls={`panel-${def.id}`}
      aria-label={badge ? `${def.label} (${badge})` : def.label} onClick={onClick}
      className={`group relative ${ICON_BTN} ${active ? 'bg-zinc-800 text-white' : ''}`}>
      <Icon size={20} strokeWidth={active ? 2.25 : 1.5} aria-hidden />
      {badge && (
        <span aria-hidden data-testid={`badge-${def.id}`}
          className="absolute -right-0.5 -top-0.5 min-w-4 rounded-full bg-amber-500 px-1 text-center text-[10px] font-bold leading-4 text-black">
          {badge}
        </span>
      )}
      <Tip side="left">{def.label}</Tip>
    </button>
  )
}

function Panel({ id, active, children }: { id: SidebarTabId; active: boolean; children: ReactNode }) {
  return (
    <div role="tabpanel" id={`panel-${id}`} aria-labelledby={`tab-${id}`} hidden={!active} className="flex flex-col gap-3">
      {children}
    </div>
  )
}

function badgeOf(id: SidebarTabId, n: { unseenRolls: number; pending: number; conflict: boolean }): string {
  if (id === 'rolls') return n.unseenRolls ? 'novo' : ''
  if (id === 'effects') return n.pending ? String(n.pending) : ''
  return id === 'conflict' && n.conflict ? '•' : ''
}

/**
 * Collapsible right sidebar (~360px). Every panel stays mounted (hidden when inactive) so live
 * subscriptions and badges keep working while the sidebar is collapsed.
 */
export function Sidebar({ campaignId, isGm, sidebar }: Props) {
  const { open, tab, setOpen, setTab } = sidebar
  const pending = usePendingEffects(campaignId)
  const conflict = useConflict()?.conflict
  const unseenRolls = useUnseenRolls(open && tab === 'rolls')
  const badge = (id: SidebarTabId) => badgeOf(id, { unseenRolls, pending: pending.list.length, conflict: !!conflict })
  const choose = (id: SidebarTabId) => (open && tab === id ? setOpen(false) : setTab(id))
  return (
    <aside aria-label="Barra lateral" className="z-20 flex shrink-0 border-l border-zinc-800 bg-zinc-950">
      <div hidden={!open} className="flex min-h-0 w-[19.5rem] flex-col overflow-y-auto p-3 xl:w-[20.5rem]">
        <Panel id="group" active={tab === 'group'}><CharactersSection campaignId={campaignId} isGm={isGm} /></Panel>
        <Panel id="rolls" active={tab === 'rolls'}><RollLog campaignId={campaignId} isGm={isGm} /></Panel>
        <Panel id="conflict" active={tab === 'conflict'}><ConflictTab /></Panel>
        <Panel id="refuge" active={tab === 'refuge'}><RefugeSection campaignId={campaignId} isGm={isGm} /></Panel>
        {isGm && <Panel id="effects" active={tab === 'effects'}><PendingEffectsView pending={pending} /></Panel>}
        {isGm && <Panel id="items" active={tab === 'items'}><ItemsPanel campaignId={campaignId} /></Panel>}
      </div>
      <div className="flex w-14 shrink-0 flex-col items-center gap-1 border-l border-zinc-800/60 py-2">
        <button type="button" aria-label={open ? 'Recolher barra lateral' : 'Expandir barra lateral'} aria-expanded={open}
          onClick={() => setOpen(!open)} className={`group relative ${ICON_BTN}`}>
          {open ? <PanelRightClose size={20} strokeWidth={1.5} aria-hidden /> : <PanelRightOpen size={20} strokeWidth={1.5} aria-hidden />}
          <Tip side="left">{open ? 'Recolher' : 'Expandir'}</Tip>
        </button>
        <div role="tablist" aria-label="Painéis" aria-orientation="vertical" className="flex flex-col gap-1">
          {tabsFor(isGm).map((d) => <TabButton key={d.id} def={d} active={open && tab === d.id} badge={badge(d.id)} onClick={() => choose(d.id)} />)}
        </div>
      </div>
    </aside>
  )
}
