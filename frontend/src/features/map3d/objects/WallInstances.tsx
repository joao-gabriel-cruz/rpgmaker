import type { ThreeEvent } from '@react-three/fiber'
import { useLayoutEffect, useMemo, useRef } from 'react'
import { BoxGeometry, Color, InstancedMesh, MeshStandardMaterial, Object3D } from 'three'
import type { MapGrid } from '../grid'
import type { MapObject } from './model'
import { placement } from './placement'
import { C } from './palette'

const geometry = new BoxGeometry(1, 3, 1)
const material = new MeshStandardMaterial({ color: '#ffffff', flatShading: true, roughness: 0.9 })

interface Props {
  walls: MapObject[]
  grid: MapGrid
  onPick?: (id: string) => void
}

// Walls are the most numerous prefab, so they share one instanced mesh.
export function WallInstances({ walls, grid, onPick }: Props) {
  const ref = useRef<InstancedMesh>(null)
  const ids = useMemo(() => walls.map((w) => w.id), [walls])

  useLayoutEffect(() => {
    const mesh = ref.current
    if (!mesh) return
    const dummy = new Object3D()
    const color = new Color()
    walls.forEach((w, i) => {
      const p = placement(grid, w)
      dummy.position.set(p.cx, p.y + 1.5, p.cz)
      dummy.updateMatrix()
      mesh.setMatrixAt(i, dummy.matrix)
      mesh.setColorAt(i, color.set(C.brick).offsetHSL(0, 0, ((w.x * 7 + w.z * 13) % 5) * 0.012))
    })
    mesh.instanceMatrix.needsUpdate = true
    if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true
    mesh.computeBoundingSphere()
  }, [walls, grid])

  const down = (e: ThreeEvent<PointerEvent>) => {
    if (!onPick || e.button !== 0 || e.instanceId === undefined) return
    e.stopPropagation()
    onPick(ids[e.instanceId])
  }

  if (!walls.length) return null
  return <instancedMesh key={walls.length} ref={ref} args={[geometry, material, walls.length]} onPointerDown={down} />
}
