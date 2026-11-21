import { useThree } from '@react-three/fiber'
import { useEffect, useRef, useState } from 'react'
import { Vector2, type Camera } from 'three'
import type { Coord, MapGrid } from '../map3d/grid'
import { pickCell } from '../tokens/pickCell'
import { rectCells } from './bits'
import type { Brush } from './useVision'

interface Props {
  grid: MapGrid
  brush: Brush
  onApply: (cells: Coord[], reveal: boolean) => void
}

interface Drag {
  from: Coord
  to: Coord
}

type Ndc = (e: PointerEvent) => Vector2

// Pointer handlers of one drag session; `cur` lives in the closure so re-renders never reset it.
function dragHandlers(grid: MapGrid, ndc: Ndc, camera: Camera, set: (d: Drag | null) => void, done: (d: Drag) => void, lock: (on: boolean) => void) {
  let cur: Drag | null = null
  return {
    down: (e: PointerEvent) => {
      const c = e.button === 0 ? pickCell(grid, ndc(e), camera) : null
      if (!c) return
      e.stopPropagation()
      lock(true)
      set((cur = { from: c, to: c }))
    },
    move: (e: PointerEvent) => {
      const c = cur && pickCell(grid, ndc(e), camera)
      if (cur && c) set((cur = { from: cur.from, to: c }))
    },
    up: () => {
      if (cur) done(cur)
      cur = null
      set(null)
      lock(false)
    },
  }
}

// Capturing on the canvas keeps the terrain editor and token drags from also reacting.
function useRectDrag(grid: MapGrid, active: boolean, onDone: (d: Drag) => void) {
  const dom = useThree((s) => s.gl.domElement)
  const camera = useThree((s) => s.camera)
  const controls = useThree((s) => s.controls) as { enabled: boolean } | null
  const [drag, setDrag] = useState<Drag | null>(null)
  const done = useRef(onDone)
  useEffect(() => { done.current = onDone })
  useEffect(() => {
    if (!active) return
    const ndc: Ndc = (e) => {
      const r = dom.getBoundingClientRect()
      return new Vector2(((e.clientX - r.left) / r.width) * 2 - 1, -((e.clientY - r.top) / r.height) * 2 + 1)
    }
    const lock = (on: boolean) => { if (controls) controls.enabled = !on }
    const h = dragHandlers(grid, ndc, camera, setDrag, (d) => done.current(d), lock)
    dom.addEventListener('pointerdown', h.down)
    window.addEventListener('pointermove', h.move)
    window.addEventListener('pointerup', h.up)
    return () => {
      dom.removeEventListener('pointerdown', h.down)
      window.removeEventListener('pointermove', h.move)
      window.removeEventListener('pointerup', h.up)
      lock(false)
    }
  }, [dom, camera, controls, grid, active])
  return drag
}

/** Rectangular reveal/hide brush for the Mestre: drag a rectangle (a click is one cell). */
export function FogBrush({ grid, brush, onApply }: Props) {
  const drag = useRectDrag(grid, brush !== null, (d) => brush && onApply(rectCells(d.from, d.to), brush === 'reveal'))
  if (!drag) return null
  const w = Math.abs(drag.to.x - drag.from.x) + 1
  const d = Math.abs(drag.to.z - drag.from.z) + 1
  return (
    <mesh position={[Math.min(drag.from.x, drag.to.x) + w / 2, 10.5, Math.min(drag.from.z, drag.to.z) + d / 2]} raycast={() => null}>
      <boxGeometry args={[w, 0.05, d]} />
      <meshBasicMaterial color={brush === 'reveal' ? '#34d399' : '#f87171'} transparent opacity={0.5} />
    </mesh>
  )
}
