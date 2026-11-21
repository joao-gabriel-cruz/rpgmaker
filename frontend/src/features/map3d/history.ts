import { getCell, type CellPatch, type MapGrid, type MapPatch } from './grid'
import type { MapObject, ObjectPatch } from './objects/model'

export type Edit =
  | { kind: 'paint'; undo: MapPatch; redo: MapPatch }
  | { kind: 'objects'; undo: ObjectPatch; redo: ObjectPatch }

export const MAX_HISTORY = 100

/** Patch that restores the cells a paint patch is about to overwrite. */
export function inversePaint(grid: MapGrid, patch: MapPatch): MapPatch {
  const cells: CellPatch[] = []
  for (const p of patch.cells) {
    const prev = getCell(grid, p.x, p.z)
    if (prev) cells.push({ x: p.x, z: p.z, terrain: prev.terrain, height: prev.height })
  }
  return { cells }
}

/** Patch that reverts an object patch applied on top of `objects`. */
export function inverseObjects(objects: readonly MapObject[], patch: ObjectPatch): ObjectPatch {
  const byId = new Map(objects.map((o) => [o.id, o]))
  return {
    add: patch.remove.flatMap((id) => byId.get(id) ?? []),
    remove: patch.add.filter((o) => !byId.has(o.id)).map((o) => o.id),
    update: patch.update.flatMap((o) => byId.get(o.id) ?? []),
  }
}

const key = (c: { x: number; z: number }) => `${c.x},${c.z}`

/** Folds a later brush patch into an open stroke: the oldest "before" and the newest "after" of each cell win. */
export function mergePaint(prev: Edit & { kind: 'paint' }, undo: MapPatch, redo: MapPatch): Edit {
  const seen = new Set(prev.undo.cells.map(key))
  const later = new Set(redo.cells.map(key))
  return {
    kind: 'paint',
    undo: { cells: [...prev.undo.cells, ...undo.cells.filter((c) => !seen.has(key(c)))] },
    redo: { cells: [...prev.redo.cells.filter((c) => !later.has(key(c))), ...redo.cells] },
  }
}
