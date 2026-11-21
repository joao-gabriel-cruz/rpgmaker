import { HEALTH_LEVEL_NAMES } from './labels'
import type { Sheet } from './types'

const PENALTIES = [0, 0, 1, 1, 2, 2]
export const HEALTH_LEVELS = 6

export function boxesPerLevel(sheet: Sheet): number {
  return 1 + sheet.instincts.potency + sheet.instincts.resolution
}

export function currentLevel(sheet: Sheet): number {
  const level = Math.floor(sheet.health.damage / boxesPerLevel(sheet))
  return Math.max(0, Math.min(HEALTH_LEVELS - 1, level))
}

export function currentLevelName(sheet: Sheet): string {
  return HEALTH_LEVEL_NAMES[currentLevel(sheet)]
}

export function successPenalty(sheet: Sheet): number {
  return PENALTIES[currentLevel(sheet)]
}

export function isDebilitated(sheet: Sheet): boolean {
  return currentLevel(sheet) >= 4
}

export function isIncapacitated(sheet: Sheet): boolean {
  return currentLevel(sheet) >= 5
}

export function maxDamage(sheet: Sheet): number {
  return boxesPerLevel(sheet) * HEALTH_LEVELS
}

export function xpSpent(sheet: Sheet): number {
  return sheet.characteristics.reduce((sum, c) => sum + 2 * c.level, 0)
}

export function xpAvailable(sheet: Sheet): number {
  return sheet.xp.total - xpSpent(sheet)
}

export function sumValues(values: Record<string, number>): number {
  return Object.values(values).reduce((a, b) => a + b, 0)
}
