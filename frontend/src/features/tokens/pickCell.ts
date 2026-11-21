import { Plane, Raycaster, Vector2, Vector3, type Camera } from 'three'
import { getCell, inBounds, type Coord, type MapGrid } from '../map3d/grid'

const ray = new Raycaster()
const hit = new Vector3()

function intersect(ndc: Vector2, camera: Camera, y: number): Vector3 | null {
  ray.setFromCamera(ndc, camera)
  return ray.ray.intersectPlane(new Plane(new Vector3(0, 1, 0), -y), hit) ? hit.clone() : null
}

/** World point under the pointer on the ground, refined once by the height of the cell it lands on. */
export function pickPoint(grid: MapGrid, ndc: Vector2, camera: Camera): { x: number; z: number } | null {
  let p = intersect(ndc, camera, 0)
  if (!p) return null
  const c = getCell(grid, Math.floor(p.x), Math.floor(p.z))
  if (c) p = intersect(ndc, camera, c.height) ?? p
  return { x: p.x, z: p.z }
}

export function pickCell(grid: MapGrid, ndc: Vector2, camera: Camera): Coord | null {
  const p = pickPoint(grid, ndc, camera)
  if (!p) return null
  const c = { x: Math.floor(p.x), z: Math.floor(p.z) }
  return inBounds(grid, c.x, c.z) ? c : null
}
