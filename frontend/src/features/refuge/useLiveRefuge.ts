import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { refugeKey, type Refuge } from './refuge'

/** Keeps the refuge cache live from refuge.updated. */
export function useLiveRefuge(campaignId: string) {
  const queryClient = useQueryClient()
  const { subscribe } = useRealtime()
  useEffect(() => subscribe('refuge.updated', (p) => {
    queryClient.setQueryData(refugeKey(campaignId), (p as { refuge: Refuge }).refuge)
  }), [subscribe, queryClient, campaignId])
}
