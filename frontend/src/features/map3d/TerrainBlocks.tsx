import { useFrame } from '@react-three/fiber'
import { useLayoutEffect, useMemo, useRef } from 'react'
import { InstancedMesh, MeshStandardMaterial, Object3D, BoxGeometry } from 'three'
import { layoutBlocks } from './blockLayout'
import type { MapGrid } from './grid'
import { TERRAIN_LIST, type TerrainType } from './terrainTypes'

const geometry = new BoxGeometry(1, 1, 1)

function makeMaterial(t: TerrainType) {
  return new MeshStandardMaterial({
    color: t.color,
    emissive: t.emissive ?? '#000000',
    emissiveIntensity: t.animated === 'lava' ? 0.6 : t.animated === 'water' ? 0.15 : 0,
    flatShading: true,
    roughness: 0.9,
  })
}

function BlockLayer({ type, positions }: { type: TerrainType; positions: number[] }) {
  const ref = useRef<InstancedMesh>(null)
  const material = useMemo(() => makeMaterial(type), [type])
  const count = positions.length / 3

  useLayoutEffect(() => {
    const mesh = ref.current
    if (!mesh) return
    const dummy = new Object3D()
    for (let i = 0; i < count; i++) {
      dummy.position.set(positions[i * 3], positions[i * 3 + 1], positions[i * 3 + 2])
      dummy.updateMatrix()
      mesh.setMatrixAt(i, dummy.matrix)
    }
    mesh.instanceMatrix.needsUpdate = true
    mesh.computeBoundingSphere()
  }, [positions, count])

  useFrame(({ clock }) => {
    const t = clock.elapsedTime
    if (type.animated === 'lava') material.emissiveIntensity = 0.55 + Math.sin(t * 2) * 0.25
    if (type.animated === 'water') {
      material.color.set(type.color).offsetHSL(0, 0, Math.sin(t * 1.5) * 0.03)
    }
  })

  if (count === 0) return null
  return <instancedMesh key={count} ref={ref} args={[geometry, material, count]} userData={{ terrain: type.id }} />
}

export function TerrainBlocks({ grid }: { grid: MapGrid }) {
  const layout = useMemo(() => layoutBlocks(grid), [grid])
  return (
    <>
      {TERRAIN_LIST.map((t) => (
        <BlockLayer key={t.id} type={t} positions={layout[t.id]} />
      ))}
    </>
  )
}

