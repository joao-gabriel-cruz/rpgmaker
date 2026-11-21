import { useCallback, useState } from 'react'
import type { SidebarState, SidebarTabId } from './sidebarTabs'

/** Sidebar open/closed + active tab. Picking a tab while collapsed expands it. */
export function useSidebarState(): SidebarState {
  const [open, setOpen] = useState(true)
  const [tab, setTabState] = useState<SidebarTabId>('group')
  const setTab = useCallback((t: SidebarTabId) => {
    setTabState(t)
    setOpen(true)
  }, [])
  return { open, tab, setOpen, setTab }
}
