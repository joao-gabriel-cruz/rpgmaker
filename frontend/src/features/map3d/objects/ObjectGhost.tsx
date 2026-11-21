import { useLayoutEffect, useRef } from 'react'
import type { Group, Material, Mesh } from 'three'
import { PREFAB_META_BY_ID } from './catalogData'
import { footprint, fitsInGrid, type MapObject } from './model'
import type { MapGrid } from '../grid'
import { ObjectInstance } from './ObjectInstance'
import { placement } from './placement'

function makeTranslucent(root: Group) {
  root.traverse((n) => {
    const mesh = n as Mesh
    if (!mesh.isMesh) return
    const mat = (mesh.material as Material).clone()
    mat.transparent = true
    mat.opacity = 0.55
    mat.depthWrite = false
    mesh.material = mat
    mesh.raycast = () => null
  })
}

// Translucent preview of an object that is not (yet) part of the map.
export function ObjectGhost({ obj, grid }: { obj: MapObject; grid: MapGrid }) {
  const ref = useRef<Group>(null)
  useLayoutEffect(() => {
    if (ref.current) makeTranslucent(ref.current)
  }, [obj.prefab])
  const p = placement(grid, obj)
  const ok = fitsInGrid(grid, obj)
  return (
    <>
      <group ref={ref} key={obj.prefab}>
        <ObjectInstance obj={obj} grid={grid} />
      </group>
      <mesh position={[p.cx, p.y + 0.04, p.cz]} raycast={() => null}>
        <boxGeometry args={[p.w, 0.08, p.d]} />
        <meshBasicMaterial color={ok ? '#34d399' : '#ef4444'} transparent opacity={0.4} />
      </mesh>
    </>
  )
}

// Wireframe box around the selected object.
export function SelectionBox({ obj, grid }: { obj: MapObject; grid: MapGrid }) {
  const p = placement(grid, obj)
  const { w, d } = footprint(obj.prefab, obj.rotation)
  const h = PREFAB_META_BY_ID[obj.prefab]?.height ?? 1
  return (
    <mesh position={[p.cx, p.y + h / 2, p.cz]} raycast={() => null}>
      <boxGeometry args={[w + 0.05, h, d + 0.05]} />
      <meshBasicMaterial color="#fbbf24" wireframe />
    </mesh>
  )
}
