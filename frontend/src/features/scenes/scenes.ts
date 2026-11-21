import { api } from '../../lib/api'
import { deserializeGrid, type MapGrid, type MapPatch, type SerializedGrid } from '../map3d/grid'
import type { MapObject, ObjectPatch } from '../map3d/objects/model'

export const SCENE_SIZES = [16, 24, 32, 48, 64, 96, 128] as const

export interface SceneSummary {
  id: string
  name: string
  width: number
  depth: number
  isActive: boolean
}

export interface SceneDto {
  id: string
  name: string
  grid: SerializedGrid
  objects: ObjectDto[] | null
  draft?: { grid: SerializedGrid; objects: ObjectDto[] | null }
}

type ObjectDto = Omit<MapObject, 'props'> & { props?: MapObject['props'] }

export interface MapState {
  grid: MapGrid
  objects: readonly MapObject[]
  /** True while the Mestre edits a hidden draft (grid/objects are then the draft's). */
  draft?: boolean
}

export interface SnapshotDto {
  id: string
  name: string
  createdAt: string
}

export interface ReplacedEvent {
  sceneId: string
  grid: SerializedGrid
  objects: ObjectDto[] | null
}

const withProps = (list: ObjectDto[] | null) => (list ?? []).map((o) => ({ ...o, props: o.props ?? {} }))

export const replacedState = (ev: ReplacedEvent): MapState =>
  ({ grid: deserializeGrid(ev.grid), objects: withProps(ev.objects), draft: false })

export const listSnapshots = (campaignId: string, sceneId: string) =>
  api.get<SnapshotDto[]>(`${base(campaignId)}/${sceneId}/snapshots`)
export const saveSnapshot = (campaignId: string, sceneId: string, name: string) =>
  api.post<SnapshotDto>(`${base(campaignId)}/${sceneId}/snapshots`, { name })
export const deleteSnapshot = (campaignId: string, sceneId: string, id: string) =>
  api.delete<void>(`${base(campaignId)}/${sceneId}/snapshots/${id}`)

export const base = (campaignId: string) => `/campaigns/${campaignId}/scenes`

export const listScenes = (campaignId: string) => api.get<SceneSummary[]>(base(campaignId))
export const createScene = (campaignId: string, name: string, size: number) =>
  api.post<SceneDto>(base(campaignId), { name, width: size, depth: size })
export const createSceneFromTemplate = (campaignId: string, name: string, templateId: string, seed: number) =>
  api.post<SceneDto>(base(campaignId), { name, templateId, seed })
export const renameScene = (campaignId: string, id: string, name: string) =>
  api.patch<void>(`${base(campaignId)}/${id}`, { name })
export const deleteScene = (campaignId: string, id: string) => api.delete<void>(`${base(campaignId)}/${id}`)
export const activateScene = (campaignId: string, id: string) =>
  api.post<void>(`${base(campaignId)}/${id}/activate`)

// One GET carries grid + objects; props default to {} because the server omits empty maps.
export async function loadScene(campaignId: string, id: string): Promise<MapState> {
  const dto = await api.get<SceneDto>(`${base(campaignId)}/${id}`)
  const src = dto.draft ?? dto
  return { grid: deserializeGrid(src.grid), objects: withProps(src.objects), draft: !!dto.draft }
}

export interface PatchedEvent<P> {
  sceneId: string
  patch: P
}

export type PaintEvent = PatchedEvent<MapPatch>
export type ObjectsEvent = PatchedEvent<ObjectPatch>
