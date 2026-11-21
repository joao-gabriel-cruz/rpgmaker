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
  const { setTool, setCameraHost } = useTableTool()
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
        <main aria-label="Mesa" className="@container relative min-w-0 flex-1 bg-zinc-900">
          {map}
          <div data-slot="overlay" className="pointer-events-none absolute inset-x-3 bottom-40 top-3 z-20 flex items-center justify-center">{overlay}</div>
          <Zones drawer={drawer} presence={presence} hotbar={hotbar} setCameraHost={setCameraHost} />
        </main>
        <Sidebar campaignId={campaign.id} isGm={isGm} sidebar={sidebar} />
      </div>
      <div data-slot="toasts">{toasts}</div>
      {dialog === 'settings' && isGm && <SettingsDialog campaign={campaign} onClose={closeDialog} />}
      {dialog === 'help' && <ShortcutsDialog isGm={isGm} onClose={closeDialog} />}
    </div>
  )
}

interface ZonesProps extends Pick<TableLayoutProps, 'drawer' | 'presence' | 'hotbar'> {
  setCameraHost: (el: HTMLElement | null) => void
}

/**
 * Floating zones over the map, laid out in flow so they never overlap: the options panel (left) and
 * the sheet drawer (right) share the top row; dock, hotbar and camera share the bottom row.
 */
function Zones({ drawer, presence, hotbar, setCameraHost }: ZonesProps) {
  return (
    <div data-zones className="pointer-events-none absolute inset-3 z-10 flex flex-col gap-3">
      <div data-zone="top" className="flex min-h-0 flex-1 items-start gap-3">
        <OptionsPanel />
        <div data-slot="drawer" className="flex h-full min-h-0 min-w-0 flex-1 justify-end">{drawer}</div>
      </div>
      <div data-zone="bottom" className="grid shrink-0 grid-cols-[auto_minmax(0,1fr)_auto] items-end gap-3">
        <div data-slot="presence" className="min-w-0">{presence}</div>
        <div data-slot="hotbar" className="flex min-w-0 justify-center">{hotbar}</div>
        <div data-slot="camera" ref={setCameraHost} className="flex items-end" />
      </div>
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
