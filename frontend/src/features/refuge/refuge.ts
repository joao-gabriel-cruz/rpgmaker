import { useQuery } from '@tanstack/react-query'
import { api } from '../../lib/api'
import type { Item } from '../sheet-assimilacao/types'
import type { Building, Proposal, WeekReport } from './buildings'

export type ResourceKey = 'water' | 'food' | 'medicine' | 'clothing' | 'plants' | 'animals' | 'wood' | 'biomass' | 'minerals' | 'fuel' | 'ammo' | 'buildingMaterial'
export type NaturalKey = 'water' | 'plants' | 'animals' | 'wood' | 'minerals' | 'biomass'
export type Abundance = 'none' | 'scarce' | 'medium' | 'abundant'
export type LevelKey = 'defense' | 'morale' | 'belligerence'
export type RegionKey = 'size' | 'danger' | 'habitation' | 'naturalResources' | 'contamination' | 'travel'

export interface Region extends Record<RegionKey, number> {
  abundance: Record<NaturalKey, Abundance>
}

/** Editable by the Mestre (PUT body). */
export interface RefugeEditable {
  name: string
  description: string
  notes: string
  defense: number
  morale: number
  belligerence: number
  mobility: number
  population: { level: number; max: number }
  reservesMax: number
  hasWaterSource: boolean
  resources: Record<ResourceKey, number>
  region: Region
  week: number
}

export interface Derived {
  defenseLabel: string
  moraleLabel: string
  belligerenceLabel: string
  inhabitants: number
  workers: number
  reservesTotal: number
  reservesFree: number
  reservesOverflow: number
}

export interface Refuge extends RefugeEditable {
  campaignId: string
  items: Item[]
  buildings: Building[]
  proposals: Proposal[]
  history: WeekReport[]
  crises: unknown[]
  residents: unknown[]
  updatedAt: string
  derived: Derived
}

export const MAX_LEVEL = 6

export const LEVEL_LABELS: Record<LevelKey, string[]> = {
  defense: ['Nenhuma', 'Linhas e sinos', 'Cerca de arame farpado', 'Muro de madeira', 'Muro de pedra/tijolo', 'Complexo prisional', 'Base militar'],
  morale: ['Amotinada', 'Desiludida', 'Hesitante', 'Resoluta', 'Animada', 'Empenhada', 'Exaltada'],
  belligerence: ['Pacifista', 'Mínima', 'Razoável', 'Eficiente', 'Ameaçadora', 'Terrível', 'Arrasadora'],
}
export const LEVEL_TITLES: Record<LevelKey, string> = { defense: 'Defesa', morale: 'Moral', belligerence: 'Beligerância' }

export const RESOURCE_LABEL: Record<ResourceKey, string> = {
  water: 'Água', food: 'Alimento', medicine: 'Remédios', clothing: 'Vestuário', plants: 'Plantas', animals: 'Animais',
  wood: 'Madeira', biomass: 'Biomassa', minerals: 'Minerais', fuel: 'Combustível', ammo: 'Munição', buildingMaterial: 'Material de Construção',
}
export const RESOURCE_GROUPS: { title: string; keys: ResourceKey[] }[] = [
  { title: 'Sobrevivência', keys: ['water', 'food', 'medicine', 'clothing'] },
  { title: 'Orgânicos', keys: ['plants', 'animals', 'wood', 'biomass'] },
  { title: 'Indústria', keys: ['minerals', 'fuel', 'ammo', 'buildingMaterial'] },
]
export const REGION_LABEL: Record<RegionKey, string> = {
  size: 'Tamanho', danger: 'Perigo', habitation: 'Habitação', naturalResources: 'Recursos naturais', contamination: 'Contaminação', travel: 'Deslocamento',
}
export const REGION_SCALE = ['Nenhum', 'Mínimo', 'Baixo', 'Médio', 'Alto', 'Muito alto', 'Extremo']
export const NATURAL_KEYS: NaturalKey[] = ['water', 'plants', 'animals', 'wood', 'minerals', 'biomass']
export const ABUNDANCE_LABEL: Record<Abundance, string> = { none: 'Inexistente', scarce: 'Escasso', medium: 'Médio', abundant: 'Abundante' }

export const refugeKey = (campaignId: string) => ['campaigns', campaignId, 'refuge'] as const

export const useRefuge = (campaignId: string) =>
  useQuery({ queryKey: refugeKey(campaignId), queryFn: () => api.get<Refuge>(`/campaigns/${campaignId}/refuge`) })

export const saveRefuge = (campaignId: string, d: RefugeEditable) =>
  api.put<Refuge>(`/campaigns/${campaignId}/refuge`, d)

export const editableOf = (r: Refuge): RefugeEditable => ({
  name: r.name, description: r.description, notes: r.notes, defense: r.defense, morale: r.morale, belligerence: r.belligerence,
  mobility: r.mobility, population: r.population, reservesMax: r.reservesMax, hasWaterSource: r.hasWaterSource,
  resources: r.resources, region: r.region, week: r.week,
})

export const clamp = (v: number, lo: number, hi: number) => Math.min(hi, Math.max(lo, v))
