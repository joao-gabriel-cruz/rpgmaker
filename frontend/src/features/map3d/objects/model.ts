import type { Coord, MapGrid } from '../grid'
import { PREFAB_META_BY_ID, type ObjectProps } from './catalogData'

export type Rotation = 0 | 90 | 180 | 270
export type Bounds = Pick<MapGrid, 'width' | 'depth'>

export interface MapObject {
  id: string
  prefab: string
  x: number
  z: number
  rotation: Rotation
  props: ObjectProps
}

export interface ObjectPatch {
  add: MapObject[]
  remove: string[]
  update: MapObject[]
}

export interface ObjectOpResult {
  objects: readonly MapObject[]
  patch: ObjectPatch
}

export const ROTATIONS: Rotation[] = [0, 90, 180, 270]
export const EMPTY_PATCH: ObjectPatch = { add: [], remove: [], update: [] }

export const isEmptyPatch = (p: ObjectPatch) => !p.add.length && !p.remove.length && !p.update.length

export const nextRotation = (r: Rotation, step: 90 | -90 = 90): Rotation =>
  ROTATIONS[(ROTATIONS.indexOf(r) + (step === 90 ? 1 : 3)) % 4]

export function footprint(prefab: string, rotation: Rotation): { w: number; d: number } {
  const meta = PREFAB_META_BY_ID[prefab]
  if (!meta) return { w: 0, d: 0 }
  return rotation === 90 || rotation === 270
    ? { w: meta.depth, d: meta.width }
    : { w: meta.width, d: meta.depth }
}

export function occupiedCells(o: Pick<MapObject, 'prefab' | 'x' | 'z' | 'rotation'>): Coord[] {
  const { w, d } = footprint(o.prefab, o.rotation)
  const out: Coord[] = []
  for (let dz = 0; dz < d; dz++) for (let dx = 0; dx < w; dx++) out.push({ x: o.x + dx, z: o.z + dz })
  return out
}

export function fitsInGrid(b: Bounds, o: Pick<MapObject, 'prefab' | 'x' | 'z' | 'rotation'>): boolean {
  const { w, d } = footprint(o.prefab, o.rotation)
  if (!w || !Number.isInteger(o.x) || !Number.isInteger(o.z)) return false
  return o.x >= 0 && o.z >= 0 && o.x + w <= b.width && o.z + d <= b.depth
}

export function objectsAt(objects: readonly MapObject[], x: number, z: number): MapObject[] {
  return objects.filter((o) => occupiedCells(o).some((c) => c.x === x && c.z === z))
}

export function applyObjectPatch(objects: readonly MapObject[], patch: ObjectPatch): MapObject[] {
  const removed = new Set(patch.remove)
  const updates = new Map(patch.update.map((o) => [o.id, o]))
  const kept = objects.filter((o) => !removed.has(o.id)).map((o) => updates.get(o.id) ?? o)
  const known = new Set(kept.map((o) => o.id))
  return [...kept, ...patch.add.filter((o) => !known.has(o.id))]
}

function result(objects: readonly MapObject[], patch: Partial<ObjectPatch>): ObjectOpResult {
  const full = { ...EMPTY_PATCH, ...patch }
  return { objects: isEmptyPatch(full) ? objects : applyObjectPatch(objects, full), patch: full }
}

export interface NewObject {
  prefab: string
  x: number
  z: number
  rotation?: Rotation
  props?: ObjectProps
  id?: string
}

export function addObject(objects: readonly MapObject[], b: Bounds, n: NewObject): ObjectOpResult {
  const meta = PREFAB_META_BY_ID[n.prefab]
  const obj: MapObject = {
    id: n.id ?? crypto.randomUUID(),
    prefab: n.prefab,
    x: n.x,
    z: n.z,
    rotation: n.rotation ?? 0,
    props: { ...meta?.defaultProps, ...n.props },
  }
  if (!meta || !fitsInGrid(b, obj) || objects.some((o) => o.id === obj.id)) return result(objects, {})
  return result(objects, { add: [obj] })
}

export function removeObject(objects: readonly MapObject[], id: string): ObjectOpResult {
  return result(objects, objects.some((o) => o.id === id) ? { remove: [id] } : {})
}

function replace(objects: readonly MapObject[], b: Bounds, id: string, change: Partial<MapObject>) {
  const cur = objects.find((o) => o.id === id)
  const next = cur && { ...cur, ...change }
  return next && fitsInGrid(b, next) ? result(objects, { update: [next] }) : result(objects, {})
}

export function rotateObject(objects: readonly MapObject[], b: Bounds, id: string, step: 90 | -90 = 90) {
  const cur = objects.find((o) => o.id === id)
  return cur ? replace(objects, b, id, { rotation: nextRotation(cur.rotation, step) }) : result(objects, {})
}

export const moveObject = (objects: readonly MapObject[], b: Bounds, id: string, x: number, z: number) =>
  replace(objects, b, id, { x, z })

export function setObjectProps(objects: readonly MapObject[], id: string, props: ObjectProps) {
  const cur = objects.find((o) => o.id === id)
  return cur ? result(objects, { update: [{ ...cur, props: { ...cur.props, ...props } }] }) : result(objects, {})
}

// Drops malformed entries from untrusted data (e.g. localStorage).
export function sanitizeObjects(raw: unknown, b: Bounds): MapObject[] {
  if (!Array.isArray(raw)) return []
  return raw.filter((o): o is MapObject => {
    const ok = o && typeof o.id === 'string' && ROTATIONS.includes(o.rotation) && o.props
    return Boolean(ok && PREFAB_META_BY_ID[o.prefab] && fitsInGrid(b, o))
  })
}
