import { useQuery } from '@tanstack/react-query'
import { useCallback } from 'react'
import { Link, useNavigate, useParams } from 'react-router'
import { api } from '../../lib/api'
import type { Campaign } from '../campaigns/campaigns'
import { CampaignSocketProvider } from '../realtime/CampaignSocketProvider'
import { RollTray } from '../rolls/RollTray'
import { CharacterEditor } from './CharacterEditor'

function BackLink({ id }: { id: string }) {
  return <Link to={`/campaigns/${id}`} className="text-sm text-zinc-400 underline">Voltar à campanha</Link>
}

function Page({ campaign, charId }: { campaign: Campaign; charId: string }) {
  const navigate = useNavigate()
  const gone = useCallback(() => navigate(`/campaigns/${campaign.id}`), [navigate, campaign.id])
  return (
    <>
      <CharacterEditor campaign={campaign} characterId={charId} links={<BackLink id={campaign.id} />} onGone={gone} />
      <div className="pointer-events-none fixed inset-0 z-40 flex items-center justify-center"><RollTray /></div>
    </>
  )
}

export function CharacterPage() {
  const { id = '', charId = '' } = useParams()
  const campaign = useQuery({ queryKey: ['campaigns', id], queryFn: () => api.get<Campaign>(`/campaigns/${id}`), retry: false })
  if (campaign.isError) {
    return <p role="alert" className="text-red-400">Personagem não encontrado. <Link to={`/campaigns/${id}`} className="underline">Voltar</Link></p>
  }
  if (!campaign.data) return <p className="text-zinc-400">Carregando…</p>
  return (
    <CampaignSocketProvider campaignId={id}>
      <Page campaign={campaign.data} charId={charId} />
    </CampaignSocketProvider>
  )
}
