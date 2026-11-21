import { PresenceDock } from '../ambience/PresenceDock'
import { PresenceProvider } from '../ambience/PresenceContext'
import { SoundProvider } from '../sound/SoundContext'
import type { Campaign } from '../campaigns/campaigns'
import { ConflictProvider } from '../conflict/ConflictContext'
import { CampaignSocketProvider } from '../realtime/CampaignSocketProvider'
import { MapArea } from '../scenes/MapArea'
import { useSceneControls } from '../scenes/useSceneControls'
import { TableCharacterProvider } from '../hotbar/TableCharacterContext'
import { DrawerSlot } from './slots/DrawerSlot'
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
      drawer={<DrawerSlot campaign={campaign} />}
      overlay={<OverlaySlot />}
      presence={<PresenceDock campaignId={campaign.id} isGm={isGm} />}
      toasts={<ToastSlot />} />
  )
}

/** The campaign table: realtime + conflict providers around the full-screen layout. */
export function TableScreen({ campaign }: { campaign: Campaign }) {
  return (
    <CampaignSocketProvider campaignId={campaign.id}>
      <ConflictProvider campaignId={campaign.id} isGm={campaign.role === 'gm'}>
        <SoundProvider><PresenceProvider>
          <TableCharacterProvider><TableContent campaign={campaign} /></TableCharacterProvider>
        </PresenceProvider></SoundProvider>
      </ConflictProvider>
    </CampaignSocketProvider>
  )
}
