import { useLayoutEffect, useMemo, useRef } from 'react'
import { Object3D, type InstancedMesh } from 'three'
import type { MapGrid } from '../map3d/grid'
import { fogCover, fogInstances, type VisionState } from './vision'

interface Props {
  grid: MapGrid
  state: VisionState
  isGm: boolean
}

const dummy = new Object3D()

/** Covers the cells the viewer cannot see: opaque for players, a subtle film for the Mestre. */
export function FogLayer({ grid, state, isGm }: Props) {
  const cover = useMemo(() => fogCover(state, grid, isGm), [state, grid, isGm])
  const items = useMemo(() => (cover ? fogInstances(grid, cover.cells, cover.gm) : []), [cover, grid])
  const ref = useRef<InstancedMesh>(null)
  useLayoutEffect(() => {
    const mesh = ref.current
    if (!mesh) return
    items.forEach((f, i) => {
      dummy.position.set(f.x, f.y, f.z)
      dummy.scale.set(1, f.h, 1)
      dummy.updateMatrix()
      mesh.setMatrixAt(i, dummy.matrix)
    })
    mesh.instanceMatrix.needsUpdate = true
  }, [items])
  if (!cover || items.length === 0) return null
  return (
    <instancedMesh key={items.length + (cover.gm ? 'g' : 'p')} ref={ref} args={[undefined, undefined, items.length]} raycast={() => null} renderOrder={5}>
      <boxGeometry args={[1, 1, 1]} />
      {cover.gm
        ? <meshBasicMaterial color="#000000" transparent opacity={0.4} depthWrite={false} />
        : <meshBasicMaterial color="#050507" />}
    </instancedMesh>
  )
}
