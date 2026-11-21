import { Html } from '@react-three/drei'
import { useThree } from '@react-three/fiber'
import { useEffect } from 'react'
import { Vector2 } from 'three'
import { getCell, surfaceHeight, type MapGrid } from '../map3d/grid'
import type { MapObject } from '../map3d/objects/model'
import { pickCell } from '../tokens/pickCell'
import type { Token } from '../tokens/tokens'
import { LootMarker } from './LootMarker'
import { lootPlacement, usePlacement, type Placement } from './lootPlacement'
import { objectAt, pickerFor, visibleLoot, type SceneLoot } from './loot'

interface Props {
  grid: MapGrid
  objects: readonly MapObject[]
  tokens: readonly Token[]
  loot: SceneLoot
  userId: string
  isGm: boolean
}

// While a placement is armed the next map click places the loot instead of editing or moving.
function usePlacePointer(p: Placement | null, grid: MapGrid, objects: readonly MapObject[], place: SceneLoot['place']) {
  const dom = useThree((s) => s.gl.domElement)
  const camera = useThree((s) => s.camera)
  useEffect(() => {
    if (!p) return
    const down = (e: PointerEvent) => {
      if (e.button !== 0) return
      e.stopPropagation()
      const r = dom.getBoundingClientRect()
      const c = pickCell(grid, new Vector2(((e.clientX - r.left) / r.width) * 2 - 1, -((e.clientY - r.top) / r.height) * 2 + 1), camera)
      if (!c) return
      lootPlacement.cancel()
      place({ x: c.x, z: c.z, objectId: objectAt(objects, c.x, c.z)?.id, itemId: p.itemId, difficulty: p.difficulty }).catch(() => {})
    }
    const key = (e: KeyboardEvent) => e.key === 'Escape' && lootPlacement.cancel()
    dom.addEventListener('pointerdown', down, true)
    window.addEventListener('keydown', key)
    return () => {
      dom.removeEventListener('pointerdown', down, true)
      window.removeEventListener('keydown', key)
    }
  }, [p, dom, camera, grid, objects, place])
}

const position = (grid: MapGrid, x: number, z: number): [number, number, number] =>
  [x + 0.5, surfaceHeight(getCell(grid, x, z)?.height ?? 1) + 0.7, z + 0.5]

/** In-canvas loot: markers (Mestre sees hidden ones ghosted), pickup for adjacent tokens and the "found" flash. */
export function LootLayer({ grid, objects, tokens, loot, userId, isGm }: Props) {
  const placing = usePlacement()
  usePlacePointer(isGm ? placing : null, grid, objects, loot.place)
  return (
    <>
      {visibleLoot(loot.loot, isGm).map((l) => {
        const who = pickerFor(l, tokens, userId)
        return (
          <Html key={l.id} position={position(grid, l.x, l.z)} center zIndexRange={[5, 0]}>
            <LootMarker loot={l} isGm={isGm} canPick={!!who} onPick={() => who && loot.pickup(l.id, who.characterId).catch(() => {})} onRemove={() => loot.remove(l.id).catch(() => {})} />
          </Html>
        )
      })}
      {loot.flashes.map((f) => (
        <Html key={f.id} position={position(grid, f.x, f.z)} center zIndexRange={[6, 0]} style={{ pointerEvents: 'none' }}>
          <span role="status" className="whitespace-nowrap rounded-full bg-amber-300/90 px-3 py-1 text-xs font-semibold text-zinc-900 shadow-[0_0_24px_8px_rgba(252,211,77,0.8)] motion-safe:animate-pulse">
            {f.who} achou {f.itemName}!
          </span>
        </Html>
      ))}
    </>
  )
}
