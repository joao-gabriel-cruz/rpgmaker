import { boxesPerLevel, currentLevel, isDebilitated, isIncapacitated, maxDamage, successPenalty, HEALTH_LEVELS } from './derive'
import { HEALTH_LEVEL_NAMES } from './labels'
import { Section, type SectionProps } from './ui'

interface RowProps {
  level: number
  boxes: number
  damage: number
  current: boolean
  readOnly: boolean
  onToggle: (n: number) => void
}

function LevelRow({ level, boxes, damage, current, readOnly, onToggle }: RowProps) {
  return (
    <li aria-current={current ? 'true' : undefined} className={`flex items-center gap-3 rounded px-2 py-1 ${current ? 'bg-emerald-950 ring-1 ring-emerald-600' : ''}`}>
      <span className="w-28 text-sm text-zinc-300">{HEALTH_LEVEL_NAMES[level]}</span>
      <span className="flex flex-wrap gap-1">
        {Array.from({ length: boxes }, (_, b) => {
          const n = level * boxes + b + 1
          const filled = n <= damage
          return (
            <button
              key={n}
              disabled={readOnly}
              aria-pressed={filled}
              aria-label={`Caixa de dano ${n}`}
              onClick={() => onToggle(n)}
              className={`h-5 w-5 rounded-sm border ${filled ? 'border-red-500 bg-red-600' : 'border-zinc-600 bg-zinc-900'} enabled:hover:border-emerald-500`}
            />
          )
        })}
      </span>
    </li>
  )
}

function Summary({ level, penalty }: { level: number; penalty: number }) {
  const text = penalty > 0 ? `−${penalty} Sucesso${penalty > 1 ? 's' : ''}` : 'sem penalidade'
  return <span className="text-sm" data-testid="health-summary">{HEALTH_LEVEL_NAMES[level]} · {text}</span>
}

export function HealthSection({ sheet, onChange, readOnly }: SectionProps) {
  const boxes = boxesPerLevel(sheet)
  const damage = sheet.health.damage
  const level = currentLevel(sheet)
  const toggle = (n: number) => {
    const next = damage === n ? n - 1 : n
    onChange({ ...sheet, health: { damage: Math.max(0, Math.min(maxDamage(sheet), next)) } })
  }
  return (
    <Section title="Saúde" aside={<Summary level={level} penalty={successPenalty(sheet)} />}>
      <ol className="space-y-1">
        {Array.from({ length: HEALTH_LEVELS }, (_, l) => (
          <LevelRow key={l} level={l} boxes={boxes} damage={damage} current={l === level} readOnly={readOnly} onToggle={toggle} />
        ))}
      </ol>
      {isIncapacitated(sheet) && <p className="mt-2 text-sm text-red-400">Incapacitado</p>}
      {isDebilitated(sheet) && !isIncapacitated(sheet) && <p className="mt-2 text-sm text-yellow-400">Debilitado — custo de Determinação</p>}
    </Section>
  )
}
