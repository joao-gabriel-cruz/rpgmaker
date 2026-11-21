import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import {
  applyAdded, applyRemoved, DRAW_COLORS, DRAW_WIDTHS, loadDrawings, simplify,
  type DrawMode, type Drawing, type Point,
} from './drawings'

const RULER_TTL_MS = 8000

export interface RemoteRuler {
  userId: string
  userName: string
  from: Point
  to: Point
}

interface RulerEvent {
  sceneId: string
  userId: string
  userName: string
  from: Point | null
  to: Point | null
}

function useRemoteDrawings(campaignId: string, sceneId: string) {
  const { subscribe, status } = useRealtime()
  const [drawings, setDrawings] = useState<Drawing[]>([])
  const wasOpen = useRef(false)
  const reload = useCallback(() => {
    loadDrawings(campaignId, sceneId).then(setDrawings, () => undefined)
  }, [campaignId, sceneId])
  useEffect(reload, [reload])
  // Events may have been missed while offline.
  useEffect(() => {
    if (status === 'open' && wasOpen.current) reload()
    if (status === 'open') wasOpen.current = true
  }, [status, reload])
  useEffect(() => {
    const offs = [
      subscribe('drawing.added', (p) => setDrawings((l) => applyAdded(l, sceneId, (p as { drawing: Drawing }).drawing))),
      subscribe('drawing.removed', (p) => setDrawings((l) => applyRemoved(l, (p as { drawingId: string }).drawingId))),
      subscribe('drawings.cleared', (p) => {
        if ((p as { sceneId: string }).sceneId === sceneId) setDrawings([])
      }),
    ]
    return () => offs.forEach((off) => off())
  }, [subscribe, sceneId])
  return drawings
}

function useRemoteRulers(sceneId: string) {
  const { subscribe } = useRealtime()
  const [rulers, setRulers] = useState<Record<string, RemoteRuler>>({})
  useEffect(() => {
    const timers = new Map<string, ReturnType<typeof setTimeout>>()
    const drop = (id: string) => setRulers(({ [id]: _gone, ...rest }) => rest)
    const off = subscribe('ruler.moved', (p) => {
      const e = p as RulerEvent
      if (e.sceneId !== sceneId) return
      clearTimeout(timers.get(e.userId))
      if (!e.from || !e.to) return drop(e.userId)
      const ruler = { userId: e.userId, userName: e.userName, from: e.from, to: e.to }
      setRulers((r) => ({ ...r, [e.userId]: ruler }))
      timers.set(e.userId, setTimeout(() => drop(e.userId), RULER_TTL_MS))
    })
    return () => {
      off()
      timers.forEach(clearTimeout)
    }
  }, [subscribe, sceneId])
  return Object.values(rulers)
}

function useCommands(sceneId: string, color: string, width: number, setError: (m: string) => void) {
  const { send } = useRealtime()
  const fail = useCallback((fallback: string) => (e: unknown) => setError(e instanceof Error ? e.message : fallback), [setError])
  const add = useCallback((points: Point[]) => {
    const pts = simplify(points)
    if (pts.length < 2 && pts[0]) pts.push({ x: pts[0].x + 0.01, z: pts[0].z })
    if (pts.length < 2) return
    setError('')
    send('drawing.add', { sceneId, color, width, points: pts }).catch(fail('Não foi possível desenhar.'))
  }, [send, sceneId, color, width, fail, setError])
  const remove = useCallback((id: string) => {
    send('drawing.remove', { drawingId: id }).catch(fail('Não foi possível apagar o desenho.'))
  }, [send, fail])
  const clear = useCallback(() => {
    send('drawing.clear', { sceneId }).catch(fail('Não foi possível limpar os desenhos.'))
  }, [send, sceneId, fail])
  const shareRuler = useCallback((from: Point | null, to: Point | null) => {
    send('ruler.move', { sceneId, from, to }).catch(() => undefined)
  }, [send, sceneId])
  return { add, remove, clear, shareRuler }
}

/** Drawings, ruler sharing and tool settings of a scene, kept live over the socket. */
export function useDrawings(campaignId: string, sceneId: string, userId: string) {
  const drawings = useRemoteDrawings(campaignId, sceneId)
  const rulers = useRemoteRulers(sceneId)
  const [mode, setMode] = useState<DrawMode>('none')
  const [color, setColor] = useState(DRAW_COLORS[0])
  const [width, setWidth] = useState<number>(DRAW_WIDTHS[1])
  const [error, setError] = useState('')
  const { add, remove, clear, shareRuler } = useCommands(sceneId, color, width, setError)
  const removeMine = useCallback(() => drawings.filter((d) => d.userId === userId).forEach((d) => remove(d.id)), [drawings, userId, remove])

  return useMemo(() => ({
    drawings, rulers, mode, setMode, color, setColor, width, setWidth, error, add, remove, removeMine, clear, shareRuler,
    mineCount: drawings.filter((d) => d.userId === userId).length,
  }), [drawings, rulers, mode, color, width, error, add, remove, removeMine, clear, shareRuler, userId])
}

export type SceneDrawings = ReturnType<typeof useDrawings>
