import { APTITUDE_LABELS, INSTINCT_LABELS } from '../sheet-assimilacao/labels'
import type { AptitudeKey, InstinctKey } from '../sheet-assimilacao/types'

export type Visibility = 'public' | 'gm'

export interface RolledDie {
  kind: string
  face: number
  difficulty: boolean
}

export interface RollResult {
  symbols: { success: number; adaptation: number; pressure: number }
  kept: number[]
  penalty: number
  assimilated: boolean
  healthLevel?: string
}

/** A roll as sent in `roll.rolled` and by the log endpoint. */
export interface RollEntry {
  rollId: string
  userId: string
  userName: string
  characterId: string | null
  characterName?: string
  kind: 'assimilacao' | 'generic' | 'threat'
  visibility: Visibility
  status?: 'pending' | 'final'
  dice: RolledDie[]
  keptCount?: number
  suggested?: number[]
  request?: {
    instinct: InstinctKey
    aptitude?: AptitudeKey
    secondInstinct?: InstinctKey
    effort?: boolean
    useOccupation?: boolean
    useMarkingEvent?: boolean
    difficultyD12?: number
  }
  expression?: string
  total?: number
  result?: RollResult
  seed: number
  createdAt?: string
}

export interface FinalizedEvent {
  rollId: string
  result: RollResult
}

const countOf = (e: RollEntry, kind: string) => e.dice.filter((d) => d.kind === kind && !d.difficulty).length

/** "Reação 2 + Atletismo 1" (or the expression of a generic roll). */
export function poolLabel(e: RollEntry): string {
  if (e.kind === 'generic' || e.kind === 'threat') return e.expression ?? ''
  const { instinct, aptitude } = e.request ?? {}
  const parts: string[] = []
  if (instinct) parts.push(`${INSTINCT_LABELS[instinct]} ${countOf(e, e.dice.some((d) => d.kind === 'd12' && !d.difficulty) ? 'd12' : 'd6')}`)
  if (aptitude) parts.push(`${APTITUDE_LABELS[aptitude]} ${countOf(e, 'd10')}`)
  return parts.join(' + ')
}

export const SYMBOL_GLYPHS = { success: '✔', adaptation: '◆', pressure: '✖' } as const
export const SYMBOL_NAMES = { success: 'Sucessos', adaptation: 'Adaptações', pressure: 'Pressões' } as const

export function penaltyText(r: RollResult): string {
  if (r.penalty <= 0) return ''
  const level = r.healthLevel ? ` (${r.healthLevel})` : ''
  return `−${r.penalty} ${r.penalty > 1 ? 'Sucessos' : 'Sucesso'}${level}`
}
