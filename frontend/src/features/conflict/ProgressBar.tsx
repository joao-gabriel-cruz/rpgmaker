import { isComplete, progressPct, type Objective } from './conflict'

/** Progress toward an objective's Successes goal. */
export function ProgressBar({ objective }: { objective: Objective }) {
  const done = isComplete(objective)
  return (
    <div className="flex items-center gap-2">
      <div role="progressbar" aria-label={`Progresso: ${objective.title}`} aria-valuemin={0} aria-valuemax={objective.target}
        aria-valuenow={objective.progress} className="h-2 flex-1 overflow-hidden rounded bg-zinc-800">
        <div className={`h-full ${done ? 'bg-emerald-400' : 'bg-emerald-700'}`} style={{ width: `${progressPct(objective)}%` }} />
      </div>
      <span className="text-sm tabular-nums">{objective.progress}/{objective.target}</span>
    </div>
  )
}
