import type { ThreeEvent } from '@react-three/fiber'
import type { MapGrid } from '../grid'
import { PREFABS } from './catalog'
import type { MapObject } from './model'
import { placement } from './placement'

interface Props {
  obj: MapObject
  grid: MapGrid
  onPick?: (id: string) => void
}

export function ObjectInstance({ obj, grid, onPick }: Props) {
  const Comp = PREFABS[obj.prefab]?.Component
  if (!Comp) return null
  const p = placement(grid, obj)
  const down = onPick && ((e: ThreeEvent<PointerEvent>) => {
    if (e.button !== 0) return
    e.stopPropagation()
    onPick(obj.id)
  })
  return (
    <group position={[p.cx, p.y, p.cz]} rotation-y={(-obj.rotation * Math.PI) / 180} onPointerDown={down}>
      <Comp props={obj.props} />
    </group>
  )
}
