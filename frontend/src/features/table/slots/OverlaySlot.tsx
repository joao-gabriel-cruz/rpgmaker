import { RollTray } from '../../rolls/RollTray'

/**
 * Slot "overlay": centered over the map, above the hotbar. Holds the dice tray for now;
 * delivery 2 turns it into the translucent overlay. Children need `pointer-events-auto`.
 */
export function OverlaySlot() {
  return <RollTray />
}
