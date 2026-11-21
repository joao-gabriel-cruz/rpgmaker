import { CampaignsList } from '../campaigns/CampaignsList'
import { useMe } from './auth'

export function HomePage() {
  const { data: user } = useMe()
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold tracking-tight">Olá, {user?.displayName}</h1>
      <CampaignsList />
    </div>
  )
}
