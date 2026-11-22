import { TriangleAlert } from 'lucide-react'
import type { WeekReport } from './buildings'
import { flowText } from './week'

function Row({ label, text, tone }: { label: string; text: string; tone?: string }) {
  if (!text) return null
  return <p className="text-sm"><span className="text-zinc-400">{label}: </span><span className={tone ?? 'text-zinc-100'}>{text}</span></p>
}

function Works({ report }: { report: WeekReport }) {
  if (!report.works.length) return null
  return (
    <ul className="flex flex-col gap-0.5 text-sm text-zinc-300" aria-label="Obras da semana">
      {report.works.map((w) => (
        <li key={w.buildingId}>
          {w.name} {w.level > 1 ? `(nível ${w.level}) ` : ''}+{w.points} PO → {w.progress}/{w.cost}
          {w.done && <span className="text-emerald-400"> concluída</span>}
          {w.blocked && <span className="text-amber-400"> bloqueada: {w.blocked}</span>}
        </li>
      ))}
    </ul>
  )
}

/** Readable week report: what came in, went out, was missing, finished works and crisis triggers. */
export function ReportView({ report }: { report: WeekReport }) {
  return (
    <div className="flex flex-col gap-1" data-testid={`report-${report.week}`}>
      <Row label="Entrou" text={flowText(report.entered)} tone="text-emerald-300" />
      <Row label="Saiu" text={flowText(report.spent)} />
      <Row label="Faltou" text={flowText(report.missing)} tone="text-red-400" />
      <Row label="Perdido (acima do teto)" text={flowText(report.lost)} tone="text-amber-400" />
      <Row label="Ajustes do Mestre" text={flowText(report.adjustments)} />
      <Works report={report} />
      <Row label="Obras concluídas" text={report.completed.join(', ')} tone="text-emerald-400" />
      {report.notes.map((n) => <p key={n} className="text-xs text-zinc-400">{n}</p>)}
      {report.triggers.map((t) => (
        <p key={t.detail} role="alert" className="flex items-center gap-1 text-sm text-red-400"><TriangleAlert size={14} />{t.detail}</p>
      ))}
    </div>
  )
}
