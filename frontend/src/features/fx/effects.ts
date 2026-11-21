import type { Animation } from '../actions/actions'

export type FxKind = 'arc' | 'projectile' | 'sparks' | 'wave' | 'rise' | 'confetti' | 'bubble' | 'number' | 'flash'

/** One visual effect on the map; positions are cell coordinates (center = +0.5). */
export interface Fx {
  id: number
  kind: FxKind
  x: number
  z: number
  /** Second point (projectile destination, arc direction). */
  toX?: number
  toZ?: number
  /** Milliseconds from creation until it shows (staggered bursts) and how long it lives. */
  delay: number
  ms: number
  glyph?: string
  color?: string
  /** Weapon style: slash | blunt | shot | arrow. */
  variant?: string
  seed: number
}

type Spec = Omit<Fx, 'id' | 'x' | 'z' | 'seed' | 'toX' | 'toZ'> & { atTarget?: boolean; toTarget?: boolean }

const SIGN = (glyph: string, color: string, ms = 1600, delay = 0): Spec => ({ kind: 'rise', glyph, color, ms, delay })

/** Effects each animation spawns at its start (the strike impact comes later with fx:hit_react). */
export const START_EFFECTS: Record<Animation, Spec[]> = {
  melee_strike: [{ kind: 'arc', ms: 350, delay: 330, toTarget: true }],
  ranged_shot: [{ kind: 'projectile', ms: 380, delay: 700, toTarget: true }],
  look_around: [{ kind: 'wave', ms: 1300, delay: 150, color: '#38bdf8' }],
  search: [SIGN('✦', '#fde047', 1700, 300), SIGN('✦', '#fef08a', 1500, 800)],
  sneak: [SIGN('…', '#a1a1aa', 1400, 200)],
  repair: [0, 500, 1000].map((delay): Spec => ({ kind: 'sparks', ms: 500, delay: delay + 200, color: '#fb923c', atTarget: true })),
  first_aid: [{ ...SIGN('✚', '#4ade80', 1700, 350), atTarget: true }],
  rest: [SIGN('z', '#93c5fd', 1700, 100), SIGN('Z', '#bfdbfe', 1700, 650), SIGN('z', '#dbeafe', 1700, 1200)],
  dance: [...['♪', '♫', '♪', '♬', '♫'].map((g, i): Spec => SIGN(g, ['#f472b6', '#facc15', '#60a5fa', '#4ade80', '#fb923c'][i], 1500, i * 330)),
    { kind: 'confetti', ms: 1500, delay: 700 }],
  persuade: [{ kind: 'bubble', glyph: '💬', ms: 1500, delay: 150 }],
  intimidate: [{ kind: 'bubble', glyph: '❗', color: '#dc2626', ms: 1500, delay: 150 }],
}

/** Static icon shown instead of the animation when the user prefers reduced motion. */
export const STATIC_GLYPH: Record<Animation, string> = {
  melee_strike: '⚔', ranged_shot: '🏹', look_around: '👁', search: '🔍', sneak: '🥷', repair: '🔨', first_aid: '✚', rest: '💤',
  dance: '🎵', persuade: '💬', intimidate: '❗',
}

export type { Spec as EffectSpec }
