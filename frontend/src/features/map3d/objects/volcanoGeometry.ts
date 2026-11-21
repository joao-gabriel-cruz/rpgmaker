import { BufferGeometry, Color, Float32BufferAttribute } from 'three'

export const VOLCANO_RADIUS = 1.5
export const SEGMENTS = 14
export const VEIN_SEGMENTS = [1, 5, 8, 12]
const CRATER_Y = 1.72

// Rings from the crater center outward: [radius, height]. The third one is the rim.
const PROFILE: [number, number][] = [[0.0, CRATER_Y], [0.34, 1.8], [0.56, 2.2], [0.95, 1.6], [1.25, 0.85], [VOLCANO_RADIUS, 0]]

const hash = (a: number, b: number) => {
  const s = Math.sin(a * 127.1 + b * 311.7) * 43758.5453
  return s - Math.floor(s)
}

export type Ring = [number, number, number][]

/** Deterministic irregular rings; the footprint never exceeds the 3x3 square (radius <= 1.5). */
export function volcanoRings(): Ring[] {
  return PROFILE.map(([r, h], k) => {
    const ring: Ring = []
    for (let i = 0; i < SEGMENTS; i++) {
      const a = (i / SEGMENTS) * Math.PI * 2
      const jr = k === 0 ? 1 : 1 - hash(i, k) * (k === PROFILE.length - 1 ? 0.06 : 0.13)
      const jh = k === 0 || k === PROFILE.length - 1 ? 0 : (hash(k, i + 9) - 0.5) * 0.22
      ring.push([Math.cos(a) * r * jr, h + jh, Math.sin(a) * r * jr])
    }
    return ring
  })
}

const ASH = new Color('#3a2e29')
const ROCK = new Color('#5a4639')
const tint = (k: number, i: number) => ASH.clone().lerp(ROCK, Math.min(1, (PROFILE.length - 1 - k) / 3 + hash(i, k + 3) * 0.25))

/** Low-poly mountain body (flat shaded via the material) with per-vertex colors. */
export function buildBody(rings = volcanoRings()): BufferGeometry {
  const pos: number[] = []
  const col: number[] = []
  const push = (k: number, i: number) => {
    pos.push(...rings[k][i % SEGMENTS])
    col.push(...tint(k, i % SEGMENTS).toArray())
  }
  for (let k = 0; k < rings.length - 1; k++) {
    for (let i = 0; i < SEGMENTS; i++) {
      push(k, i); push(k + 1, i + 1); push(k + 1, i)
      if (k > 0) { push(k, i); push(k, i + 1); push(k + 1, i + 1) }
    }
  }
  const g = new BufferGeometry()
  g.setAttribute('position', new Float32BufferAttribute(pos, 3))
  g.setAttribute('color', new Float32BufferAttribute(col, 3))
  g.computeVertexNormals()
  return g
}

/** Flat lava ribbon draped over the slope from the rim down to the ring before the base. */
export function buildVein(segment: number, rings = volcanoRings()): BufferGeometry {
  const pos: number[] = []
  const idx: number[] = []
  for (let k = 1; k <= rings.length - 2; k++) {
    const [x, y, z] = rings[k][segment % SEGMENTS]
    const len = Math.hypot(x, z) || 1
    const w = 0.09 * (1 - (k - 1) * 0.12)
    const tx = -z / len * w, tz = x / len * w
    pos.push(x - tx, y + 0.035, z - tz, x + tx, y + 0.035, z + tz)
    if (k > 1) { const b = (k - 2) * 2; idx.push(b, b + 1, b + 2, b + 1, b + 3, b + 2) }
  }
  const g = new BufferGeometry()
  g.setAttribute('position', new Float32BufferAttribute(pos, 3))
  g.setIndex(idx)
  g.computeVertexNormals()
  return g
}

export const CRATER_HEIGHT = CRATER_Y
