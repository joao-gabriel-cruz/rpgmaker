interface ItemLike {
  name?: string
  icon?: string
}

interface InventoryLike {
  inventory?: { body?: ItemLike[]; backpack?: ItemLike[] }
}

const LIGHT_NAME = /\b(lanterna|tocha)s?\b/i
const LIGHT_ICONS = ['flashlight', 'flame']

/**
 * Whether a character carries a light: any body/backpack item whose name contains "lanterna" or "tocha"
 * (whole word, any case) or whose icon is `flashlight` or `flame`. Works on a full sheet or a token status.
 */
export function hasLightSource(sheet: InventoryLike | null | undefined): boolean {
  const inv = sheet?.inventory
  return [...(inv?.body ?? []), ...(inv?.backpack ?? [])].some(
    (i) => LIGHT_NAME.test(i.name ?? '') || LIGHT_ICONS.includes((i.icon ?? '').toLowerCase()),
  )
}
