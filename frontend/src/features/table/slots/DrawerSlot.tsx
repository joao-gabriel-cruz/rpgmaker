import type { Campaign } from '../../campaigns/campaigns'
import { SheetDrawer } from '../../hotbar/SheetDrawer'

/** Slot "drawer": the character sheet, docked right over the map above the hotbar. */
export function DrawerSlot({ campaign }: { campaign: Campaign }) {
  return <SheetDrawer campaign={campaign} />
}
