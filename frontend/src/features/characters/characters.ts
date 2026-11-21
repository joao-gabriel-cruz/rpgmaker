import { useQuery } from '@tanstack/react-query'
import { api } from '../../lib/api'
import type { Sheet, SheetIssue } from '../sheet-assimilacao/types'

export interface ServerIssue {
  code?: string
  severity: 'blocking' | 'limit'
  field: string
  message: string
}

export interface Derived {
  boxesPerLevel: number
  currentLevel: number
  levelName: string
}

export interface CharacterView {
  id: string
  campaignId: string
  ownerId: string | null
  ownerName: string
  kind: 'pc' | 'npc'
  name: string
  sheet: Sheet
  updatedAt: string
  derived?: Derived
  issues: ServerIssue[]
}

export const charactersKey = (campaignId: string) => ['campaigns', campaignId, 'characters'] as const
export const characterKey = (campaignId: string, id: string) => [...charactersKey(campaignId), id] as const

export const toSheetIssues = (issues: ServerIssue[]): SheetIssue[] =>
  issues.map((i) => ({ code: i.code ?? '', field: i.field, message: i.message, blocking: i.severity === 'blocking' }))

export const useCharacters = (campaignId: string) =>
  useQuery({ queryKey: charactersKey(campaignId), queryFn: () => api.get<CharacterView[]>(`/campaigns/${campaignId}/characters`) })

export const useCharacter = (campaignId: string, id: string) =>
  useQuery({
    queryKey: characterKey(campaignId, id),
    queryFn: () => api.get<CharacterView>(`/campaigns/${campaignId}/characters/${id}`),
    retry: false,
    staleTime: Infinity,
    refetchOnWindowFocus: false,
  })

export const createCharacter = (campaignId: string) =>
  api.post<CharacterView>(`/campaigns/${campaignId}/characters`, {})

export const saveSheet = (campaignId: string, id: string, sheet: Sheet) =>
  api.put<CharacterView>(`/campaigns/${campaignId}/characters/${id}`, { sheet })

export const deleteCharacter = (campaignId: string, id: string) =>
  api.delete<void>(`/campaigns/${campaignId}/characters/${id}`)
