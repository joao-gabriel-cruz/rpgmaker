import { api } from '../../lib/api'
import type { Ambience } from '../ambience/ambience'
import { deserializeGrid, type MapGrid, type SerializedGrid } from '../map3d/grid'
import type { ThumbObject } from './thumbnail'

export interface TemplateInfo {
  id: string
  name: string
  description: string
  tags: string[]
  width: number
  depth: number
  ambience: Pick<Ambience, 'timeOfDay' | 'weather'>
}

interface PreviewDto {
  grid: SerializedGrid
  objects: ThumbObject[] | null
}

export interface TemplatePreview {
  grid: MapGrid
  objects: ThumbObject[]
}

const base = (campaignId: string) => `/campaigns/${campaignId}/map-templates`

export const templatesKey = (campaignId: string) => ['campaigns', campaignId, 'map-templates']

export const listTemplates = (campaignId: string) => api.get<TemplateInfo[]>(base(campaignId))

export async function fetchPreview(campaignId: string, id: string, seed: number): Promise<TemplatePreview> {
  const dto = await api.get<PreviewDto>(`${base(campaignId)}/${id}/preview?seed=${seed}`)
  return { grid: deserializeGrid(dto.grid), objects: dto.objects ?? [] }
}

export const randomSeed = () => 1 + Math.floor(Math.random() * 2_000_000_000)

/** Every tag the gallery can filter by (the order of the chips). */
export const ALL_TAGS = ['urbano', 'natureza', 'refúgio', 'combate', 'social', 'perigo']
