import { EMOTE_MS, isEmote, poseAt, REST_POSE, type EmoteName, type Pose } from '../miniature/emotes'

interface Playing {
  name: EmoteName
  start: number
  /** Direction (radians around y) the figure faces during the emote; null keeps its own. */
  heading: number | null
}

const playing = new Map<string, Playing>()

/** Faces from `from` towards `to` (figure +z is forward). */
export const headingTo = (from: { x: number; z: number }, to: { x: number; z: number }) => Math.atan2(to.x - from.x, to.z - from.z)

/** Starts an emote on a token; a new one replaces the running one. */
export function playEmote(tokenId: string, name: string, heading: number | null = null, now = performance.now()) {
  if (isEmote(name)) playing.set(tokenId, { name, start: now, heading })
}

export interface EmoteFrame {
  pose: Pose
  /** Extra rotation around y: faces the target during the emote and eases back at the ends. */
  turn: number
  active: boolean
}

const IDLE: EmoteFrame = { pose: REST_POSE, turn: 0, active: false }

/** The token's pose at `now`, or the idle frame when nothing plays (finished emotes are forgotten). */
export function emoteFrame(tokenId: string, now = performance.now()): EmoteFrame {
  const p = playing.get(tokenId)
  if (!p) return IDLE
  const t = (now - p.start) / EMOTE_MS[p.name]
  if (t >= 1) {
    playing.delete(tokenId)
    return IDLE
  }
  const ease = Math.min(1, t / 0.12, (1 - t) / 0.12)
  return { pose: poseAt(p.name, t), turn: (p.heading ?? 0) * Math.max(0, ease), active: true }
}

export const clearEmotes = () => playing.clear()
