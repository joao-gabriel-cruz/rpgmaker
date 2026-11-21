import { useThree, type ThreeEvent } from '@react-three/fiber'
import { useCallback, useEffect, useState } from 'react'
import { Vector2 } from 'three'
import { getCell, surfaceHeight, type Coord, type MapGrid } from '../map3d/grid'
import { CameraRig } from './CameraRig'
import { PingMarkers } from './PingMarkers'
import { pickCell, pickPoint } from './pickCell'
import { TokenFigure } from './TokenFigure'
import { canMoveToken, type Token } from './tokens'
import type { SceneTokens } from './useSceneTokens'

interface Props {
  grid: MapGrid
  tokens: SceneTokens
  userId: string
  isGm: boolean
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

// Alt+click pings; the capture on the canvas keeps the editor from also painting.
function useAltPing(grid: MapGrid, ping: (x: number, z: number) => void) {
  const dom = useThree((s) => s.gl.domElement)
  const camera = useThree((s) => s.camera)
  const ndc = useNdc()
  useEffect(() => {
    const down = (e: PointerEvent) => {
      if (!e.altKey || e.button !== 0) return
      e.stopPropagation()
      const p = pickPoint(grid, ndc(e), camera)
      if (p) ping(p.x, p.z)
    }
    dom.addEventListener('pointerdown', down)
    return () => dom.removeEventListener('pointerdown', down)
  }, [dom, camera, grid, ndc, ping])
}

/** In-canvas part of the tokens feature: figures, drag preview, pings and camera moves. */
export function TokenLayer({ grid, tokens, userId, isGm }: Props) {
  const { move, ping } = tokens
  const onDrop = useCallback((id: string, c: Coord) => move(id, c.x, c.z), [move])
  const { drag, start } = useDrag(grid, onDrop)
  useAltPing(grid, ping)
  const down = (t: Token) => (e: ThreeEvent<PointerEvent>) => {
    if (e.button !== 0 || e.altKey || !canMoveToken(t, userId, isGm)) return
    e.stopPropagation()
    start(t)
  }
  return (
    <>
      {tokens.tokens.map((t) => {
        const at = drag?.id === t.id ? drag.at : t
        const y = surfaceHeight(getCell(grid, at.x, at.z)?.height ?? 1)
        return <TokenFigure key={t.id} token={t} x={at.x} z={at.z} y={y} dragging={drag?.id === t.id} onPointerDown={down(t)} />
      })}
      {drag && <DropMarker grid={grid} at={drag.at} />}
      <PingMarkers grid={grid} pings={tokens.pings} />
      <CameraRig target={tokens.camera} view={tokens.view} />
    </>
  )
}

function DropMarker({ grid, at }: { grid: MapGrid; at: Coord }) {
  const y = surfaceHeight(getCell(grid, at.x, at.z)?.height ?? 1)
  return (
    <mesh position={[at.x + 0.5, y + 0.06, at.z + 0.5]} raycast={() => null}>
      <boxGeometry args={[1, 0.08, 1]} />
      <meshBasicMaterial color="#34d399" transparent opacity={0.5} />
    </mesh>
  )
}
