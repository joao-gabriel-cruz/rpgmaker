import { useEffect, useState } from 'react'
import { Modal } from '../table/Modal'
import type { Flow, WeekReport } from './buildings'
import { ReportView } from './ReportView'
import { RESOURCE_LABEL, type ResourceKey } from './refuge'
import { Stepper } from './Stepper'
import { confirmWeek, previewWeek } from './week'

const KEYS = Object.keys(RESOURCE_LABEL) as ResourceKey[]

function Adjustments({ adjust, onChange }: { adjust: Flow; onChange: (k: ResourceKey, n: number) => void }) {
  return (
    <details className="text-sm">
      <summary className="cursor-pointer text-zinc-300">Ajustar números (somados antes do consumo)</summary>
      <div className="mt-2 grid grid-cols-2 gap-1">
        {KEYS.map((k) => (
          <div key={k} className="flex items-center justify-between rounded bg-zinc-800 px-2 py-1">{RESOURCE_LABEL[k]}
            <Stepper label={`Ajuste de ${RESOURCE_LABEL[k]}`} value={adjust[k] ?? 0} min={-99} max={99} editable onChange={(n) => onChange(k, n)} />
          </div>
        ))}
      </div>
    </details>
  )
}

/** Mestre flow: preview (nothing is written) → adjust numbers → confirm. */
export function WeekDialog({ campaignId, onClose }: { campaignId: string; onClose: () => void }) {
  const [adjust, setAdjust] = useState<Flow>({})
  const [report, setReport] = useState<WeekReport | null>(null)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  useEffect(() => {
    let live = true
    previewWeek(campaignId, adjust).then((r) => live && (setReport(r), setError(''))).catch((e: Error) => live && setError(e.message))
    return () => { live = false }
  }, [campaignId, adjust])
  const confirm = () => {
    setBusy(true)
    confirmWeek(campaignId, adjust).then(onClose).catch((e: Error) => { setError(e.message); setBusy(false) })
  }
  return (
    <Modal title="Avançar semana" onClose={onClose} wide>
      <div className="flex flex-col gap-3">
        <p className="text-sm text-zinc-400">Prévia: nada foi gravado ainda. Revise, ajuste se quiser e confirme.</p>
        {report && <><h3 className="font-semibold text-emerald-300">Semana {report.week}</h3><ReportView report={report} /></>}
        <Adjustments adjust={adjust} onChange={(k, n) => setAdjust((a) => ({ ...a, [k]: n }))} />
        {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
        <button type="button" disabled={!report || busy} onClick={confirm}
          className="self-end rounded bg-emerald-600 px-4 py-2 text-sm font-semibold hover:bg-emerald-500 disabled:opacity-50">Confirmar semana</button>
      </div>
    </Modal>
  )
}
