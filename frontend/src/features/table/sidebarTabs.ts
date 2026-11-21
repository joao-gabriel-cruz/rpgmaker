import { Dices, Home, Swords, UsersRound, Zap, type LucideIcon } from 'lucide-react'

export type SidebarTabId = 'group' | 'rolls' | 'conflict' | 'refuge' | 'effects'

export interface SidebarTabDef {
  id: SidebarTabId
  label: string
  icon: LucideIcon
  gmOnly: boolean
}

export const SIDEBAR_TABS: SidebarTabDef[] = [
  { id: 'group', label: 'Grupo', icon: UsersRound, gmOnly: false },
  { id: 'rolls', label: 'Rolagens', icon: Dices, gmOnly: false },
  { id: 'conflict', label: 'Conflito', icon: Swords, gmOnly: false },
  { id: 'refuge', label: 'Refúgio', icon: Home, gmOnly: false },
  { id: 'effects', label: 'Efeitos', icon: Zap, gmOnly: true },
]

export const tabsFor = (isGm: boolean) => SIDEBAR_TABS.filter((t) => isGm || !t.gmOnly)

export interface SidebarState {
  open: boolean
  tab: SidebarTabId
  setOpen: (open: boolean) => void
  setTab: (tab: SidebarTabId) => void
}
