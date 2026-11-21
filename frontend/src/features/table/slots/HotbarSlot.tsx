import { Hotbar } from '../../hotbar/Hotbar'

/** Slot "hotbar": bottom-center of the map. The wrapper is pointer-events-none; the hotbar re-enables it. */
export function HotbarSlot({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  return <Hotbar campaignId={campaignId} isGm={isGm} />
}
