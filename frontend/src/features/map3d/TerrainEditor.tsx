import { useEffect, useState } from 'react'
import { HoverCursor } from './HoverCursor'
import type { MapGrid, MapPatch } from './grid'
import { TerrainScene } from './TerrainScene'
import { Toolbar } from './Toolbar'
import { usePainting } from './usePainting'
import { BRUSH_SIZES, DEFAULT_TOOL_STATE, TOOLS, type ToolState } from './tools'

interface Props {
  grid: MapGrid
  onPatch: (patch: MapPatch) => void
}

function useShortcuts(setState: (f: (s: ToolState) => ToolState) => void) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.metaKey || e.ctrlKey) return
      const tool = TOOLS.find((t) => t.key === e.key.toLowerCase())
      if (tool) return setState((s) => ({ ...s, tool: tool.id }))
      const dir = e.key === ']' ? 1 : e.key === '[' ? -1 : 0
      if (!dir) return
      setState((s) => {
        const i = BRUSH_SIZES.indexOf(s.brushSize) + dir
        return { ...s, brushSize: BRUSH_SIZES[Math.min(2, Math.max(0, i))] }
      })
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [setState])
}

export function TerrainEditor({ grid, onPatch }: Props) {
  const [state, setState] = useState<ToolState>(DEFAULT_TOOL_STATE)
  const [showGrid, setShowGrid] = useState(true)
  const { hover, ...handlers } = usePainting(grid, state, onPatch)
  useShortcuts(setState)
  const size = state.tool === 'fill' ? 1 : state.brushSize

  return (
    <div className="flex h-full w-full">
      <Toolbar state={state} onChange={setState} showGrid={showGrid} onToggleGrid={() => setShowGrid((v) => !v)} />
      <div className="relative flex-1">
        <TerrainScene grid={grid} showGrid={showGrid} editable {...handlers}>
          {hover && <HoverCursor grid={grid} at={hover} size={size} />}
        </TerrainScene>
      </div>
    </div>
  )
}
