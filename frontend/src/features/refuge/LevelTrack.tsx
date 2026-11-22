import { MAX_LEVEL } from './refuge'
import { Stepper } from './Stepper'

interface Props {
  title: string
  icon?: React.ReactNode
  value: number
  labels: string[]
  isGm: boolean
  onChange: (value: number) => void
}

/** Card with a 7-step track (0-6) showing the current level label. */
export function LevelTrack({ title, icon, value, labels, isGm, onChange }: Props) {
  return (
    <div className="flex flex-col gap-2 rounded-lg bg-zinc-900 p-3" data-testid={`track-${title}`}>
      <div className="flex items-center gap-2 text-sm text-zinc-300">
        {icon}<span>{title}</span>
        <span className="ml-auto"><Stepper label={title} value={value} max={MAX_LEVEL} editable={isGm} onChange={onChange} /></span>
      </div>
      <div className="flex gap-1" role="presentation">
        {Array.from({ length: MAX_LEVEL + 1 }, (_, i) => (
          <span key={i} data-filled={i <= value && value > 0 ? 'true' : 'false'} title={labels[i]}
            className={`h-2 flex-1 rounded-full ${i <= value && value > 0 ? 'bg-emerald-500' : 'bg-zinc-700'} ${i === value ? 'ring-1 ring-emerald-300' : ''}`} />
        ))}
      </div>
      <p aria-label={`${title} atual`} className="text-lg font-semibold text-emerald-300">{labels[value] ?? '—'}</p>
    </div>
  )
}
