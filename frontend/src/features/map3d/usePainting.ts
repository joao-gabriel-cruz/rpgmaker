import { useEffect, useRef, useState } from 'react'
import { applyTool } from './applyTool'
import { lineCells } from './line'
import { inBounds, type Coord, type MapGrid, type MapPatch } from './grid'
import type { ToolState } from './tools'

function tryVisit(visited: Set<string>, g: MapGrid, at: Coord) {
  const key = `${at.x},${at.z}`
  if (!inBounds(g, at.x, at.z) || visited.has(key)) return false
  visited.add(key)
  return true
}

function useLatest<T>(value: T) {
  const ref = useRef(value)
  useEffect(() => {
    ref.current = value
  })
  return ref
}

// Drag-paint session: each cell is touched once per stroke so raise/lower don't repeat.
function useApply(latest: { current: { grid: MapGrid; state: ToolState } }, onPatch: (p: MapPatch) => void) {
  const visited = useRef(new Set<string>())
  const apply = (at: Coord) => {
    const { grid: g, state: s } = latest.current
    const patch = tryVisit(visited.current, g, at) && applyTool(g, at, s)
    if (patch && patch.cells.length) onPatch(patch)
  }
  return { apply, visited }
}

export function usePainting(grid: MapGrid, state: ToolState, onPatch: (p: MapPatch) => void) {
  const [hover, setHover] = useState<Coord | null>(null)
  const painting = useRef(false)
  const last = useRef<Coord | null>(null)
  const latest = useLatest({ grid, state })
  const { apply, visited } = useApply(latest, onPatch)

  return {
    hover,
    onCellDown: (at: Coord) => {
      painting.current = true
      last.current = at
      visited.current.clear()
      apply(at)
    },
    onCellMove: (at: Coord) => {
      setHover(at)
      if (!painting.current || latest.current.state.tool === 'fill') return
      // Fast drags skip cells; fill the gap between the previous and current point.
      lineCells(last.current ?? at, at).forEach(apply)
      last.current = at
    },
    onUp: () => {
      painting.current = false
      last.current = null
    },
  }
}
