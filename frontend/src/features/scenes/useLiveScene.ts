import { useCallback, useEffect, useRef, useState, type Dispatch, type SetStateAction } from 'react'
import { applyPatch, type MapPatch } from '../map3d/grid'
import { applyObjectPatch, type ObjectPatch } from '../map3d/objects/model'
import { useRealtime } from '../realtime/realtimeContext'
import { applyRemoteObjects, applyRemotePaint } from './remote'
import { loadScene, type MapState, type ObjectsEvent, type PaintEvent } from './scenes'

function useRemoteSync(campaignId: string, sceneId: string, setState: Dispatch<SetStateAction<MapState | null>>) {
  const { subscribe, status } = useRealtime()
  const [error, setError] = useState('')
  const wasOpen = useRef(false)

  const reload = useCallback(() => {
    loadScene(campaignId, sceneId).then(setState, () => setError('Não foi possível carregar a cena.'))
  }, [campaignId, sceneId, setState])

  // The owner keys this hook by scene, so the initial null state is already correct.
  useEffect(reload, [reload])

  // Events may have been missed while offline, so resync after a reconnection.
  useEffect(() => {
    if (status === 'open' && wasOpen.current) reload()
    if (status === 'open') wasOpen.current = true
  }, [status, reload])

  useEffect(() => {
    const offPaint = subscribe('map.patched', (p) =>
      setState((s) => s && applyRemotePaint(s, sceneId, p as PaintEvent)))
    const offObjects = subscribe('map.objects.patched', (p) =>
      setState((s) => s && applyRemoteObjects(s, sceneId, p as ObjectsEvent)))
    return () => {
      offPaint()
      offObjects()
    }
  }, [subscribe, sceneId, setState])
  return { error, setError, reload }
}

/** Loads a scene (grid + objects in one GET) and keeps it live; the Mestre's edits are optimistic. */
export function useLiveScene(campaignId: string, sceneId: string) {
  const { send } = useRealtime()
  const [state, setState] = useState<MapState | null>(null)
  const { error, setError, reload } = useRemoteSync(campaignId, sceneId, setState)

  const fail = useCallback((e: unknown) => {
    setError(e instanceof Error ? e.message : 'Não foi possível salvar a alteração.')
    reload()
  }, [reload, setError])

  const paint = useCallback((patch: MapPatch) => {
    setState((s) => s && { ...s, grid: applyPatch(s.grid, patch) })
    send('map.paint', { sceneId, patch }).catch(fail)
  }, [send, sceneId, fail])

  const editObjects = useCallback((patch: ObjectPatch) => {
    setState((s) => s && { ...s, objects: applyObjectPatch(s.objects, patch) })
    send('map.objects', { sceneId, patch }).catch(fail)
  }, [send, sceneId, fail])

  return { state, error, paint, editObjects }
}
