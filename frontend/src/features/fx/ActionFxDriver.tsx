import { useEffect, useRef } from 'react'
import { fxBus } from '../../lib/fxBus'
import type { ActionTarget } from '../actions/actions'
import { useRealtime } from '../realtime/realtimeContext'
import { useEmoteListener } from './useEmoteListener'
import { useSpotlightFollow } from './spotlight'

interface Started {
  actionRunId: string
  rollId: string
  animation: string
  actorTokenId: string
  actor: { x: number; z: number }
  target: ActionTarget
  variant?: string
}

interface Resolved {
  actionRunId: string
  target: ActionTarget
  outcome: { kind: string; text: string }
}

/** Milliseconds from the start of the strike to its impact. */
const IMPACT_MS: Record<string, number> = { melee_strike: 450, ranged_shot: 1100 }
/** Safety net: play the action even if the dice tray never reports closing. */
const TRAY_WAIT_MS = 8000

/** The roll tray covers the map, so an action plays right after the dice tray closes (the roll result decides how). */
function whenTrayClosed(rollId: string, trayRoll: { current: string | null }, play: () => void) {
  if (trayRoll.current !== rollId) return play()
  let done = false
  const go = () => { if (!done) { done = true; off(); clearTimeout(timer); play() } }
  const off = fxBus.on('ui:tray', (p) => { if (p.rollId !== rollId) go() })
  const timer = setTimeout(go, TRAY_WAIT_MS)
}

function playAction(run: Started, res: Resolved) {
  const t = res.target
  fxBus.emit(`fx:${run.animation}`, { x: run.actor.x, z: run.actor.z, target: run.target, tokenId: run.actorTokenId, variant: run.variant })
  if (res.outcome.kind === 'damage_suggested' && t.x !== undefined && t.z !== undefined) {
    const { x, z } = t
    setTimeout(() => fxBus.emit('fx:hit_react', { x, z, target: t, tokenId: t.tokenId, from: run.actor, variant: run.variant }), IMPACT_MS[run.animation] ?? 500)
  }
}

/** Turns the server's action events into `fx:*` bus events for everyone (the visuals and sounds subscribe to those). */
export function ActionFxDriver({ isGm = false, sceneId = null }: { isGm?: boolean; sceneId?: string | null }) {
  const { subscribe } = useRealtime()
  const runs = useRef(new Map<string, Started>())
  const trayRoll = useRef<string | null>(null)
  useSpotlightFollow(isGm, sceneId)
  useEmoteListener()
  useEffect(() => {
    const offs = [
      fxBus.on('ui:tray', (p) => { trayRoll.current = (p.rollId as string | null) ?? null }),
      subscribe('action.started', (p) => { const e = p as Started; runs.current.set(e.actionRunId, e) }),
      subscribe('action.resolved', (p) => {
        const res = p as Resolved
        const run = runs.current.get(res.actionRunId)
        runs.current.delete(res.actionRunId)
        if (run) whenTrayClosed(run.rollId, trayRoll, () => playAction(run, res))
      }),
      subscribe('loot.found', (p) => {
        const loot = (p as { loot?: { x: number; z: number } }).loot
        if (loot) whenTrayClosed(trayRoll.current ?? '', trayRoll, () => fxBus.emit('fx:loot_found', { x: loot.x, z: loot.z }))
      }),
    ]
    return () => offs.forEach((off) => off())
  }, [subscribe])
  return null
}
