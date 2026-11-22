import { CalendarClock } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import type { WeekReport } from './buildings'
import { ReportView } from './ReportView'
import type { Refuge } from './refuge'
import { WeekDialog } from './WeekDialog'
import { reportSummary } from './week'

function useWeekToast() {
  const { subscribe } = useRealtime()
  const [toast, setToast] = useState('')
  useEffect(() => subscribe('refuge.week.advanced', (p) => setToast(reportSummary((p as { report: WeekReport }).report))), [subscribe])
  useEffect(() => {
    if (!toast) return
    const t = setTimeout(() => setToast(''), 10000)
    return () => clearTimeout(t)
  }, [toast])
  return [toast, () => setToast('')] as const
}

function History({ history }: { history: WeekReport[] }) {
  if (!history.length) return <p className="text-sm text-zinc-500">Nenhuma semana registrada ainda.</p>
  return (
    <div className="flex flex-col gap-1" aria-label="Histórico de semanas">
      {[...history].reverse().map((r) => (
        <details key={r.week} className="rounded bg-zinc-900 px-3 py-2">
          <summary className="cursor-pointer text-sm text-zinc-200">Semana {r.week}{r.triggers.length ? ` · ${r.triggers.length} alerta(s)` : ''}</summary>
          <div className="mt-2"><ReportView report={r} /></div>
        </details>
      ))}
    </div>
  )
}

export function WeekPanel({ campaignId, refuge, isGm }: { campaignId: string; refuge: Refuge; isGm: boolean }) {
  const [open, setOpen] = useState(false)
  const [toast, dismiss] = useWeekToast()
  return (
    <section className="flex max-w-3xl flex-col gap-2">
      <div className="flex items-center gap-3">
        <h3 className="text-sm text-zinc-400">Semana {refuge.week}</h3>
        {isGm && (
          <button type="button" onClick={() => setOpen(true)}
            className="flex items-center gap-1 rounded bg-emerald-600 px-3 py-1.5 text-sm font-semibold hover:bg-emerald-500">
            <CalendarClock size={16} />Avançar semana</button>
        )}
      </div>
      {toast && (
        <div role="status" className="flex items-start justify-between gap-2 rounded bg-emerald-900/60 px-3 py-2 text-sm text-emerald-100">
          <span>{toast}</span><button type="button" aria-label="Dispensar aviso" onClick={dismiss}>×</button>
        </div>
      )}
      <History history={refuge.history} />
      {open && <WeekDialog campaignId={campaignId} onClose={() => setOpen(false)} />}
    </section>
  )
}
