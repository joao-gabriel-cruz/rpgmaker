import { normalizeAppearance } from '../miniature/appearance'
import type { Token } from '../tokens/tokens'

/** Height of the name label's top edge: overlays stack from here so they never cover the miniature. */
export function tokenTop(t: Pick<Token, 'appearance'>): number {
  return t.appearance ? 1.35 * normalizeAppearance(t.appearance).body.height + 0.2 : 1.45
}
