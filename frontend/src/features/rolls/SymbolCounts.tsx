import { SYMBOL_COLORS } from '../dice3d/symbolIcons'
import { penaltyText, SYMBOL_GLYPHS, SYMBOL_NAMES, type RollResult } from './rolls'

const ORDER = ['success', 'adaptation', 'pressure'] as const

/** Sucessos / Adaptações / Pressões with their icons, plus the health penalty applied. */
export function SymbolCounts({ result }: { result: RollResult }) {
  const penalty = penaltyText(result)
  return (
    <div className="flex flex-wrap items-center gap-x-3 text-sm">
      {ORDER.map((s) => (
        <span key={s} title={SYMBOL_NAMES[s]}>
          <span aria-hidden style={{ color: SYMBOL_COLORS[s] }}>{SYMBOL_GLYPHS[s]}</span> {SYMBOL_NAMES[s]}: {result.symbols[s]}
        </span>
      ))}
      {penalty && <span className="text-amber-400">{penalty}</span>}
    </div>
  )
}
