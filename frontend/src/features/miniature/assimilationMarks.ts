import { MAX_INFECTED } from './appearance'

export interface MarkLevel {
  /** How many vein/blotch marks are drawn. */
  count: number
  /** 0-1 glow strength. */
  intensity: number
}

export const MAX_MARKS = 14

/** Level 1 (the starting level) is clean; marks grow in number and glow up to level 10. */
export function marksFor(level: number | null | undefined): MarkLevel {
  const l = Math.min(MAX_INFECTED, Math.max(0, Math.round(level ?? 0)))
  if (l <= 1) return { count: 0, intensity: 0 }
  return { count: Math.round((l - 1) * 1.5), intensity: 0.2 + ((l - 2) / (MAX_INFECTED - 2)) * 0.8 }
}

/** The sheet's Assimilation level wins; NPCs fall back to the optional `appearance.infected`. */
export const effectiveLevel = (assimilationLevel: number | null | undefined, infected: number | undefined) =>
  assimilationLevel ?? infected ?? 0
