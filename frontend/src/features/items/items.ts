import { useCallback, useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'

export type ItemKind = 'melee' | 'ranged' | 'tool' | 'consumable' | 'clothing' | 'resource' | 'key' | 'other'
export type ActionCategory = 'combat' | 'exploration' | 'social' | 'survival'
export type AttackAnimation = 'slash' | 'blunt' | 'shot' | 'arrow'
export type RefugeResourceKey = 'water' | 'food' | 'wood'

export interface ItemEffects {
  keptBonus?: { actionCategory: ActionCategory; amount: number }
  heal?: number
  refugeResource?: { resource: RefugeResourceKey; amount: number }
}

export interface CompendiumItem {
  id: string
  name: string
  description: string
  icon: string
  color: string
  kind: ItemKind
  quality: number
  scarcity: number
  pressure: number
  traits: string[]
  effects: ItemEffects
  attackAnimation?: AttackAnimation
}

export type ItemDraft = Omit<CompendiumItem, 'id'>

export const KIND_LABEL: Record<ItemKind, string> = {
  melee: 'Arma corpo a corpo', ranged: 'Arma à distância', tool: 'Ferramenta', consumable: 'Consumível',
  clothing: 'Vestuário', resource: 'Recurso', key: 'Chave / missão', other: 'Outro',
}
export const CATEGORY_LABEL: Record<ActionCategory, string> = {
  combat: 'Combate', exploration: 'Exploração', social: 'Social', survival: 'Sobrevivência',
}
export const ANIMATION_LABEL: Record<AttackAnimation, string> = { slash: 'Corte', blunt: 'Impacto', shot: 'Tiro', arrow: 'Flecha' }
export const RESOURCE_LABEL: Record<RefugeResourceKey, string> = { water: 'Água', food: 'Alimento', wood: 'Madeira' }

// Must match the server's curated list (backend/internal/items/validate.go).
export const ICON_NAMES = [
  'sword', 'swords', 'slice', 'crosshair', 'target', 'hammer', 'axe', 'wrench', 'key', 'backpack',
  'wine', 'milk', 'pill', 'cross', 'bandage', 'flashlight', 'map', 'radio', 'battery', 'cable',
  'drumstick', 'apple', 'droplet', 'trees', 'gem', 'book', 'skull', 'shield', 'shirt', 'flame',
  'package', 'cog', 'scissors', 'tent', 'lock', 'bomb', 'syringe', 'briefcase', 'bone', 'leaf',
] as const

export const DEFAULT_COLOR = '#34d399'

export const blankItem = (): ItemDraft => ({
  name: '', description: '', icon: 'package', color: DEFAULT_COLOR, kind: 'other',
  quality: 0, scarcity: 0, pressure: 0, traits: [], effects: {},
})

export const isWeapon = (k: ItemKind) => k === 'melee' || k === 'ranged'

export function filterItems(list: CompendiumItem[], query: string, kind: ItemKind | ''): CompendiumItem[] {
  const q = query.trim().toLowerCase()
  return list.filter((i) => (!kind || i.kind === kind) && (!q || i.name.toLowerCase().includes(q) || i.traits.some((t) => t.toLowerCase().includes(q))))
}

/** Short PT-BR summary of an item's effects, for cards. */
export function effectLines(e: ItemEffects | undefined): string[] {
  const out: string[] = []
  if (e?.keptBonus) out.push(`+${e.keptBonus.amount} mantido em ${CATEGORY_LABEL[e.keptBonus.actionCategory].toLowerCase()}`)
  if (e?.heal) out.push(`Cura ${e.heal}`)
  if (e?.refugeResource) out.push(`+${e.refugeResource.amount} ${RESOURCE_LABEL[e.refugeResource.resource].toLowerCase()} no Refúgio`)
  return out
}

const base = (campaignId: string) => `/campaigns/${campaignId}/items`

export const saveItem = (campaignId: string, draft: ItemDraft, id?: string) =>
  id ? api.put<CompendiumItem>(`${base(campaignId)}/${id}`, draft) : api.post<CompendiumItem>(base(campaignId), draft)
export const deleteItem = (campaignId: string, id: string) => api.delete<void>(`${base(campaignId)}/${id}`)

/** The Mestre's compendium, loaded once and kept live by items.changed. */
export function useItems(campaignId: string) {
  const { subscribe } = useRealtime()
  const [items, setItems] = useState<CompendiumItem[]>([])
  const [error, setError] = useState('')
  const reload = useCallback(() => {
    api.get<{ items: CompendiumItem[] }>(base(campaignId)).then((r) => setItems(Array.isArray(r?.items) ? r.items : []), (e: Error) => setError(e.message))
  }, [campaignId])
  useEffect(reload, [reload])
  useEffect(() => subscribe('items.changed', (p) => setItems((p as { items: CompendiumItem[] }).items ?? [])), [subscribe])
  return { items, error, reload }
}
