import { useEffect } from 'react'
import { toolByKey, type TableToolId } from './tools'

export interface ShortcutActions {
  isGm: boolean
  setTool: (t: TableToolId) => void
  setSidebarOpen: (open: boolean) => void
  openHelp: () => void
}

const isTyping = (t: EventTarget | null) =>
  t instanceof HTMLElement && (t.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(t.tagName))

/** Applies one key press; returns true when it was a table shortcut. */
export function runShortcut(key: string, a: ShortcutActions): boolean {
  const tool = toolByKey(key.toLowerCase(), a.isGm)
  if (tool) a.setTool(tool.id)
  else if (key === 'Escape') a.setTool('move')
  else if (key === '[') a.setSidebarOpen(false)
  else if (key === ']') a.setSidebarOpen(true)
  else if (key === '?') a.openHelp()
  else return false
  return true
}

/** Global one-key shortcuts; ignored while typing in a field or with Ctrl/Cmd/Alt held. */
export function useTableShortcuts(actions: ShortcutActions) {
  const { isGm, setTool, setSidebarOpen, openHelp } = actions
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (isTyping(e.target) || e.metaKey || e.ctrlKey || e.altKey) return
      if (e.key.length === 1 && e.shiftKey && e.key !== '?') return
      runShortcut(e.key, { isGm, setTool, setSidebarOpen, openHelp })
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [isGm, setTool, setSidebarOpen, openHelp])
}
