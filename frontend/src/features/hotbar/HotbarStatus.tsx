import { CharacterAvatar } from '../miniature/CharacterAvatar'
import { playerColor } from '../tokens/tokens'
import type { CharacterView } from '../characters/characters'
import { penaltyLabel, type Vitals } from './vitals'

/** Avatar of a character: 2D bust from its appearance, else portrait URL, else initials. */
export function Portrait({ character }: { character: CharacterView }) {
  const look = character.sheet.appearance
  return <CharacterAvatar name={character.name} appearance={look} portraitUrl={character.sheet.identity?.portraitUrl}
    ring={look ? undefined : playerColor(character.ownerId)} size="h-12 w-12" />
}

/** Six segments, one per health level: red fills with the damage taken, the current level is outlined. */
export function HealthBar({ v }: { v: Vitals }) {
  const label = `Saúde: ${v.levelName}, ${v.damage} caixas de dano, ${penaltyLabel(v.penalty)}`
  return (
    <div className="min-w-[13rem] flex-1">
      <div className="flex items-baseline justify-between gap-2 whitespace-nowrap text-xs">
        <span className="font-medium text-zinc-300">Saúde</span>
        <span className={v.penalty > 0 ? 'text-amber-400' : 'text-zinc-400'}>{v.levelName}{v.penalty > 0 && ` · ${penaltyLabel(v.penalty)}`}</span>
      </div>
      <div role="img" aria-label={label} className="mt-1 flex gap-0.5">
        {v.filled.map((n, l) => (
          <span key={l} className={`h-2.5 flex-1 overflow-hidden rounded-sm bg-emerald-700/60 ${l === v.level ? 'ring-1 ring-white/80' : ''}`}>
            <span className="block h-full bg-red-500" style={{ width: `${(n / v.boxes) * 100}%` }} />
          </span>
        ))}
      </div>
    </div>
  )
}

/** Compact tug of war: Determinação against Assimilação, with the Suscetível/Debilitado state. */
export function TugBar({ v }: { v: Vitals }) {
  const t = v.tug
  if (!t) return null
  return (
    <div className="min-w-[12rem] flex-1">
      <div className="flex items-baseline justify-between gap-2 whitespace-nowrap text-xs">
        <span className="text-emerald-400">Determinação {t.determinationPoints}/{t.determinationLevel}</span>
        <span className="text-fuchsia-400">Assimilação {t.assimilationPoints}/{t.assimilationLevel}</span>
      </div>
      <div role="img" aria-label={`Cabo de guerra: Determinação ${t.determinationLevel}, Assimilação ${t.assimilationLevel}`} className="mt-1 flex gap-0.5">
        {Array.from({ length: 10 }, (_, i) => (
          <span key={i} className={`h-2.5 flex-1 rounded-sm ${i < t.determinationLevel ? 'bg-emerald-500' : 'bg-fuchsia-600'}`} />
        ))}
      </div>
    </div>
  )
}

export function StateBadges({ v }: { v: Vitals }) {
  const badge = 'rounded px-1.5 text-[11px] font-medium'
  return (
    <>
      {v.susceptible && <span className={`${badge} bg-red-900 text-red-200`}>Suscetível</span>}
      {v.incapacitated
        ? <span className={`${badge} bg-red-900 text-red-200`}>Incapacitado</span>
        : v.debilitated && <span className={`${badge} bg-yellow-900 text-yellow-200`}>Debilitado</span>}
    </>
  )
}
