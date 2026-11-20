import { Quaternion, Vector3 } from 'three'
import type { DieShape } from './polyhedra'
import type { Rng } from './prng'

export interface Pose {
  position: Vector3
  quaternion: Quaternion
}

export interface RollPlan {
  duration: number
  finalQuaternion: Quaternion
  sample: (t: number) => Pose
}

const UP = new Vector3(0, 1, 0)

// Rotation that points `normal` up, then spins `yaw` around the vertical axis.
export function finalQuaternion(normal: Vector3, yaw: number): Quaternion {
  const align = new Quaternion().setFromUnitVectors(normal.clone().normalize(), UP)
  return new Quaternion().setFromAxisAngle(UP, yaw).multiply(align)
}

// Drop segment then decaying bounces: [duration share, apex height].
const SEGMENTS: [number, number][] = [
  [0.4, 4.5],
  [0.28, 1.4],
  [0.2, 0.5],
  [0.12, 0.12],
]

function bounceHeight(u: number): number {
  let start = 0
  for (let i = 0; i < SEGMENTS.length; i++) {
    const [len, apex] = SEGMENTS[i]
    if (u <= start + len || i === SEGMENTS.length - 1) {
      const s = Math.min(1, (u - start) / len)
      return i === 0 ? apex * (1 - s * s) : apex * 4 * s * (1 - s)
    }
    start += len
  }
  return 0
}

function randomAxis(rng: Rng): Vector3 {
  const v = new Vector3(rng() - 0.5, rng() - 0.5, rng() - 0.5)
  return v.lengthSq() < 1e-6 ? new Vector3(1, 0, 0) : v.normalize()
}

export function planRoll(shape: DieShape, faceValue: number, slot: Vector3, rng: Rng): RollPlan {
  const face = shape.faces[faceValue - 1]
  if (!face) throw new Error(`invalid face ${faceValue}`)
  const rest = face.inradius
  const final = finalQuaternion(face.normal, rng() * Math.PI * 2)
  const axis = randomAxis(rng)
  const spin = (4 + rng() * 4) * Math.PI
  const start = new Vector3(slot.x - 9 - rng() * 2, 0, slot.z + (rng() - 0.5) * 4)
  const duration = 1.9 + rng() * 0.4
  const sample = (t: number): Pose => {
    const u = Math.min(1, Math.max(0, t / duration))
    const move = 1 - (1 - u) ** 2
    const spinLeft = (1 - u) ** 3
    const lift = (shape.radius - rest) * spinLeft
    const position = start.clone().lerp(slot, move)
    position.y = rest + lift + bounceHeight(u)
    const q = new Quaternion().setFromAxisAngle(axis, spin * spinLeft).multiply(final)
    return { position, quaternion: u >= 1 ? final.clone() : q }
  }
  return { duration, finalQuaternion: final, sample }
}
