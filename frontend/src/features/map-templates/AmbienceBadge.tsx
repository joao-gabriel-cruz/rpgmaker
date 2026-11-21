import { TIMES, WEATHERS } from '../ambience/ambience'
import type { TemplateInfo } from './templates'

/** Suggested time of day and weather as two small icons with tooltips. */
export function AmbienceBadge({ ambience }: { ambience: TemplateInfo['ambience'] }) {
  const time = TIMES.find((t) => t.id === ambience.timeOfDay)
  const weather = WEATHERS.find((w) => w.id === ambience.weather)
  const label = `Ambiente sugerido: ${time?.label ?? ''}, ${weather?.label ?? ''}`
  return (
    <span title={label} aria-label={label} role="img" className="flex items-center gap-1 text-zinc-300">
      {time && <time.icon size={14} aria-hidden />}
      {weather && <weather.icon size={14} aria-hidden />}
    </span>
  )
}
