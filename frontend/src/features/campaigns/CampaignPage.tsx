import { useQuery } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { Link, useParams } from 'react-router'
import { api } from '../../lib/api'
import { TableScreen } from '../table/TableScreen'
import type { Campaign } from './campaigns'

function Notice({ children }: { children: ReactNode }) {
  return <div className="flex min-h-dvh items-center justify-center bg-zinc-950 p-6 text-zinc-100">{children}</div>
}

/** Campaign route: loads the campaign and opens the full-screen table. */
export function CampaignPage() {
  const { id = '' } = useParams()
  const { data: campaign, isPending, isError } = useQuery({
    queryKey: ['campaigns', id],
    queryFn: () => api.get<Campaign>(`/campaigns/${id}`),
    retry: false,
  })
  if (isPending) return <Notice><p className="text-zinc-400">Carregando…</p></Notice>
  if (isError) {
    return (
      <Notice>
        <p role="alert" className="text-red-400">
          Campanha não encontrada. <Link to="/" className="underline">Voltar</Link>
        </p>
      </Notice>
    )
  }
  return <TableScreen campaign={campaign} />
}
