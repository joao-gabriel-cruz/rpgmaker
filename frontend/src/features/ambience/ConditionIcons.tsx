import { autoConditions, conditionDef, type ConditionId, type SheetStatus } from './conditions'

interface Props {
  ids: ConditionId[]
  status?: SheetStatus | null
  size?: number
  className?: string
}

/** Small icon chips, one per condition; automatic ones say so in the tooltip. */
export function ConditionIcons({ ids, status, size = 14, className = '' }: Props) {
  if (ids.length === 0) return null
  const auto = autoConditions(status)
  return (
    <ul aria-label="Condições" className={`flex items-center gap-0.5 ${className}`}>
      {ids.map((id) => {
        const def = conditionDef(id)
        if (!def) return null
        const Icon = def.icon
        const tip = auto.includes(id) ? `${def.label} (automático)` : def.label
        return (
          <li key={id} title={tip} aria-label={tip} className={`flex items-center justify-center rounded-full bg-zinc-950/80 p-0.5 ring-1 ring-white/15 ${def.tone}`}>
            <Icon size={size} strokeWidth={2} aria-hidden />
          </li>
        )
      })}
    </ul>
  )
}
