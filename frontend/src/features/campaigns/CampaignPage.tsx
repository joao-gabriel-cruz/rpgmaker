import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useParams } from 'react-router'
import { api } from '../../lib/api'
import { useMe } from '../auth/auth'
import { CampaignSocketProvider } from '../realtime/CampaignSocketProvider'
import { MapArea } from '../scenes/MapArea'
import { OnlinePanel } from '../realtime/OnlinePanel'
import { RoundTracker } from '../rounds/RoundTracker'
import { CharactersSection } from '../characters/CharactersSection'
import { InvitePanel } from './InvitePanel'
import { SettingsPanel } from './SettingsPanel'
import { removeMember, roleLabel, type Campaign, type MemberInfo } from './campaigns'

function Members({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const { data: me } = useMe()
  const queryClient = useQueryClient()
  const key = ['campaigns', campaignId, 'members']
  const { data: members } = useQuery({
    queryKey: key,
    queryFn: () => api.get<MemberInfo[]>(`/campaigns/${campaignId}/members`),
  })
  const remove = useMutation({
    mutationFn: (userId: string) => removeMember(campaignId, userId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: key }),
  })

  return (
    <ul className="flex max-w-xl flex-col gap-2">
      {members?.map((m) => (
        <li key={m.userId} className="flex items-center gap-3 rounded bg-zinc-900 px-4 py-2">
          <span className="font-medium">{m.displayName}</span>
          <span className="text-sm text-zinc-400">{roleLabel(m.role)}</span>
          {isGm && m.userId !== me?.id && m.role === 'player' && (
            <button onClick={() => remove.mutate(m.userId)} aria-label={`Remover ${m.displayName}`}
              className="ml-auto rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-red-900">
              Remover
            </button>
          )}
        </li>
      ))}
    </ul>
  )
}

export function CampaignPage() {
  const { id = '' } = useParams()
  const { data: campaign, isPending, isError } = useQuery({
    queryKey: ['campaigns', id],
    queryFn: () => api.get<Campaign>(`/campaigns/${id}`),
    retry: false,
  })
  if (isPending) return <p className="text-zinc-400">Carregando…</p>
  if (isError) {
    return (
      <p role="alert" className="text-red-400">
        Campanha não encontrada. <Link to="/" className="underline">Voltar</Link>
      </p>
    )
  }
  return <CampaignView campaign={campaign} />
}

function CampaignView({ campaign }: { campaign: Campaign }) {
  const { id } = campaign
  const isGm = campaign.role === 'gm'
  return (
    <section className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{campaign.name}</h1>
        <p className="text-sm text-zinc-400">{campaign.systemName} · Você é {roleLabel(campaign.role)}</p>
      </div>
      <CampaignSocketProvider campaignId={id}>
        <RoundTracker campaignId={id} isGm={isGm} />
        <OnlinePanel />
        <MapArea campaignId={id} isGm={isGm} />
        <CharactersSection campaignId={id} />
      </CampaignSocketProvider>
      <div className="flex flex-col gap-2">
        <h2 className="text-xl font-semibold">Membros</h2>
        <Members campaignId={id} isGm={isGm} />
      </div>
      {isGm && <SettingsPanel campaign={campaign} />}
      {isGm && <InvitePanel campaignId={id} />}
    </section>
  )
}
