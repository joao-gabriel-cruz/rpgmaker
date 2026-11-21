import type { CharacterView } from '../characters/characters'
import { asNpc } from '../characters/npc'
import { boxesPerLevel, HEALTH_LEVELS } from '../sheet-assimilacao/derive'
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

/** Health and tug-of-war of a PC or NPC character, ready to render. */
export function vitalsOf(c: CharacterView): Vitals {
  const npc = c.kind === 'npc'
  const boxes = npc ? Math.max(1, asNpc(c.sheet).health.boxesPerLevel) : boxesPerLevel(c.sheet)
  const damage = c.sheet.health.damage
  const level = Math.max(0, Math.min(HEALTH_LEVELS - 1, Math.floor(damage / boxes)))
  const tug = npc ? undefined : c.sheet.tugOfWar
  return {
    level, levelName: HEALTH_LEVEL_NAMES[level], penalty: PENALTIES[level], boxes, damage,
    filled: Array.from({ length: HEALTH_LEVELS }, (_, l) => Math.max(0, Math.min(boxes, damage - l * boxes))),
    incapacitated: level >= 5, debilitated: level >= 4, tug, susceptible: !!tug && tug.determinationPoints === 0,
  }
}

export const penaltyLabel = (penalty: number) => (penalty > 0 ? `−${penalty} Sucesso${penalty > 1 ? 's' : ''}` : 'sem penalidade')
