import { SYMBOL_COLORS } from '../dice3d/symbolIcons'
import { penaltyText, SYMBOL_GLYPHS, SYMBOL_NAMES, type RollResult } from './rolls'

const ORDER = ['success', 'adaptation', 'pressure'] as const

/** The roll's outcome at a glance: big symbol counts and the health penalty already applied. */
export function ResultBanner({ result }: { result: RollResult }) {
  const penalty = penaltyText(result)
  return (
    <div role="status" aria-label="Resultado da rolagem" className="flex flex-wrap items-center justify-center gap-x-6 gap-y-1">
      {ORDER.map((s) => (
        <span key={s} title={SYMBOL_NAMES[s]} className="flex items-baseline gap-2">
          <span aria-hidden className="text-3xl" style={{ color: SYMBOL_COLORS[s] }}>{SYMBOL_GLYPHS[s]}</span>
          <span className="text-4xl font-bold tabular-nums">{result.symbols[s]}</span>
          <span className="text-sm text-zinc-300">{SYMBOL_NAMES[s]}</span>
        </span>
      ))}
      {penalty && <span className="basis-full text-center text-sm font-medium text-amber-400">{penalty}</span>}
    </div>
  )
}
