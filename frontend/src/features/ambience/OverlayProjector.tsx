import { useFrame } from '@react-three/fiber'
import { useEffect, useRef } from 'react'
import { Vector3 } from 'three'
import { getCell, surfaceHeight, type MapGrid } from '../map3d/grid'
import type { Token } from '../tokens/tokens'
import type { OverlayRegistry } from './overlayRegistry'
import { tokenTop } from './tokenTop'

interface Props {
  grid: MapGrid
  tokens: readonly Token[]
  registry: OverlayRegistry
}

const point = new Vector3()

/**
 * Lives in the Canvas: each frame it projects every token's head position to screen pixels and moves
 * that token's overlay node (a plain DOM layer, so React DOM updates never go through a second root).
 */
export function OverlayProjector({ grid, tokens, registry }: Props) {
  const latest = useRef({ grid, tokens })
  useEffect(() => { latest.current = { grid, tokens } })
  useFrame(({ camera, size }) => {
    for (const t of latest.current.tokens) {
      const el = registry.get(t.id)
      if (!el) continue
      const y = surfaceHeight(getCell(latest.current.grid, t.x, t.z)?.height ?? 1) + tokenTop(t)
      point.set(t.x + 0.5, y, t.z + 0.5).project(camera)
      const visible = point.z < 1 && Math.abs(point.x) < 1.2 && Math.abs(point.y) < 1.2
      el.style.visibility = visible ? 'visible' : 'hidden'
      el.style.transform = `translate3d(${((point.x + 1) / 2) * size.width}px, ${((1 - point.y) / 2) * size.height}px, 0)`
    }
  })
  return null
}
