import type { Coord } from '../map3d/grid'

/** Decodes the server bitset: cell (x,z) is bit z*width+x, least significant bit first. */
export function decodeBits(b64: string | undefined, width: number, depth: number): Uint8Array {
  const out = new Uint8Array(Math.ceil((width * depth) / 8))
  if (!b64) return out
  try {
    const raw = atob(b64)
    for (let i = 0; i < out.length && i < raw.length; i++) out[i] = raw.charCodeAt(i)
  } catch {
    return new Uint8Array(out.length)
  }
  return out
}

export function hasBit(bits: Uint8Array, width: number, x: number, z: number): boolean {
  const i = z * width + x
  return x >= 0 && z >= 0 && x < width && (bits[i >> 3] & (1 << (i & 7))) !== 0
}

/** Cells whose bit is `want` (false = covered by fog). */
export function cellsWhere(bits: Uint8Array, width: number, depth: number, want: boolean): Coord[] {
  const out: Coord[] = []
  for (let z = 0; z < depth; z++) {
    for (let x = 0; x < width; x++) if (hasBit(bits, width, x, z) === want) out.push({ x, z })
  }
  return out
}

/** Every cell of the rectangle spanned by two corner cells (inclusive). */
export function rectCells(a: Coord, b: Coord): Coord[] {
  const out: Coord[] = []
  for (let z = Math.min(a.z, b.z); z <= Math.max(a.z, b.z); z++) {
    for (let x = Math.min(a.x, b.x); x <= Math.max(a.x, b.x); x++) out.push({ x, z })
  }
  return out
}
