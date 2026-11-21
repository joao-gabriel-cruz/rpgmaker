import { type ReactNode } from 'react'
import { RealtimeContext } from './realtimeContext'
import { useCampaignSocket } from './useCampaignSocket'

/** Opens the campaign socket for its subtree; screens read it with useRealtime(). */
export function CampaignSocketProvider({ campaignId, children }: { campaignId: string; children: ReactNode }) {
  const api = useCampaignSocket(campaignId)
  return <RealtimeContext.Provider value={api}>{children}</RealtimeContext.Provider>
}
