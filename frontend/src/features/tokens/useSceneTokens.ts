import { useCallback, useEffect, useMemo, useRef, useState, type Dispatch, type SetStateAction } from 'react'
import type { MapGrid } from '../map3d/grid'
import { useRealtime } from '../realtime/realtimeContext'
import {
  applyCharacter, applyMoved, applyPlaced, applyRemoved, loadTokens, spawnCandidates,
  type CameraTarget, type CharacterEvent, type PingMarker, type Token,
} from './tokens'

const PING_MS = 3000
const message = (e: unknown, fallback: string) => (e instanceof Error ? e.message : fallback)

interface PointEvent {
  sceneId: string
  x: number
  z: number
  userName?: string
}

function usePings(sceneId: string) {
  const { subscribe } = useRealtime()
  const [pings, setPings] = useState<PingMarker[]>([])
  const [camera, setCamera] = useState<CameraTarget | null>(null)
  const seq = useRef(0)
  useEffect(() => {
    const offPing = subscribe('map.pinged', (p) => {
      const e = p as PointEvent
      if (e.sceneId !== sceneId) return
      const id = ++seq.current
      setPings((l) => [...l, { id, x: e.x, z: e.z, userName: e.userName ?? '' }])
      setCamera({ x: e.x, z: e.z, n: id })
      setTimeout(() => setPings((l) => l.filter((m) => m.id !== id)), PING_MS)
    })
    const offFocus = subscribe('camera.focused', (p) => {
      const e = p as PointEvent
      if (e.sceneId === sceneId) setCamera({ x: e.x, z: e.z, n: ++seq.current })
    })
    return () => {
      offPing()
      offFocus()
    }
  }, [subscribe, sceneId])
  // Local-only camera move (the camera controls), unlike camera.focusAll which reaches everyone.
  const focus = useCallback((x: number, z: number) => setCamera({ x, z, n: ++seq.current }), [])
  return { pings, camera, focus }
}

type Subscribe = ReturnType<typeof useRealtime>['subscribe']

function useTokenEvents(subscribe: Subscribe, sceneId: string, reload: () => void, known: { current: Token[] }, setTokens: Dispatch<SetStateAction<Token[]>>) {
  useEffect(() => {
    const offs = [
      // Sight changed: the server's filtered list says which tokens appear or vanish.
      subscribe('vision.updated', reload),
      subscribe('token.placed', (p) => setTokens((l) => applyPlaced(l, sceneId, (p as { token: Token }).token))),
      subscribe('token.removed', (p) => setTokens((l) => applyRemoved(l, (p as { tokenId: string }).tokenId))),
      subscribe('character.updated', (p) => setTokens((l) => applyCharacter(l, p as CharacterEvent))),
      subscribe('token.moved', (p) => {
        const e = p as { tokenId: string; x: number; z: number }
        // A token we do not know yet just walked into sight.
        if (!known.current.some((t) => t.id === e.tokenId)) reload()
        setTokens((l) => applyMoved(l, e.tokenId, e.x, e.z))
      }),
    ]
    return () => offs.forEach((off) => off())
  }, [subscribe, sceneId, reload, known, setTokens])
}

function useRemoteTokens(campaignId: string, sceneId: string) {
  const { subscribe, status } = useRealtime()
  const [tokens, setTokens] = useState<Token[]>([])
  const wasOpen = useRef(false)
  const known = useRef<Token[]>([])
  useEffect(() => { known.current = tokens }, [tokens])
  const reload = useCallback(() => {
    loadTokens(campaignId, sceneId).then(setTokens, () => undefined)
  }, [campaignId, sceneId])
  useEffect(reload, [reload])
  // Events may have been missed while offline.
  useEffect(() => {
    if (status === 'open' && wasOpen.current) reload()
    if (status === 'open') wasOpen.current = true
  }, [status, reload])
  useTokenEvents(subscribe, sceneId, reload, known, setTokens)
  return { tokens, setTokens }
}

type Fail = (fallback: string) => (e: unknown) => void
type Send = ReturnType<typeof useRealtime>['send']

function useMove(send: Send, tokensRef: { current: Token[] }, setTokens: Dispatch<SetStateAction<Token[]>>, report: Fail) {
  return useCallback((id: string, x: number, z: number, force = false) => {
    const prev = tokensRef.current.find((t) => t.id === id)
    if (!prev || (prev.x === x && prev.z === z)) return
    setTokens((l) => applyMoved(l, id, x, z))
    send('token.move', { tokenId: id, x, z, force }).catch((e) => {
      setTokens((l) => applyMoved(l, id, prev.x, prev.z))
      report('Não foi possível mover o token.')(e)
    })
  }, [send, setTokens, tokensRef, report])
}

// Tries successive free cells while the server says the cell itself is unusable.
function usePlace(send: Send, sceneId: string, grid: MapGrid, tokensRef: { current: Token[] }, report: Fail) {
  return useCallback(async (characterId: string, at?: { x: number; z: number }) => {
    let last: unknown
    // A chosen cell is tried alone: the server's refusal tells the Mestre why it cannot go there.
    for (const c of at ? [at] : spawnCandidates(grid, tokensRef.current)) {
      try {
        await send('token.place', { sceneId, characterId, x: c.x, z: c.z })
        return
      } catch (e) {
        last = e
        if (!/bloque|terreno|altura|objeto/i.test(message(e, ''))) break
      }
    }
    report('Não foi possível colocar o token.')(last)
  }, [send, sceneId, grid, tokensRef, report])
}

function useCommands(sceneId: string, grid: MapGrid, tokensRef: { current: Token[] }, setTokens: Dispatch<SetStateAction<Token[]>>, fail: (m: string) => void) {
  const { send } = useRealtime()
  const view = useRef<(() => { x: number; z: number }) | null>(null)
  const report: Fail = useCallback((fallback) => (e) => fail(message(e, fallback)), [fail])
  const move = useMove(send, tokensRef, setTokens, report)
  const place = usePlace(send, sceneId, grid, tokensRef, report)

  const remove = useCallback((tokenId: string) => {
    send('token.remove', { tokenId }).catch(report('Não foi possível remover o token.'))
  }, [send, report])

  const ping = useCallback((x: number, z: number) => {
    send('map.ping', { sceneId, x, z }).catch(report('Não foi possível pingar.'))
  }, [send, sceneId, report])

  const focusAll = useCallback(() => {
    const at = view.current?.() ?? { x: grid.width / 2, z: grid.depth / 2 }
    send('camera.focusAll', { sceneId, ...at }).catch(report('Não foi possível focar a câmera.'))
  }, [send, sceneId, grid, report])

  return useMemo(() => ({ view, move, place, remove, ping, focusAll }), [move, place, remove, ping, focusAll])
}

/** Tokens of a scene kept live, with optimistic moves that roll back when the server refuses. */
export function useSceneTokens(campaignId: string, sceneId: string, grid: MapGrid) {
  const { tokens, setTokens } = useRemoteTokens(campaignId, sceneId)
  const { pings, camera, focus } = usePings(sceneId)
  const [error, setError] = useState('')
  const tokensRef = useRef(tokens)
  useEffect(() => { tokensRef.current = tokens }, [tokens])
  const cmd = useCommands(sceneId, grid, tokensRef, setTokens, setError)
  return useMemo(() => ({ tokens, pings, camera, error, focus, ...cmd, clearError: () => setError('') }),
    [tokens, pings, camera, error, focus, cmd])
}

export type SceneTokens = ReturnType<typeof useSceneTokens>
