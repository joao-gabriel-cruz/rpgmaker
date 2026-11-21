import { useEffect } from 'react'
import { toolByKey, type TableToolId } from './tools'

export interface ShortcutActions {
  isGm: boolean
  setTool: (t: TableToolId) => void
  setSidebarOpen: (open: boolean) => void
  openHelp: () => void
  /** Space: open the roll composer. */
  openRoll?: () => void
  /** Enter: repeat the last roll. */
  rollAgain?: () => void
  /** ' (apostrophe): speak in character. */
  speak?: () => void
  /** G: open the reaction wheel. */
  react?: () => void
}

const isTyping = (t: EventTarget | null) =>
  t instanceof HTMLElement && (t.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(t.tagName))

// Space/Enter keep their native meaning on buttons, links, tabs, dialogs and the sheet drawer.
const isActivatable = (t: EventTarget | null) =>
  t instanceof HTMLElement && !!t.closest('button, a, summary, select, [role=tab], [role=dialog], [role=menuitem], aside[aria-label="Ficha do personagem"]')

/** Applies one key press; returns true when it was a table shortcut. */
export function runShortcut(key: string, a: ShortcutActions): boolean {
  const tool = toolByKey(key.toLowerCase(), a.isGm)
  if (tool) a.setTool(tool.id)
  else if (key === 'Escape') a.setTool('move')
  else if (key === '[') a.setSidebarOpen(false)
  else if (key === ']') a.setSidebarOpen(true)
  else if (key === '?') a.openHelp()
  else if (key === ' ' && a.openRoll) a.openRoll()
  else if (key === 'Enter' && a.rollAgain) a.rollAgain()
  else if (key === "'" && a.speak) a.speak()
  else if (key.toLowerCase() === 'g' && a.react) a.react()
  else return false
  return true
}

/** Global one-key shortcuts; ignored while typing in a field or with Ctrl/Cmd/Alt held. */
export function useTableShortcuts(actions: ShortcutActions) {
  const { isGm, setTool, setSidebarOpen, openHelp, openRoll, rollAgain, speak, react } = actions
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (isTyping(e.target) || e.metaKey || e.ctrlKey || e.altKey) return
      if (e.key.length === 1 && e.shiftKey && e.key !== '?') return
      const activation = e.key === ' ' || e.key === 'Enter'
      if (activation && isActivatable(e.target)) return
      if (runShortcut(e.key, { isGm, setTool, setSidebarOpen, openHelp, openRoll, rollAgain, speak, react }) && (activation || e.key === "'")) e.preventDefault()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [isGm, setTool, setSidebarOpen, openHelp, openRoll, rollAgain, speak, react])
}
