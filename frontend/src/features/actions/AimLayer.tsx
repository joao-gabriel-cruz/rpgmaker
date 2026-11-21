import { useThree } from '@react-three/fiber'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { EdgesGeometry, BoxGeometry, Vector2, Vector3, type Camera } from 'three'
import { getCell, surfaceHeight, type Coord, type MapGrid } from '../map3d/grid'
import { occupiedCells, type MapObject } from '../map3d/objects/model'
import { pickCell } from '../tokens/pickCell'
import type { Token } from '../tokens/tokens'
import { useActionFlowOptional } from './ActionFlow'
import type { ActionTarget } from './actions'
import { resolveAim, type AimResult } from './aim'

interface Props {
  grid: MapGrid
  objects: readonly MapObject[]
  tokens: readonly Token[]
}

function ndcOf(dom: HTMLElement, e: PointerEvent) {
  const r = dom.getBoundingClientRect()
  return new Vector2(((e.clientX - r.left) / r.width) * 2 - 1, -((e.clientY - r.top) / r.height) * 2 + 1)
}

type Hover = { cell: Coord; aim: AimResult }

/** Things the pointer may snap to: figures and objects are taller than the ground plane the cursor ray hits first. */
export interface Snap {
  cell: Coord
  height: number
}

const SNAP_PX = 46

/** The snap point whose screen position is closest to the pointer (within SNAP_PX), if any. */
export function nearestSnap(snaps: readonly Snap[], e: PointerEvent, dom: HTMLElement, camera: Camera): Coord | null {
  const r = dom.getBoundingClientRect()
  let best: { cell: Coord; d: number } | null = null
  for (const s of snaps) {
    const p = new Vector3(s.cell.x + 0.5, s.height, s.cell.z + 0.5).project(camera)
    const d = Math.hypot(r.left + ((p.x + 1) / 2) * r.width - e.clientX, r.top + ((1 - p.y) / 2) * r.height - e.clientY)
    if (d < SNAP_PX && (!best || d < best.d)) best = { cell: s.cell, d }
  }
  return best?.cell ?? null
}

function snapsFor(kind: string, grid: MapGrid, objects: readonly MapObject[], tokens: readonly Token[]): Snap[] {
  const top = (c: Coord) => surfaceHeight(getCell(grid, c.x, c.z)?.height ?? 1)
  if (kind === 'token') return tokens.map((t) => ({ cell: { x: t.x, z: t.z }, height: top(t) + 0.55 }))
  if (kind !== 'object') return []
  return objects.map((o) => ({ cell: { x: o.x, z: o.z }, height: top(o) + 0.5 })).concat(
    objects.flatMap((o) => occupiedCells(o).map((c) => ({ cell: c, height: top(c) + 0.5 }))))
}

function useCrosshair(dom: HTMLElement, on: boolean) {
  useEffect(() => {
    if (!on) return
    const prev = dom.style.cursor
    dom.style.cursor = 'crosshair'
    return () => { dom.style.cursor = prev }
  }, [dom, on])
}

// While aiming: hover resolution, click to pick (capture keeps the editor from acting), right click cancels.
function useAimPointer(grid: MapGrid, resolve: ((cell: Coord) => AimResult) | null, snaps: readonly Snap[], onHover: (h: Hover | null) => void,
  onPick: (t: ActionTarget) => void, onCancel: () => void) {
  const dom = useThree((s) => s.gl.domElement)
  const camera = useThree((s) => s.camera)
  useCrosshair(dom, !!resolve)
  useEffect(() => {
    if (!resolve) return
    const at = (e: PointerEvent): Hover | null => {
      const cell = nearestSnap(snaps, e, dom, camera) ?? pickCell(grid, ndcOf(dom, e), camera)
      return cell && { cell, aim: resolve(cell) }
    }
    const move = (e: PointerEvent) => onHover(at(e))
    const down = (e: PointerEvent) => {
      if (e.button !== 0) return
      e.stopPropagation()
      const hit = at(e)
      if (hit?.aim.valid && hit.aim.target) onPick(hit.aim.target)
    }
    const menu = (e: Event) => { e.preventDefault(); onCancel() }
    dom.addEventListener('pointermove', move)
    dom.addEventListener('pointerdown', down, true)
    dom.addEventListener('contextmenu', menu)
    return () => {
      dom.removeEventListener('pointermove', move)
      dom.removeEventListener('pointerdown', down, true)
      dom.removeEventListener('contextmenu', menu)
      onHover(null)
    }
  }, [dom, camera, grid, resolve, snaps, onHover, onPick, onCancel])
}

function useAim(grid: MapGrid, objects: readonly MapObject[], tokens: readonly Token[]) {
  const flow = useActionFlowOptional()
  const [hover, setHover] = useState<Hover | null>(null)
  const pending = flow?.pending
  const aiming = pending && !pending.target ? pending : null
  const actor = aiming ? tokens.find((t) => t.characterId === aiming.character.id) : undefined
  const resolve = useMemo(
    () => (aiming && actor ? (cell: Coord) => resolveAim(aiming.action, cell, actor, tokens, objects) : null),
    [aiming, actor, tokens, objects],
  )
  const noop = useCallback(() => undefined, [])
  const snaps = useMemo(() => (aiming ? snapsFor(aiming.action.target, grid, objects, tokens) : []), [aiming, grid, objects, tokens])
  useAimPointer(grid, resolve, snaps, setHover, flow?.setTarget ?? noop, flow?.cancel ?? noop)
  const reason = hover?.aim.valid === false ? hover.aim.reason ?? '' : ''
  const setNote = flow?.setNote
  useEffect(() => setNote?.(reason), [reason, setNote])
  return { aiming, actor, hover }
}

const frame = new EdgesGeometry(new BoxGeometry(1, 1, 1))

function RangeFrame({ grid, at, range }: { grid: MapGrid; at: Coord; range: number }) {
  const y = surfaceHeight(getCell(grid, at.x, at.z)?.height ?? 1) + 0.05
  const size = range * 2 + 1
  return (
    <lineSegments geometry={frame} position={[at.x + 0.5, y, at.z + 0.5]} scale={[size, 0.02, size]} raycast={() => null}>
      <lineBasicMaterial color="#34d399" transparent opacity={0.6} />
    </lineSegments>
  )
}

/** In-canvas aiming: a crosshair cursor, the range frame and a highlight on the cell under the mouse (green valid, red not). */
export function AimLayer({ grid, objects, tokens }: Props) {
  const { aiming, actor, hover } = useAim(grid, objects, tokens)
  const marker = useMemo(() => (hover ? surfaceHeight(getCell(grid, hover.cell.x, hover.cell.z)?.height ?? 1) : 0), [grid, hover])
  if (!aiming || !actor) return null
  return (
    <>
      {aiming.action.range ? <RangeFrame grid={grid} at={actor} range={aiming.action.range} /> : null}
      {hover && (
        <mesh position={[hover.cell.x + 0.5, marker + 0.07, hover.cell.z + 0.5]} raycast={() => null}>
          <boxGeometry args={[1, 0.1, 1]} />
          <meshBasicMaterial color={hover.aim.valid ? '#34d399' : '#f87171'} transparent opacity={0.55} />
        </mesh>
      )}
    </>
  )
}
