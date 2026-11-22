import { Biohazard } from 'lucide-react'
import { useRefuge } from './refuge'
import { useLiveRefuge } from './useLiveRefuge'

/** TopBar alert while the Refúgio has the Doença condition (-1 die on every test). */
export function DiseaseAlert({ campaignId }: { campaignId: string }) {
  const { data } = useRefuge(campaignId)
  useLiveRefuge(campaignId)
  if (!Array.isArray(data?.conditions) || !data.conditions.includes('disease')) return null
  return (
    <span role="status" aria-label="Doença no Refúgio" title="Doença no Refúgio: −1 dado em todos os testes até a cura"
      className="flex items-center gap-1 rounded bg-lime-900/60 px-2 py-1 text-xs font-medium text-lime-300">
      <Biohazard size={14} aria-hidden />Doença: −1 dado
    </span>
  )
}
