import { useEffect, useSyncExternalStore } from 'react'
import { useRealtime } from '../realtime/realtimeContext'

const KEY = 'rpg:spotlight'
const listeners = new Set<() => void>()

const read = () => typeof localStorage !== 'undefined' && localStorage.getItem(KEY) === '1'
let on = read()

/** Turn-spotlight switch of the Mestre (kept in this browser); read by the top bar toggle and the driver. */
export const spotlight = {
  get: () => on,
  set(v: boolean) {
    on = v
    if (typeof localStorage !== 'undefined') localStorage.setItem(KEY, v ? '1' : '0')
    listeners.forEach((l) => l())
  },
  subscribe(l: () => void) {
    listeners.add(l)
    return () => listeners.delete(l)
  },
}

export const useSpotlight = () => useSyncExternalStore(spotlight.subscribe, spotlight.get)

/** While the spotlight is on, each `action.started` makes every camera glide to the actor's token (Mestre's client relays it). */
export function useSpotlightFollow(isGm: boolean, sceneId: string | null) {
  const { subscribe, send } = useRealtime()
  const enabled = useSpotlight()
  useEffect(() => {
    if (!isGm || !enabled || !sceneId) return
    return subscribe('action.started', (p) => {
      const a = (p as { actor?: { x: number; z: number } }).actor
      if (a) send('camera.focusAll', { sceneId, x: a.x + 0.5, z: a.z + 0.5 }).catch(() => undefined)
    })
  }, [isGm, enabled, sceneId, subscribe, send])
}
