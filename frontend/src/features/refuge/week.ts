import { api } from '../../lib/api'
import type { Flow, WeekReport } from './buildings'
import { RESOURCE_LABEL, type ResourceKey } from './refuge'

export const previewWeek = (campaignId: string, adjust: Flow) =>
  api.post<WeekReport>(`/campaigns/${campaignId}/refuge/week/preview`, { adjust })

export const confirmWeek = (campaignId: string, adjust: Flow) =>
  api.post<WeekReport>(`/campaigns/${campaignId}/refuge/week/confirm`, { adjust })

/** "3 Água, 2 Madeira" from a flow map; empty string when nothing moved. */
export const flowText = (f: Flow) =>
  (Object.entries(f) as [ResourceKey, number][]).filter(([, n]) => n !== 0).map(([k, n]) => `${n} ${RESOURCE_LABEL[k]}`).join(', ')

/** One-line summary used by the toast. */
export const reportSummary = (r: WeekReport) => {
  const parts = [`Semana ${r.week} concluída`]
  if (r.completed.length) parts.push(`obras: ${r.completed.join(', ')}`)
  if (r.triggers.length) parts.push(`${r.triggers.length} alerta(s): ${r.triggers.map((t) => t.detail).join('; ')}`)
  return parts.join(' · ')
}
