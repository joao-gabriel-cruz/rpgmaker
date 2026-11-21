import type { SoundName } from './synth'

/** Sound for each `fx:<animation>` event of the actions contract (and the extra `fx:arrow`). */
export const FX_SOUNDS: Record<string, SoundName> = {
  'fx:melee_strike': 'hit',
  'fx:ranged_shot': 'shot',
  'fx:arrow': 'arrow',
  'fx:hit_react': 'hit',
  'fx:look_around': 'chime',
  'fx:search': 'rustle',
  'fx:sneak': 'step',
  'fx:repair': 'clink',
  'fx:first_aid': 'chime',
  'fx:rest': 'zzz',
  'fx:dance': 'cheer',
  'fx:persuade': 'murmur',
  'fx:intimidate': 'murmur',
}
