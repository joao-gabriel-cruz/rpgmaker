import { useMe } from '../auth/auth'
import { useCharacters } from '../characters/characters'
import { CrisesSection } from './CrisesSection'
import { ExpeditionsSection } from './ExpeditionsSection'
import type { Refuge } from './refuge'
import { ResidentsSection } from './ResidentsSection'

/** Crises, expeditions and notable residents of the Refúgio tab. */
export function RefugeLife({ campaignId, refuge, isGm }: { campaignId: string; refuge: Refuge; isGm: boolean }) {
  const { data: chars } = useCharacters(campaignId)
  const { data: me } = useMe()
  const everyone = Array.isArray(chars) ? chars : []
  const mine = everyone.filter((c) => c.kind === 'pc' && (isGm || c.ownerId === me?.id))
  return (
    <>
      <CrisesSection refuge={refuge} isGm={isGm} />
      <ExpeditionsSection refuge={refuge} isGm={isGm} userId={me?.id} candidates={mine} everyone={everyone} />
      <ResidentsSection refuge={refuge} isGm={isGm} everyone={everyone} />
    </>
  )
}
