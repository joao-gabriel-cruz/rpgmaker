export interface Vec3 {
  x: number
  y: number
  z: number
}

export const MIN_DISTANCE = 4
export const MAX_DISTANCE = 120
export const ZOOM_STEP = 0.8
/** Just short of straight down: OrbitControls misbehaves exactly at the pole. */
export const TOP_DOWN_POLAR = 0.02

const length = (v: Vec3) => Math.hypot(v.x, v.y, v.z)

/** Camera offset (from the orbit target) scaled by `factor`, kept within the zoom limits. */
export function zoomOffset(offset: Vec3, factor: number): Vec3 {
  const d = length(offset)
  if (d === 0) return offset
  const next = Math.min(MAX_DISTANCE, Math.max(MIN_DISTANCE, d * factor)) / d
  return { x: offset.x * next, y: offset.y * next, z: offset.z * next }
}

/** Offset turned around the vertical axis; +90 swings the camera to the right of the target. */
export function rotateOffset(offset: Vec3, degrees: number): Vec3 {
  const a = (degrees * Math.PI) / 180
  const cos = Math.cos(a)
  const sin = Math.sin(a)
  return { x: offset.x * cos + offset.z * sin, y: offset.y, z: -offset.x * sin + offset.z * cos }
}

/** Same distance, camera straight above the target (slightly tilted along the old azimuth). */
export function topDownOffset(offset: Vec3): Vec3 {
  const d = length(offset) || 1
  const flat = Math.hypot(offset.x, offset.z)
  const [dx, dz] = flat === 0 ? [0, 1] : [offset.x / flat, offset.z / flat]
  const r = d * Math.sin(TOP_DOWN_POLAR)
  return { x: dx * r, y: d * Math.cos(TOP_DOWN_POLAR), z: dz * r }
}

export const isTopDown = (offset: Vec3) => {
  const d = length(offset)
  return d > 0 && offset.y / d > Math.cos(0.1)
}
