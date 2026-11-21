import { useThree, type ThreeEvent } from '@react-three/fiber'
import { useCallback, useEffect, useState } from 'react'
import { Vector2 } from 'three'
import type { Coord, MapGrid } from '../map3d/grid'
import { standHeight } from '../map3d/objects/bridge'
import type { MapObject } from '../map3d/objects/model'
import { useHold, type HoldPoint } from '../ambience/useTokenHold'
import { CameraRig } from './CameraRig'
import { PingMarkers } from './PingMarkers'
import { pickCell, pickPoint } from './pickCell'
import { TokenFigure } from './TokenFigure'
import { canMoveToken, type Token } from './tokens'
import type { SceneTokens } from './useSceneTokens'

interface Props {
  grid: MapGrid
  /** Scene objects: a bridge lifts the token to its deck. */
  objects?: readonly MapObject[]
  tokens: SceneTokens
  userId: string
  isGm: boolean
  /** The Ping tool is active: a plain click pings instead of dragging tokens. */
  pingTool?: boolean
  /** Tokens can be dragged (false while painting or measuring, so gestures never collide). */
  dragEnabled?: boolean
  /** A character armed in the Tokens tool: the next click on a map cell places its token there. */
  onPlaceAt?: ((c: Coord) => void) | null
  /** A token was clicked with the Move/Tokens tool: the table selects its character. */
  onSelect?: (characterId: string) => void
  /** The viewer held the pointer on a token they may move (opens the reaction wheel). */
  onHold?: (t: Token, at: HoldPoint) => void
}

interface Drag {
  id: string
  at: Coord
}

function useNdc() {
  const dom = useThree((s) => s.gl.domElement)
  return useCallback((e: { clientX: number; clientY: number }) => {
    const r = dom.getBoundingClientRect()
    return new Vector2(((e.clientX - r.left) / r.width) * 2 - 1, -((e.clientY - r.top) / r.height) * 2 + 1)
  }, [dom])
}

function useDrag(grid: MapGrid, onDrop: (id: string, c: Coord) => void) {
  const camera = useThree((s) => s.camera)
  const controls = useThree((s) => s.controls) as { enabled: boolean } | null
  const ndc = useNdc()
  const [drag, setDrag] = useState<Drag | null>(null)
  const id = drag?.id
  useEffect(() => {
    if (!id) return
    if (controls) controls.enabled = false
    const move = (e: PointerEvent) => {
      const c = pickCell(grid, ndc(e), camera)
      if (c) setDrag((d) => d && { ...d, at: c })
    }
    const up = () => {
      setDrag((d) => {
        if (d) onDrop(d.id, d.at)
        return null
      })
    }
    window.addEventListener('pointermove', move)
    window.addEventListener('pointerup', up)
    return () => {
      if (controls) controls.enabled = true
      window.removeEventListener('pointermove', move)
      window.removeEventListener('pointerup', up)
    }
  }, [id, grid, camera, controls, ndc, onDrop])
  return { drag, start: (t: Token) => setDrag({ id: t.id, at: { x: t.x, z: t.z } }) }
}

// Alt+click (or any click with the Ping tool) pings; the capture keeps the editor from also painting.
function useAltPing(grid: MapGrid, ping: (x: number, z: number) => void, always: boolean) {
  const dom = useThree((s) => s.gl.domElement)
  const camera = useThree((s) => s.camera)
  const ndc = useNdc()
  useEffect(() => {
    const down = (e: PointerEvent) => {
      if ((!e.altKey && !always) || e.button !== 0) return
      e.stopPropagation()
      const p = pickPoint(grid, ndc(e), camera)
      if (p) ping(p.x, p.z)
    }
    dom.addEventListener('pointerdown', down)
    return () => dom.removeEventListener('pointerdown', down)
  }, [dom, camera, grid, ndc, ping, always])
}

// While armed, a plain click picks the cell for the new token (capture keeps the editor from also acting).
function usePlaceClick(grid: MapGrid, onPlaceAt: ((c: Coord) => void) | null | undefined) {
  const dom = useThree((s) => s.gl.domElement)
  const camera = useThree((s) => s.camera)
  const ndc = useNdc()
  useEffect(() => {
    if (!onPlaceAt) return
    const down = (e: PointerEvent) => {
      if (e.button !== 0 || e.altKey) return
      e.stopPropagation()
      const c = pickCell(grid, ndc(e), camera)
      if (c) onPlaceAt(c)
    }
    const prev = dom.style.cursor
    dom.style.cursor = 'crosshair'
    dom.addEventListener('pointerdown', down, true)
    return () => {
      dom.style.cursor = prev
      dom.removeEventListener('pointerdown', down, true)
    }
  }, [dom, camera, grid, ndc, onPlaceAt])
}

/** In-canvas part of the tokens feature: figures, drag preview, pings and camera moves. */
export function TokenLayer({ grid, objects, tokens, userId, isGm, pingTool = false, dragEnabled = true, onPlaceAt, onSelect, onHold }: Props) {
  const { move, ping } = tokens
  const onDrop = useCallback((id: string, c: Coord) => move(id, c.x, c.z), [move])
  const { drag, start } = useDrag(grid, onDrop)
  useAltPing(grid, ping, pingTool)
  usePlaceClick(grid, onPlaceAt)
  const hold = useHold(onHold)
  const down = (t: Token) => (e: ThreeEvent<PointerEvent>) => {
    if (!dragEnabled || pingTool || onPlaceAt || e.button !== 0 || e.altKey) return
    onSelect?.(t.characterId)
    if (!canMoveToken(t, userId, isGm)) return
    e.stopPropagation()
    start(t)
    hold(t, { x: e.clientX, y: e.clientY })
  }
  return (
    <>
      {tokens.tokens.map((t) => {
        const at = drag?.id === t.id ? drag.at : t
        const y = standHeight(grid, objects, at.x, at.z)
        return <TokenFigure key={t.id} token={t} x={at.x} z={at.z} y={y} dragging={drag?.id === t.id} onPointerDown={down(t)} />
      })}
      {drag && <DropMarker grid={grid} objects={objects} at={drag.at} />}
      <PingMarkers grid={grid} pings={tokens.pings} />
      <CameraRig target={tokens.camera} view={tokens.view} />
    </>
  )
}

function DropMarker({ grid, objects, at }: { grid: MapGrid; objects?: readonly MapObject[]; at: Coord }) {
  const y = standHeight(grid, objects, at.x, at.z)
  return (
    <mesh position={[at.x + 0.5, y + 0.06, at.z + 0.5]} raycast={() => null}>
      <boxGeometry args={[1, 0.08, 1]} />
      <meshBasicMaterial color="#34d399" transparent opacity={0.5} />
    </mesh>
  )
}
