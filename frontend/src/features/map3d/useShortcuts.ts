import { useEffect } from 'react'
import { nextRotation } from './objects/model'
import { BRUSH_SIZES, TOOLS, type ToolState } from './tools'

type SetState = (f: (s: ToolState) => ToolState) => void

function handleKey(key: string, setState: SetState) {
  if (key === 'r') return setState((s) => ({ ...s, rotation: nextRotation(s.rotation) }))
  const tool = TOOLS.find((t) => t.key === key)
  if (tool) return setState((s) => ({ ...s, tool: tool.id }))
  const dir = key === ']' ? 1 : key === '[' ? -1 : 0
  if (!dir) return
  setState((s) => {
    const i = BRUSH_SIZES.indexOf(s.brushSize) + dir
    return { ...s, brushSize: BRUSH_SIZES[Math.min(2, Math.max(0, i))] }
  })
}

export function useShortcuts(setState: SetState) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.metaKey || e.ctrlKey) return
      handleKey(e.key.toLowerCase(), setState)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [setState])
}
