/** Procedural poses for the miniature's emotes; `t` runs 0..1 over the emote's duration. Figure space: +z is forward. */
export interface Pose {
  /** Offsets of the whole figure (y up, z forward). */
  dy: number
  dz: number
  /** Rotations: lean forward (x), twist (y), tilt sideways (z), in radians. */
  lean: number
  twist: number
  tilt: number
  /** Vertical scale (crouch, kneel, sit). */
  squash: number
  /** Arm rotations as [swing around x, spread around z]. */
  armL: [number, number]
  armR: [number, number]
  /** 1 = solid, lower = ghostly. */
  opacity: number
  /** 0..1 red flash. */
  flash: number
}

export const REST_POSE: Pose = { dy: 0, dz: 0, lean: 0, twist: 0, tilt: 0, squash: 1, armL: [0, 0], armR: [0, 0], opacity: 1, flash: 0 }

export type EmoteName =
  | 'melee_strike' | 'ranged_shot' | 'look_around' | 'search' | 'sneak' | 'repair' | 'first_aid' | 'rest'
  | 'dance' | 'persuade' | 'intimidate' | 'hit_react'

export const EMOTE_MS: Record<EmoteName, number> = {
  melee_strike: 1200, ranged_shot: 1400, look_around: 1600, search: 1800, sneak: 1800, repair: 1800, first_aid: 1800,
  rest: 2400, dance: 2200, persuade: 1400, intimidate: 1400, hit_react: 550,
}

const TAU = Math.PI * 2
const clamp01 = (v: number) => Math.min(1, Math.max(0, v))
/** Smooth 0 -> 1 -> 0 bump between a and b (a..b rises over the first half). */
const bump = (t: number, a: number, b: number) => Math.sin(clamp01((t - a) / (b - a)) * Math.PI)
/** Eases in over [0, edge] and out over [1 - edge, 1]; 1 in between. */
const envelope = (t: number, edge = 0.15) => Math.min(1, t / edge, (1 - t) / edge)

type PosePart = Partial<Pose>

const make = (p: PosePart): Pose => ({ ...REST_POSE, ...p })

function meleeStrike(t: number): Pose {
  const wind = bump(t, 0, 0.4)
  const hit = bump(t, 0.3, 0.65)
  return make({ lean: -0.25 * wind + 0.55 * hit, dz: 0.32 * hit, armR: [-2.6 * wind + 3.2 * hit, 0.2], armL: [-0.4 * hit, 0] })
}

function rangedShot(t: number): Pose {
  const aim = envelope(t, 0.25)
  const recoil = bump(t, 0.5, 0.65)
  return make({ lean: 0.1 * aim, dz: -0.1 * recoil, armL: [-1.45 * aim, 0], armR: [-1.45 * aim, -0.1], twist: 0.2 * aim })
}

function lookAround(t: number): Pose {
  return make({ twist: Math.sin(t * TAU * 1.25) * 0.9 * envelope(t), lean: 0.05, armR: [-0.3 * envelope(t), 0] })
}

function search(t: number): Pose {
  const e = envelope(t, 0.2)
  const swing = Math.sin(t * TAU * 2.5) * 0.7
  return make({ squash: 1 - 0.3 * e, lean: 0.55 * e, armL: [-0.9 * e + swing * 0.5, 0], armR: [-0.9 * e - swing * 0.5, 0] })
}

function sneak(t: number): Pose {
  const e = envelope(t, 0.2)
  return make({ squash: 1 - 0.25 * e, lean: 0.3 * e, dz: 0.25 * t * e, opacity: 1 - 0.55 * e, armL: [0.4 * e, 0], armR: [0.4 * e, 0] })
}

function repair(t: number): Pose {
  const e = envelope(t, 0.2)
  const hammer = Math.sin(t * TAU * 3.5) * 0.9
  return make({ lean: 0.35 * e, squash: 1 - 0.1 * e, armR: [(-1.6 + hammer) * e, 0], armL: [-0.9 * e, 0] })
}

function firstAid(t: number): Pose {
  const e = envelope(t, 0.2)
  return make({ squash: 1 - 0.38 * e, lean: 0.25 * e, armL: [-1.1 * e, 0.2], armR: [-1.1 * e, -0.2] })
}

function rest(t: number): Pose {
  const e = envelope(t, 0.2)
  const nod = 0.25 * Math.sin(t * TAU * 1.2) * e
  return make({ squash: 1 - 0.42 * e, lean: 0.25 * e + nod, tilt: 0.12 * e, armL: [0.3 * e, 0], armR: [0.3 * e, 0] })
}

function dance(t: number): Pose {
  const e = envelope(t, 0.1)
  const beat = Math.sin(t * TAU * 4)
  return make({ dy: Math.abs(beat) * 0.12 * e, twist: Math.sin(t * TAU * 2) * 0.6 * e, tilt: beat * 0.12 * e,
    armL: [0, 2.3 * (0.5 + 0.5 * beat) * e], armR: [0, -2.3 * (0.5 - 0.5 * beat) * e] })
}

function persuade(t: number): Pose {
  const e = envelope(t, 0.2)
  const talk = Math.sin(t * TAU * 3)
  return make({ lean: 0.1 * e, armL: [(-1.0 + talk * 0.4) * e, 0.3], armR: [(-1.1 - talk * 0.4) * e, -0.3] })
}

function intimidate(t: number): Pose {
  const e = envelope(t, 0.2)
  const shake = Math.sin(t * TAU * 8) * 0.06
  return make({ lean: 0.4 * e, dz: 0.2 * bump(t, 0.2, 0.7), tilt: shake * e, armL: [-1.8 * e, 0.2], armR: [-1.8 * e, -0.2] })
}

function hitReact(t: number): Pose {
  const back = bump(t, 0, 0.6)
  const blink = Math.sin(t * TAU * 3) > 0 ? 1 : 0
  return make({ dz: -0.25 * back, lean: -0.3 * back, tilt: 0.15 * back, flash: blink * (1 - t) })
}

const POSES: Record<EmoteName, (t: number) => Pose> = {
  melee_strike: meleeStrike, ranged_shot: rangedShot, look_around: lookAround, search, sneak, repair, first_aid: firstAid, rest,
  dance, persuade, intimidate, hit_react: hitReact,
}

export const isEmote = (name: string): name is EmoteName => name in POSES

/** Pose of an emote at progress `t` (clamped to 0..1). */
export const poseAt = (name: EmoteName, t: number): Pose => POSES[name](clamp01(t))
