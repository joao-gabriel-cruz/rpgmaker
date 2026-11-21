import { createContext, useContext } from 'react'
import type { CampaignSocketApi } from './useCampaignSocket'

export const RealtimeContext = createContext<CampaignSocketApi | null>(null)

/** The campaign socket of the nearest CampaignSocketProvider. */
export function useRealtime(): CampaignSocketApi {
  const api = useContext(RealtimeContext)
  if (!api) throw new Error('useRealtime must be used inside CampaignSocketProvider')
  return api
}
