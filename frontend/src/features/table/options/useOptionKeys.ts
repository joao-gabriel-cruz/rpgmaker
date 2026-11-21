import { useEffect } from 'react'
import type { ToolState } from '../../map3d/tools'
import type { TableToolId } from '../tools'
import { OBJECT_MODES, TERRAIN_SUBTOOLS } from './subTools'

const isTyping = (t: EventTarget | null) =>
  t instanceof HTMLElement && (t.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(t.tagName))

/** Sub-tool picked by a number key in the given rail tool, if any. */
export function subToolForKey(tool: TableToolId, key: string): ToolState['tool'] | undefined {
  const list = tool === 'terrain' ? TERRAIN_SUBTOOLS : tool === 'objects' ? OBJECT_MODES : []
  return list.find((o) => o.keys === key)?.id
}

/** 1-6 (Terrain) / 1-3 (Objects) choose the sub-tool; the letters stay with the rail and the camera. */
export function useOptionKeys(tool: TableToolId, editor: ToolState, setEditor: (s: ToolState) => void, enabled: boolean) {
  useEffect(() => {
    if (!enabled) return
    const onKey = (e: KeyboardEvent) => {
      if (isTyping(e.target) || e.metaKey || e.ctrlKey || e.altKey || e.shiftKey) return
      const next = subToolForKey(tool, e.key)
      if (next) setEditor({ ...editor, tool: next })
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [tool, editor, setEditor, enabled])
}
