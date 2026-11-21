import type { Appearance } from '../miniature/appearance'
export type Generation = 'pre_collapse' | 'during_collapse' | 'post_collapse'
export type Suit = 'hearts' | 'diamonds' | 'spades' | 'clubs'

export const INSTINCT_KEYS = ['influence', 'perception', 'potency', 'reaction', 'resolution', 'sagacity'] as const
export const KNOWLEDGE_KEYS = ['biology', 'erudition', 'engineering', 'geography', 'medicine', 'security'] as const
export const PRACTICE_KEYS = ['weapons', 'athletics', 'expression', 'stealth', 'crafting', 'survival'] as const

export type InstinctKey = (typeof INSTINCT_KEYS)[number]
export type KnowledgeKey = (typeof KNOWLEDGE_KEYS)[number]
export type PracticeKey = (typeof PRACTICE_KEYS)[number]
export type AptitudeKey = KnowledgeKey | PracticeKey

export interface Item {
  name: string
  quality: number
  scarcity: number
  pressure: number
  traits: string[]
}

export interface Characteristic {
  name: string
  level: number
  description: string
}

export interface SheetAssimilation {
  name: string
  suit: Suit
  description: string
}

export interface TugOfWar {
  determinationLevel: number
  determinationPoints: number
  assimilationLevel: number
  assimilationPoints: number
}

export interface Sheet {
  identity: {
    name: string
    portraitUrl: string | null
    generation: Generation
    occupation: string
    markingEvent: string
    personalPurposes: string[]
    collectivePurposes: string[]
  }
  instincts: Record<InstinctKey, number>
  knowledge: Record<KnowledgeKey, number>
  practices: Record<PracticeKey, number>
  tugOfWar: TugOfWar
  health: { damage: number }
  inventory: { body: Item[]; backpack: Item[] }
  characteristics: Characteristic[]
  xp: { total: number }
  assimilations: SheetAssimilation[]
  notes: string
  /** Optional 3D miniature description (validated by the server). */
  appearance?: Appearance
}

export interface SheetIssue {
  code: string
  field: string
  message: string
  blocking: boolean
}

export interface RollRequest {
  instinct: InstinctKey
  aptitude: AptitudeKey
}

export const BODY_LIMIT = 3
export const BACKPACK_LIMIT = 6

export function newSheet(): Sheet {
  return {
    identity: {
      name: '',
      portraitUrl: null,
      generation: 'post_collapse',
      occupation: '',
      markingEvent: '',
      personalPurposes: ['', ''],
      collectivePurposes: ['', ''],
    },
    instincts: { influence: 1, perception: 1, potency: 1, reaction: 1, resolution: 1, sagacity: 1 },
    knowledge: { biology: 0, erudition: 0, engineering: 0, geography: 0, medicine: 0, security: 0 },
    practices: { weapons: 0, athletics: 0, expression: 0, stealth: 0, crafting: 0, survival: 0 },
    tugOfWar: { determinationLevel: 9, determinationPoints: 9, assimilationLevel: 1, assimilationPoints: 0 },
    health: { damage: 0 },
    inventory: { body: [], backpack: [] },
    characteristics: [],
    xp: { total: 7 },
    assimilations: [],
    notes: '',
  }
}

export function newItem(): Item {
  return { name: '', quality: 0, scarcity: 0, pressure: 0, traits: [] }
}
