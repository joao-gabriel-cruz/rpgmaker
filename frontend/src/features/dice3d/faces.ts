export type DieKind = 'd4' | 'd6' | 'd8' | 'd10' | 'd12' | 'd20'
export type Symbol = 'success' | 'adaptation' | 'pressure'
export type DieVariant = 'normal' | 'difficulty'

export interface DieResult {
  kind: DieKind
  face: number
  variant?: DieVariant
}

export type SymbolCounts = Record<Symbol, number>

const S: Symbol = 'success'
const A: Symbol = 'adaptation'
const P: Symbol = 'pressure'

// Index 0 is face 1 (spec 4.2).
const TABLE: Symbol[][] = [
  [],
  [],
  [P],
  [P],
  [A, P],
  [S],
  [S, S],
  [S, A],
  [S, A, P],
  [S, S, P],
  [S, A, A, P],
  [P, P],
]

const FACE_COUNT: Record<DieKind, number> = { d4: 4, d6: 6, d8: 8, d10: 10, d12: 12, d20: 20 }

/** Dice with numbers instead of Assimilação symbols (generic rolls). */
export const isNumbered = (kind: DieKind) => kind === 'd4' || kind === 'd8' || kind === 'd20'

export function faceCount(kind: DieKind): number {
  return FACE_COUNT[kind]
}

export function facesFor(kind: DieKind): Symbol[][] {
  if (isNumbered(kind)) return Array.from({ length: FACE_COUNT[kind] }, () => [])
  return TABLE.slice(0, FACE_COUNT[kind]).map((symbols) => [...symbols])
}

export function symbolsOf(kind: DieKind, face: number): Symbol[] {
  const faces = facesFor(kind)
  if (face < 1 || face > faces.length) throw new Error(`invalid face ${face} for ${kind}`)
  return faces[face - 1]
}

export function countSymbols(dice: DieResult[]): SymbolCounts {
  const counts: SymbolCounts = { success: 0, adaptation: 0, pressure: 0 }
  for (const die of dice) {
    for (const symbol of symbolsOf(die.kind, die.face)) counts[symbol]++
  }
  return counts
}
