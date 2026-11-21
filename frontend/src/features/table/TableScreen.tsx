import type { Campaign } from '../campaigns/campaigns'
import { ConflictProvider } from '../conflict/ConflictContext'
import { CampaignSocketProvider } from '../realtime/CampaignSocketProvider'
import { MapArea } from '../scenes/MapArea'
import { useSceneControls } from '../scenes/useSceneControls'
import { HotbarSlot } from './slots/HotbarSlot'
import { OverlaySlot } from './slots/OverlaySlot'
import { ToastSlot } from './slots/ToastSlot'
import { TableLayout } from './TableLayout'

function TableContent({ campaign }: { campaign: Campaign }) {
  const isGm = campaign.role === 'gm'
  const scenes = useSceneControls(campaign.id, isGm)
  return (
    <TableLayout campaign={campaign} scenes={scenes}
      map={<MapArea campaignId={campaign.id} isGm={isGm} scenes={scenes} />}
      hotbar={<HotbarSlot campaignId={campaign.id} isGm={isGm} />}
      overlay={<OverlaySlot />}
      toasts={<ToastSlot />} />
  )
}

/** The campaign table: realtime + conflict providers around the full-screen layout. */
export function TableScreen({ campaign }: { campaign: Campaign }) {
  return (
    <CampaignSocketProvider campaignId={campaign.id}>
      <ConflictProvider campaignId={campaign.id} isGm={campaign.role === 'gm'}>
        <TableContent campaign={campaign} />
      </ConflictProvider>
    </CampaignSocketProvider>
  )
}
