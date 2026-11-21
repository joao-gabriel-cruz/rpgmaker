import { useEffect } from 'react'
import { nextRotation } from './objects/model'
import { BRUSH_SIZES, TOOLS, type ToolState } from './tools'

type SetState = (f: (s: ToolState) => ToolState) => void

export type ShortcutMode = 'standalone' | 'external' | 'off'

function brushStep(dir: number, setState: SetState) {
  setState((s) => {
    const i = BRUSH_SIZES.indexOf(s.brushSize) + dir
    return { ...s, brushSize: BRUSH_SIZES[Math.min(2, Math.max(0, i))] }
  })
}

// External mode: the table owns the letters, so the editor only keeps Shift+R and -/+.
function handleExternal(e: KeyboardEvent, setState: SetState) {
  const key = e.key.toLowerCase()
  if (e.shiftKey && key === 'r') return setState((s) => ({ ...s, rotation: nextRotation(s.rotation) }))
  if (key === '-') brushStep(-1, setState)
  if (key === '+' || key === '=') brushStep(1, setState)
}

function handleKey(key: string, setState: SetState) {
  if (key === 'r') return setState((s) => ({ ...s, rotation: nextRotation(s.rotation) }))
  const tool = TOOLS.find((t) => t.key === key)
  if (tool) return setState((s) => ({ ...s, tool: tool.id }))
  const dir = key === ']' ? 1 : key === '[' ? -1 : 0
  if (dir) brushStep(dir, setState)
}

export function useShortcuts(setState: SetState, mode: ShortcutMode = 'standalone') {
  useEffect(() => {
    if (mode === 'off') return
    const onKey = (e: KeyboardEvent) => {
      const t = e.target
      if (t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement || e.metaKey || e.ctrlKey || e.altKey) return
      if (mode === 'external') handleExternal(e, setState)
      else handleKey(e.key.toLowerCase(), setState)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [setState, mode])
}
