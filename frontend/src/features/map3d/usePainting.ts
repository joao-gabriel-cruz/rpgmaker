import { useEffect, useRef, useState } from 'react'
import { applyTool } from './applyTool'
import { inBounds, type Coord, type MapGrid, type MapPatch } from './grid'
import type { ToolState } from './tools'

function tryVisit(visited: Set<string>, g: MapGrid, at: Coord) {
  const key = `${at.x},${at.z}`
  if (!inBounds(g, at.x, at.z) || visited.has(key)) return false
  visited.add(key)
  return true
}

// Drag-paint session: each cell is touched once per stroke so raise/lower don't repeat.
export function usePainting(grid: MapGrid, state: ToolState, onPatch: (p: MapPatch) => void) {
  const [hover, setHover] = useState<Coord | null>(null)
  const painting = useRef(false)
  const visited = useRef(new Set<string>())
  const latest = useRef({ grid, state })
  useEffect(() => {
    latest.current = { grid, state }
  })

  const apply = (at: Coord) => {
    const { grid: g, state: s } = latest.current
    const patch = tryVisit(visited.current, g, at) && applyTool(g, at, s)
    if (patch && patch.cells.length) onPatch(patch)
  }

  return {
    hover,
    onCellDown: (at: Coord) => {
      painting.current = true
      visited.current.clear()
      apply(at)
    },
    onCellMove: (at: Coord) => {
      setHover(at)
      if (painting.current && latest.current.state.tool !== 'fill') apply(at)
    },
    onUp: () => void (painting.current = false),
  }
}
