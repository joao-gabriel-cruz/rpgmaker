import { api } from '../../lib/api'
import { deserializeGrid, type MapGrid, type MapPatch, type SerializedGrid } from '../map3d/grid'
import type { MapObject, ObjectPatch } from '../map3d/objects/model'

export const SCENE_SIZES = [16, 24, 32, 48, 64] as const

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
  objects: (Omit<MapObject, 'props'> & { props?: MapObject['props'] })[] | null
}

export interface MapState {
  grid: MapGrid
  objects: readonly MapObject[]
}

export const base = (campaignId: string) => `/campaigns/${campaignId}/scenes`

export const listScenes = (campaignId: string) => api.get<SceneSummary[]>(base(campaignId))
export const createScene = (campaignId: string, name: string, size: number) =>
  api.post<SceneDto>(base(campaignId), { name, width: size, depth: size })
export const renameScene = (campaignId: string, id: string, name: string) =>
  api.patch<void>(`${base(campaignId)}/${id}`, { name })
export const deleteScene = (campaignId: string, id: string) => api.delete<void>(`${base(campaignId)}/${id}`)
export const activateScene = (campaignId: string, id: string) =>
  api.post<void>(`${base(campaignId)}/${id}/activate`)

// One GET carries grid + objects; props default to {} because the server omits empty maps.
export async function loadScene(campaignId: string, id: string): Promise<MapState> {
  const dto = await api.get<SceneDto>(`${base(campaignId)}/${id}`)
  const objects = (dto.objects ?? []).map((o) => ({ ...o, props: o.props ?? {} }))
  return { grid: deserializeGrid(dto.grid), objects }
}

export interface PatchedEvent<P> {
  sceneId: string
  patch: P
}

export type PaintEvent = PatchedEvent<MapPatch>
export type ObjectsEvent = PatchedEvent<ObjectPatch>
