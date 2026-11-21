import { applyPatch, type MapPatch } from '../map3d/grid'
import { applyObjectPatch } from '../map3d/objects/model'
import type { MapState, ObjectsEvent, PaintEvent } from './scenes'

// Remote patches are applied only when they target the scene on screen.
export function applyRemotePaint(state: MapState, sceneId: string, ev: PaintEvent): MapState {
  if (ev?.sceneId !== sceneId || !ev.patch?.cells) return state
  return { ...state, grid: applyPatch(state.grid, ev.patch) }
}

export function applyRemoteObjects(state: MapState, sceneId: string, ev: ObjectsEvent): MapState {
  if (ev?.sceneId !== sceneId || !ev.patch) return state
  const { add = [], remove = [], update = [] } = ev.patch
  const fix = (o: (typeof add)[number]) => ({ ...o, props: o.props ?? {} })
  const patch = { add: add.map(fix), remove, update: update.map(fix) }
  return { ...state, objects: applyObjectPatch(state.objects, patch) }
}

/** Max cells per map.paint message: the socket read limit is 64 KB, so big strokes are split. */
export const PAINT_CHUNK = 1000

export function chunkPatch(patch: MapPatch): MapPatch[] {
  const out: MapPatch[] = []
  for (let i = 0; i < patch.cells.length; i += PAINT_CHUNK) out.push({ cells: patch.cells.slice(i, i + PAINT_CHUNK) })
  return out.length ? out : [patch]
}
