import type { AptitudeKey, InstinctKey, Sheet } from '../sheet-assimilacao/types'
import type { Modifiers } from './modifiers'

export interface PoolSelection {
  instinct?: InstinctKey
  aptitude?: AptitudeKey
}

export interface PoolPreview {
  /** Dice of the pool, in roll order; the difficulty d12s are separate. */
  parts: { count: number; kind: 'd6' | 'd10' | 'd12' }[]
  difficulty: number
  kept: number
  text: string
}

export const aptitudeValue = (sheet: Sheet, key: AptitudeKey): number =>
  (sheet.knowledge as Record<string, number>)[key] ?? (sheet.practices as Record<string, number>)[key] ?? 0

/** Dice of the pool for the instinct kind: d6, or d12 when the roll is assimilated. */
export const instinctDieKind = (m: Modifiers) => (m.assimilated && m.secondInstinct ? 'd12' : 'd6')

/** Mirrors the server's BuildRoll/keptCount to show "2 d6 + 1 d10 · manter 2" before rolling. */
export function poolPreview(sheet: Sheet, sel: PoolSelection, m: Modifiers, isGm: boolean, bonusKept = 0): PoolPreview {
  const assimilated = !!(m.assimilated && m.secondInstinct)
  const parts: PoolPreview['parts'] = []
  if (sel.instinct) parts.push({ count: sheet.instincts[sel.instinct], kind: instinctDieKind(m) })
  if (assimilated && m.secondInstinct) parts.push({ count: sheet.instincts[m.secondInstinct], kind: 'd12' })
  if (sel.aptitude) parts.push({ count: aptitudeValue(sheet, sel.aptitude), kind: 'd10' })
  const shown = parts.filter((p) => p.count > 0)
  const pool = shown.reduce((n, p) => n + p.count, 0)
  const bonus = [m.effort, m.useOccupation, m.useMarkingEvent].filter(Boolean).length + bonusKept
  const kept = Math.min(Math.max(1 + bonus, assimilated ? 2 : 0), pool)
  const difficulty = isGm ? m.difficultyD12 : 0
  const dice = shown.map((p) => `${p.count} ${p.kind}`).join(' + ') || 'sem dados'
  const hard = difficulty > 0 ? ` (+${difficulty} d12 dificuldade)` : ''
  return { parts: shown, difficulty, kept, text: `${dice}${hard} · manter ${kept}` }
}
