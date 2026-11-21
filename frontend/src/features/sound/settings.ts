import type { SoundSettings } from './engine'

const KEY = 'rpg.sound.v1'
export const DEFAULT_SOUND: SoundSettings = { volume: 0.6, muted: false }

/** Per-player volume and mute, kept in localStorage (storage may be blocked, so every access is guarded). */
export function loadSoundSettings(): SoundSettings {
  try {
    const raw = JSON.parse(localStorage.getItem(KEY) ?? 'null') as Partial<SoundSettings> | null
    const volume = typeof raw?.volume === 'number' && raw.volume >= 0 && raw.volume <= 1 ? raw.volume : DEFAULT_SOUND.volume
    return { volume, muted: raw?.muted === true }
  } catch {
    return DEFAULT_SOUND
  }
}

export function saveSoundSettings(s: SoundSettings): void {
  try {
    localStorage.setItem(KEY, JSON.stringify(s))
  } catch {
    // Storage unavailable: the setting just does not persist.
  }
}
