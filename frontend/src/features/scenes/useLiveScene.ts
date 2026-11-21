import { useCallback, useEffect, useRef, useState, type Dispatch, type SetStateAction } from 'react'
import { applyPatch, type MapPatch } from '../map3d/grid'
import { applyObjectPatch, type ObjectPatch } from '../map3d/objects/model'
import { useRealtime } from '../realtime/realtimeContext'
import { applyRemoteObjects, applyRemotePaint } from './remote'
import { loadScene, replacedState, type MapState, type ObjectsEvent, type PaintEvent, type ReplacedEvent } from './scenes'

function useSceneEvents(sceneId: string, setState: Dispatch<SetStateAction<MapState | null>>, reload: () => void, onReplaced: () => void) {
  const { subscribe } = useRealtime()
  useEffect(() => {
    // Draft patches (other Mestre sessions only) share the shape of the real ones.
    const paint = (p: unknown) => setState((s) => s && applyRemotePaint(s, sceneId, p as PaintEvent))
    const objs = (p: unknown) => setState((s) => s && applyRemoteObjects(s, sceneId, p as ObjectsEvent))
    const offs = [subscribe('map.patched', paint), subscribe('map.draft.patched', paint),
      subscribe('map.objects.patched', objs), subscribe('map.draft.objects.patched', objs)]
    // Draft started/discarded elsewhere: refetch which layer we are looking at.
    const offDraft = subscribe('map.draft.changed', (p) => {
      if ((p as { sceneId?: string })?.sceneId === sceneId) reload()
    })
    const offReplaced = subscribe('map.replaced', (p) => {
      const ev = p as ReplacedEvent
      if (ev?.sceneId !== sceneId || !ev.grid) return
      setState(replacedState(ev))
      onReplaced()
    })
    return () => {
      offs.forEach((off) => off())
      offDraft()
      offReplaced()
    }
  }, [subscribe, sceneId, setState, reload, onReplaced])
}

function useRemoteSync(campaignId: string, sceneId: string, setState: Dispatch<SetStateAction<MapState | null>>) {
  const { status } = useRealtime()
  const [error, setError] = useState('')
  const [replacedAt, setReplacedAt] = useState(0)
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

  const bump = useCallback(() => setReplacedAt((n) => n + 1), [])
  useSceneEvents(sceneId, setState, reload, bump)
  return { error, setError, reload, replacedAt }
}

/** Loads a scene (grid + objects in one GET) and keeps it live; the Mestre's edits are optimistic. */
export function useLiveScene(campaignId: string, sceneId: string) {
  const { send } = useRealtime()
  const [state, setState] = useState<MapState | null>(null)
  const { error, setError, reload, replacedAt } = useRemoteSync(campaignId, sceneId, setState)

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

  // Draft and snapshot commands only change state through the events the server sends back.
  const command = useCallback((type: string, extra: object = {}) => {
    send(type, { sceneId, ...extra }).catch(fail)
  }, [send, sceneId, fail])

  return { state, error, paint, editObjects, command, replacedAt }
}
