import { useEffect } from 'react'

interface Actions {
  undo: () => void
  redo: () => void
}

const isTyping = (t: EventTarget | null) =>
  t instanceof HTMLElement && (t.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(t.tagName))

/** Applies one key press; returns true when it was undo/redo. */
export function runHistoryKey(e: Pick<KeyboardEvent, 'key' | 'ctrlKey' | 'metaKey' | 'shiftKey'>, a: Actions): boolean {
  if (!e.ctrlKey && !e.metaKey) return false
  const key = e.key.toLowerCase()
  if (key === 'z') (e.shiftKey ? a.redo : a.undo)()
  else if (key === 'y') a.redo()
  else return false
  return true
}

/** Ctrl/Cmd+Z undoes, Shift+Ctrl/Cmd+Z (or Ctrl+Y) redoes; ignored while typing in a field. */
export function useHistoryKeys(actions: Actions, enabled: boolean) {
  const { undo, redo } = actions
  useEffect(() => {
    if (!enabled) return
    const onKey = (e: KeyboardEvent) => {
      if (isTyping(e.target)) return
      if (runHistoryKey(e, { undo, redo })) e.preventDefault()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [undo, redo, enabled])
}
