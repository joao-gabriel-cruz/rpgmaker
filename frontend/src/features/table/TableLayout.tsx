import { useCallback, useState, type ReactNode } from 'react'
import type { Campaign } from '../campaigns/campaigns'
import type { SceneControls } from '../scenes/useSceneControls'
import { OptionsPanel } from './OptionsPanel'
import { SettingsDialog } from './SettingsDialog'
import { ShortcutsDialog } from './ShortcutsDialog'
import { Sidebar } from './Sidebar'
import { ToolRail } from './ToolRail'
import { TopBar } from './TopBar'
import { useTableCharacter } from '../hotbar/TableCharacterContext'
import { useSidebarState } from './useSidebarState'
import { usePresenceState } from '../ambience/PresenceContext'
import { useTableShortcuts } from './useTableShortcuts'
import { TableToolProvider, useTableTool } from './useTableTool'

export interface TableLayoutProps {
  campaign: Campaign
  scenes: SceneControls
  /** The map; fills the whole central area. */
  map: ReactNode
  /** Slot "hotbar": bottom-center over the map. */
  hotbar?: ReactNode
  /** Slot "overlay": centered over the map (dice tray). */
  overlay?: ReactNode
  /** Slot "drawer": side panel over the map (character sheet). */
  drawer?: ReactNode
  /** Slot "presence": bottom-left over the map (speak, react). */
  presence?: ReactNode
  /** Slot "toasts": fixed notices. */
  toasts?: ReactNode
}

type Dialog = 'settings' | 'help' | null

// Space opens the roll composer and Enter repeats the last roll.
function useRollShortcuts() {
  const { setComposerOpen, rollAgain } = useTableCharacter()
  const openRoll = useCallback(() => setComposerOpen(true), [setComposerOpen])
  return { openRoll, rollAgain }
}

// ' speaks and G reacts; both are absent outside the presence provider.
function usePresenceShortcuts() {
  const chat = usePresenceState()
  const { setSpeakOpen, setReactOpen } = chat ?? {}
  const speak = useCallback(() => setSpeakOpen?.(true), [setSpeakOpen])
  const react = useCallback(() => setReactOpen?.(true), [setReactOpen])
  return chat ? { speak, react } : {}
}

function Frame({ campaign, scenes, map, hotbar, drawer, overlay, presence, toasts }: TableLayoutProps) {
  const isGm = campaign.role === 'gm'
  const { setTool } = useTableTool()
  const sidebar = useSidebarState()
  const [dialog, setDialog] = useState<Dialog>(null)
  const closeDialog = useCallback(() => setDialog(null), [])
  const openHelp = useCallback(() => setDialog('help'), [])
  useTableShortcuts({ isGm, setTool, setSidebarOpen: sidebar.setOpen, openHelp, ...useRollShortcuts(), ...usePresenceShortcuts() })
  return (
    <div className="flex h-dvh flex-col overflow-hidden bg-zinc-950 text-zinc-100">
      <TopBar campaign={campaign} scenes={scenes} onOpenHelp={openHelp} onOpenSettings={() => setDialog('settings')} />
      <div className="flex min-h-0 flex-1">
        <ToolRail isGm={isGm} />
        <main aria-label="Mesa" className="relative min-w-0 flex-1 bg-zinc-900">
          {map}
          <OptionsPanel />
          <div data-slot="overlay" className="pointer-events-none absolute inset-0 z-20 flex items-center justify-center">{overlay}</div>
          <div data-slot="drawer" className="pointer-events-none absolute bottom-[6.75rem] right-3 top-3 z-10 flex">{drawer}</div>
          <div data-slot="presence" className="pointer-events-none absolute bottom-[6.25rem] left-3 z-10">{presence}</div>
          <div data-slot="hotbar" className="pointer-events-none absolute inset-x-0 bottom-3 z-10 flex justify-center">{hotbar}</div>
        </main>
        <Sidebar campaignId={campaign.id} isGm={isGm} sidebar={sidebar} />
      </div>
      <div data-slot="toasts">{toasts}</div>
      {dialog === 'settings' && isGm && <SettingsDialog campaign={campaign} onClose={closeDialog} />}
      {dialog === 'help' && <ShortcutsDialog isGm={isGm} onClose={closeDialog} />}
    </div>
  )
}

/** Full-screen table (100dvh, no page scroll): top bar, tool rail, map, sidebar and named slots. */
export function TableLayout(props: TableLayoutProps) {
  return (
    <TableToolProvider isGm={props.campaign.role === 'gm'}>
      <Frame {...props} />
    </TableToolProvider>
  )
}
