import { useEffect } from 'react'
import { LIGHT_MIN_DARKNESS, atmosphereFor, isWet, type Ambience } from '../ambience/ambience'
import { useSound } from './SoundContext'

/** Drives the continuous beds from the scene: rain on wet weather, crackle for a campfire at night, wind otherwise. */
export function useSceneSounds(ambience: Pick<Ambience, 'timeOfDay' | 'weather'>, campfireInScene: boolean) {
  const engine = useSound()?.engine
  const { timeOfDay, weather } = ambience
  useEffect(() => {
    const dark = atmosphereFor(timeOfDay, weather).darkness >= LIGHT_MIN_DARKNESS
    engine?.setBeds({
      rain: weather === 'storm' ? 'storm' : isWet(weather) ? 'rain' : 'none',
      fire: campfireInScene && dark,
      wind: weather === 'snow' || weather === 'fog' || weather === 'ash',
    })
    return () => engine?.setBeds({ rain: 'none', fire: false, wind: false })
  }, [engine, timeOfDay, weather, campfireInScene])
}
