import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'
import { DEFAULT_AMBIENCE, isAmbience, type Ambience, type TimeOfDay, type Weather } from './ambience'

export interface SceneAmbience {
  ambience: Ambience
  /** Mestre only: persists and announces the change to the whole table. */
  set: (change: Partial<{ timeOfDay: TimeOfDay; weather: Weather }>) => void
  error: string
}

// Initial load plus `ambience.updated` events for this scene.
function useAmbienceFeed(campaignId: string, sceneId: string) {
  const { subscribe } = useRealtime()
  const [ambience, setAmbience] = useState<Ambience>({ sceneId, ...DEFAULT_AMBIENCE })
  useEffect(() => {
    let alive = true
    api.get<Ambience>(`/campaigns/${campaignId}/scenes/${sceneId}/ambience`).then((a) => alive && isAmbience(a) && setAmbience(a), () => undefined)
    const off = subscribe('ambience.updated', (p) => {
      const a = p as Ambience
      if (isAmbience(a) && a.sceneId === sceneId) setAmbience(a)
    })
    return () => {
      alive = false
      off()
    }
  }, [campaignId, sceneId, subscribe])
  return [ambience, setAmbience] as const
}

/** Time of day and weather of a scene, kept live through `ambience.updated`. */
export function useSceneAmbience(campaignId: string, sceneId: string): SceneAmbience {
  const { send } = useRealtime()
  const [ambience, setAmbience] = useAmbienceFeed(campaignId, sceneId)
  const [error, setError] = useState('')
  const latest = useRef(ambience)
  useEffect(() => { latest.current = ambience }, [ambience])
  const set = useCallback<SceneAmbience['set']>((change) => {
    setError('')
    const prev = latest.current
    // Optimistic: two quick clicks (time, then weather) must not read a stale state.
    const next = { ...prev, ...change }
    latest.current = next
    setAmbience(next)
    send('ambience.set', { sceneId, timeOfDay: next.timeOfDay, weather: next.weather }).catch((e: Error) => {
      latest.current = prev
      setAmbience(prev)
      setError(e.message)
    })
  }, [send, sceneId, setAmbience])
  return useMemo(() => ({ ambience, set, error }), [ambience, set, error])
}
