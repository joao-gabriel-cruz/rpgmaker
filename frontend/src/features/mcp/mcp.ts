import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'

export interface LinkStatus { exists: boolean; createdAt?: string; lastUsedAt?: string | null }
export interface ClaudeChange { id: string; sceneId: string; summary: string; at: string }

export const getLink = (campaignId: string) => api.get<LinkStatus>(`/campaigns/${campaignId}/mcp-link`)
export const createLink = (campaignId: string) =>
  api.post<{ url: string; createdAt: string }>(`/campaigns/${campaignId}/mcp-link`)
export const revokeLink = (campaignId: string) => api.delete<void>(`/campaigns/${campaignId}/mcp-link`)
export const getChanges = (campaignId: string) =>
  api.get<{ changes: ClaudeChange[] }>(`/campaigns/${campaignId}/mcp-changes`)

/** Recent Claude changes, newest first, kept fresh by the `mcp.changed` event. */
export function useClaudeChanges(campaignId: string): ClaudeChange[] {
  const { subscribe } = useRealtime()
  const client = useQueryClient()
  const key = ['mcp-changes', campaignId]
  const query = useQuery({ queryKey: key, queryFn: () => getChanges(campaignId) })

  useEffect(() => subscribe('mcp.changed', (p) => {
    const { changes } = p as { changes: ClaudeChange[] }
    client.setQueryData(['mcp-changes', campaignId], { changes })
  }), [subscribe, client, campaignId])

  return query.data?.changes ?? []
}

/** Short PT-BR relative time: agora, há 3 min, há 2 h, há 4 d. */
export function relativeTime(iso: string, now = Date.now()): string {
  const minutes = Math.max(0, Math.floor((now - new Date(iso).getTime()) / 60000))
  if (minutes < 1) return 'agora'
  if (minutes < 60) return `há ${minutes} min`
  if (minutes < 1440) return `há ${Math.floor(minutes / 60)} h`
  return `há ${Math.floor(minutes / 1440)} d`
}
