import { useQuery } from '@tanstack/react-query'
import { api } from '../../lib/api'
import type { Item } from '../sheet-assimilacao/types'

export interface RefugeDetails {
  name: string
  description: string
  defense: number
  morale: number
  mobility: number
  belligerence: number
  water: number
  food: number
  wood: number
  notes: string
}

export interface Refuge extends RefugeDetails {
  campaignId: string
  items: Item[]
  updatedAt: string
}

export type NumericField = 'defense' | 'morale' | 'mobility' | 'belligerence' | 'water' | 'food' | 'wood'

export const ATTRIBUTES: { key: NumericField; label: string }[] = [
  { key: 'defense', label: 'Defesa' },
  { key: 'morale', label: 'Moral' },
  { key: 'mobility', label: 'Mobilidade' },
  { key: 'belligerence', label: 'Beligerância' },
]
export const RESOURCES: { key: NumericField; label: string }[] = [
  { key: 'water', label: 'Água' },
  { key: 'food', label: 'Alimento' },
  { key: 'wood', label: 'Madeira' },
]
export const ATTRIBUTE_MAX = 10

export const refugeKey = (campaignId: string) => ['campaigns', campaignId, 'refuge'] as const

export const useRefuge = (campaignId: string) =>
  useQuery({ queryKey: refugeKey(campaignId), queryFn: () => api.get<Refuge>(`/campaigns/${campaignId}/refuge`) })

export const saveRefuge = (campaignId: string, d: RefugeDetails) =>
  api.put<Refuge>(`/campaigns/${campaignId}/refuge`, d)

export const detailsOf = ({ name, description, defense, morale, mobility, belligerence, water, food, wood, notes }: Refuge): RefugeDetails =>
  ({ name, description, defense, morale, mobility, belligerence, water, food, wood, notes })

export function stepValue(field: NumericField, value: number, delta: number): number {
  const max = ATTRIBUTES.some((a) => a.key === field) ? ATTRIBUTE_MAX : Infinity
  return Math.min(max, Math.max(0, value + delta))
}
