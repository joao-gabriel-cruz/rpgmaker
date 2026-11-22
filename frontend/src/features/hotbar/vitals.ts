import type { CharacterView } from '../characters/characters'
import { HEALTH_LEVELS } from '../sheet-assimilacao/derive'
import { HEALTH_LEVEL_NAMES } from '../sheet-assimilacao/labels'
import type { TugOfWar } from '../sheet-assimilacao/types'

const PENALTIES = [0, 0, 1, 1, 2, 2]

export interface Vitals {
  level: number
  levelName: string
  penalty: number
  boxes: number
  damage: number
  /** Damage boxes filled in each of the six levels. */
  filled: number[]
  incapacitated: boolean
  debilitated: boolean
  tug?: TugOfWar
  susceptible: boolean
}

const num = (v: unknown, fallback: number) => (typeof v === 'number' && Number.isFinite(v) ? v : fallback)

/** Reads the tug of war of a possibly incomplete sheet; absent when the sheet has none. */
function tugOf(sheet: { tugOfWar?: Partial<TugOfWar> }): TugOfWar | undefined {
  const t = sheet.tugOfWar
  if (!t || typeof t !== 'object') return undefined
  return { determinationLevel: num(t.determinationLevel, 0), determinationPoints: num(t.determinationPoints, 0),
    assimilationLevel: num(t.assimilationLevel, 0), assimilationPoints: num(t.assimilationPoints, 0) }
}

/** Health and tug-of-war of a PC or NPC character, ready to render. */
export function vitalsOf(c: CharacterView): Vitals {
  const npc = c.kind === 'npc'
  const sheet = (c.sheet ?? {}) as unknown as Partial<{ health: Partial<{ damage: number; boxesPerLevel: number }>; instincts: Partial<{ potency: number; resolution: number }> }>
  const boxes = npc ? Math.max(1, num(sheet.health?.boxesPerLevel, 1)) : 1 + num(sheet.instincts?.potency, 0) + num(sheet.instincts?.resolution, 0)
  const damage = Math.max(0, num(sheet.health?.damage, 0))
  const level = Math.max(0, Math.min(HEALTH_LEVELS - 1, Math.floor(damage / boxes)))
  const tug = npc ? undefined : tugOf((c.sheet ?? {}) as { tugOfWar?: Partial<TugOfWar> })
  return {
    level, levelName: HEALTH_LEVEL_NAMES[level], penalty: PENALTIES[level], boxes, damage,
    filled: Array.from({ length: HEALTH_LEVELS }, (_, l) => Math.max(0, Math.min(boxes, damage - l * boxes))),
    incapacitated: level >= 5, debilitated: level >= 4, tug, susceptible: !!tug && tug.determinationPoints === 0,
  }
}

export const penaltyLabel = (penalty: number) => (penalty > 0 ? `−${penalty} Sucesso${penalty > 1 ? 's' : ''}` : 'sem penalidade')
