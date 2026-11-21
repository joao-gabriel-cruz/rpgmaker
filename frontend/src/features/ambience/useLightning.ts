import { useEffect, useRef } from 'react'
import { fxBus } from '../../lib/fxBus'
import type { Weather } from './ambience'

/**
 * Random lightning while it storms: `flash` (0..1) decays in the scene's frame loop, and an
 * `ambience:lightning` event reaches the sound engine. With reduced motion nothing flashes.
 */
export function useLightning(weather: Weather, reduced: boolean) {
  const flash = useRef(0)
  useEffect(() => {
    if (weather !== 'storm') return
    const timers: ReturnType<typeof setTimeout>[] = []
    const strike = () => {
      if (!reduced) {
        flash.current = 1
        timers.push(setTimeout(() => { flash.current = 0.8 }, 170))
      }
      fxBus.emit('ambience:lightning', {})
      timers.push(setTimeout(strike, 7000 + Math.random() * 11000))
    }
    timers.push(setTimeout(strike, 2500 + Math.random() * 4000))
    return () => timers.forEach(clearTimeout)
  }, [weather, reduced])
  return flash
}
