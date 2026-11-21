import { useQuery } from '@tanstack/react-query'
import { api } from '../../lib/api'

export type Role = 'gm' | 'player'

export type Campaign = {
  id: string
  name: string
  systemId: string
  systemName: string
  role: Role
}

export type SystemInfo = { id: string; name: string; version: string }
export type MemberInfo = { userId: string; displayName: string; role: Role }
export type InviteInfo = { token: string; campaignName: string; expiresAt: string }

export const roleLabel = (role: Role) => (role === 'gm' ? 'Mestre' : 'Jogador')

export const campaignsKey = ['campaigns'] as const

export const useCampaigns = () =>
  useQuery({ queryKey: campaignsKey, queryFn: () => api.get<Campaign[]>('/campaigns') })

export const useSystems = () =>
  useQuery({ queryKey: ['systems'], queryFn: () => api.get<SystemInfo[]>('/systems') })

export const createCampaign = (name: string, systemId: string) =>
  api.post<Campaign>('/campaigns', { name, systemId })

export const createInvite = (campaignId: string) =>
  api.post<InviteInfo>(`/campaigns/${campaignId}/invites`)

export const removeMember = (campaignId: string, userId: string) =>
  api.delete<void>(`/campaigns/${campaignId}/members/${userId}`)

export const inviteLink = (token: string) => `${window.location.origin}/invite/${token}`
