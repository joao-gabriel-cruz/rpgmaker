import type { Coord } from '../grid'
import { addObject, applyObjectPatch, EMPTY_PATCH, objectsAt, type Bounds, type MapObject, type ObjectOpResult } from './model'

const range = (a: number, b: number) =>
  Array.from({ length: Math.abs(b - a) + 1 }, (_, i) => a + i * Math.sign(b - a))

// Straight line along x or z, or an L (x first, then z) when both axes differ.
export function wallPath(from: Coord, to: Coord): Coord[] {
  const horizontal = range(from.x, to.x).map((x) => ({ x, z: from.z }))
  const vertical = range(from.z, to.z).slice(1).map((z) => ({ x: to.x, z }))
  return [...horizontal, ...vertical]
}

// Adds wall segments along the path, skipping cells that already hold a wall.
export function drawWall(objects: readonly MapObject[], b: Bounds, from: Coord, to: Coord): ObjectOpResult {
  const patch = { ...EMPTY_PATCH, add: [] as MapObject[] }
  let current = objects
  for (const { x, z } of wallPath(from, to)) {
    if (objectsAt(current, x, z).some((o) => o.prefab === 'wall')) continue
    const r = addObject(current, b, { prefab: 'wall', x, z })
    patch.add.push(...r.patch.add)
    current = r.objects
  }
  return { objects: applyObjectPatch(objects, patch), patch }
}
