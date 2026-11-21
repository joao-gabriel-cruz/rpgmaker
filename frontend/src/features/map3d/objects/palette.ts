// Colors chosen to sit next to the terrain palette.
export const C = {
  wood: '#8a5a3b',
  darkWood: '#5e3d27',
  trunk: '#6b4a2f',
  stone: '#7d8590',
  brick: '#8d8f94',
  metal: '#5f6368',
  dark: '#2a2d31',
  leaf: '#4f8f43',
  leafDark: '#3d7a37',
  rust: '#9a4b2d',
  glass: '#3a4a5c',
  fire: '#ff8a1f',
  lava: '#e8531c',
  ash: '#4a3b34',
  canvas: '#a89f80',
  sand: '#b9a57a',
  oldRust: '#7a3f25',
}

export type V3 = [number, number, number]

// Stall awning colors selectable through the `color` prop (a hex value is accepted too).
export const AWNING_COLORS: Record<string, string> = {
  red: '#b8453a', blue: '#3f6fa8', green: '#4f8f5a', yellow: '#d4a72c', orange: '#d9772b', purple: '#7b4f9a',
}
export const DEFAULT_AWNING = 'red'

export function awningColor(value: unknown): string {
  if (typeof value !== 'string') return AWNING_COLORS[DEFAULT_AWNING]
  return AWNING_COLORS[value] ?? (/^#[0-9a-f]{6}$/i.test(value) ? value : AWNING_COLORS[DEFAULT_AWNING])
}
