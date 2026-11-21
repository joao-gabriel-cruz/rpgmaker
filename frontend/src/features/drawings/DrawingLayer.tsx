import { Html, Line } from '@react-three/drei'
import { useThree } from '@react-three/fiber'
import { useEffect, useState } from 'react'
import { Vector2, type Camera } from 'three'
import { inBounds, type MapGrid } from '../map3d/grid'
import { pickCell, pickPoint } from '../tokens/pickCell'
import { formatMeasure, measure, projectStroke, type Point } from './drawings'
import type { SceneDrawings } from './useDrawings'

const RULER_COLOR = '#fbbf24'
const center = (c: Point): Point => ({ x: c.x + 0.5, z: c.z + 0.5 })
const at = (r: { from: Point; to: Point }): Active => ({ from: r.from, to: r.to })

function Ruler({ grid, from, to, label }: { grid: MapGrid; from: Point; to: Point; label: string }) {
  const pts = projectStroke(grid, [center(from), center(to)], 0.15)
  const mid = pts[Math.floor(pts.length / 2)]
  return (
    <>
      {pts.length > 1 && <Line points={pts} color={RULER_COLOR} lineWidth={3} dashed dashSize={0.3} gapSize={0.15} raycast={() => null} />}
      <Html position={[mid[0], mid[1] + 0.5, mid[2]]} center style={{ pointerEvents: 'none' }}>
        <span className="whitespace-nowrap rounded bg-zinc-900/90 px-2 py-0.5 text-xs text-amber-300">{label}</span>
      </Html>
    </>
  )
}

interface Active {
  from: Point
  to: Point
}

interface Gesture {
  grid: MapGrid
  draw: SceneDrawings
  ndc: (e: PointerEvent) => Vector2
  camera: Camera
  onRuler: (a: Active | null) => void
  onStroke: (p: Point[] | null) => void
}

function trackRuler(g: Gesture, from: Point) {
  let ruler: Active = { from, to: from }
  g.onRuler(ruler)
  return {
    move: (e: PointerEvent) => {
      const c = pickCell(g.grid, g.ndc(e), g.camera)
      if (!c || (c.x === ruler.to.x && c.z === ruler.to.z)) return
      ruler = { from, to: c }
      g.onRuler(ruler)
      g.draw.shareRuler(ruler.from, ruler.to)
    },
    end: () => {
      g.draw.shareRuler(null, null)
      g.onRuler(null)
    },
  }
}

function trackStroke(g: Gesture, first: Point) {
  let stroke: Point[] = [first]
  return {
    move: (e: PointerEvent) => {
      const p = pickPoint(g.grid, g.ndc(e), g.camera)
      if (p && inBounds(g.grid, Math.floor(p.x), Math.floor(p.z))) g.onStroke((stroke = [...stroke, p]))
    },
    end: () => {
      g.draw.add(stroke)
      g.onStroke(null)
    },
  }
}

// Returns a cleanup-free starter: drags are tracked on window until pointerup.
function startGesture(g: Gesture, e: PointerEvent, controls: { enabled: boolean } | null) {
  const ruler = g.draw.mode === 'ruler'
  const first = ruler ? pickCell(g.grid, g.ndc(e), g.camera) : pickPoint(g.grid, g.ndc(e), g.camera)
  if (!first) return
  if (controls) controls.enabled = false
  const t = ruler ? trackRuler(g, first) : trackStroke(g, first)
  const up = () => {
    window.removeEventListener('pointermove', t.move)
    window.removeEventListener('pointerup', up)
    if (controls) controls.enabled = true
    t.end()
  }
  window.addEventListener('pointermove', t.move)
  window.addEventListener('pointerup', up)
}

// Capture on the canvas so the editor and token drags never see the pointer while a tool is on.
function useToolPointer(grid: MapGrid, draw: SceneDrawings, onRuler: Gesture['onRuler'], onStroke: Gesture['onStroke']) {
  const dom = useThree((s) => s.gl.domElement)
  const camera = useThree((s) => s.camera)
  const controls = useThree((s) => s.controls) as { enabled: boolean } | null
  useEffect(() => {
    if (draw.mode === 'none') return
    const ndc = (e: PointerEvent) => {
      const r = dom.getBoundingClientRect()
      return new Vector2(((e.clientX - r.left) / r.width) * 2 - 1, -((e.clientY - r.top) / r.height) * 2 + 1)
    }
    const down = (e: PointerEvent) => {
      if (e.button !== 0) return
      e.stopPropagation()
      startGesture({ grid, draw, ndc, camera, onRuler, onStroke }, e, controls)
    }
    dom.addEventListener('pointerdown', down)
    return () => dom.removeEventListener('pointerdown', down)
  }, [dom, camera, controls, grid, draw, onRuler, onStroke])
}

/** In-canvas part of the drawings feature: strokes, the live preview, and the local and remote rulers. */
export function DrawingLayer({ grid, draw }: { grid: MapGrid; draw: SceneDrawings }) {
  const [ruler, setRuler] = useState<Active | null>(null)
  const [preview, setPreview] = useState<Point[] | null>(null)
  useToolPointer(grid, draw, setRuler, setPreview)
  const label = (a: Active) => formatMeasure(measure(grid, a.from, a.to))
  return (
    <>
      {draw.drawings.map((d) => (
        <Line key={d.id} points={projectStroke(grid, d.points)} color={d.color} lineWidth={d.width} raycast={() => null} />
      ))}
      {preview && preview.length > 1 && (
        <Line points={projectStroke(grid, preview)} color={draw.color} lineWidth={draw.width} raycast={() => null} />
      )}
      {ruler && <Ruler grid={grid} from={ruler.from} to={ruler.to} label={label(ruler)} />}
      {draw.rulers.map((r) => (
        <Ruler key={r.userId} grid={grid} from={r.from} to={r.to} label={`${r.userName}: ${label(at(r))}`} />
      ))}
    </>
  )
}
