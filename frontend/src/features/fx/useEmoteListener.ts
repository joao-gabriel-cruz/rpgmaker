import { useEffect } from 'react'
import { fxBus, type FxPayload } from '../../lib/fxBus'
import { ANIMATIONS } from '../actions/actions'
import { headingTo, playEmote } from './emoteState'
import { prefersReducedMotion } from './reducedMotion'

const targetPoint = (p: FxPayload) => (p.target && p.target.x !== undefined && p.target.z !== undefined ? { x: p.target.x, z: p.target.z } : null)

/** Starts the miniature emotes on `fx:<animation>` (towards the target) and `fx:hit_react` (facing the attacker); static under reduced motion. */
export function useEmoteListener() {
  useEffect(() => {
    const offs = ANIMATIONS.map((a) => fxBus.on(`fx:${a}`, (p) => {
      if (prefersReducedMotion() || typeof p.tokenId !== 'string') return
      const target = targetPoint(p)
      playEmote(p.tokenId, a, target ? headingTo(p, target) : null)
    }))
    offs.push(fxBus.on('fx:hit_react', (p) => {
      if (prefersReducedMotion() || typeof p.tokenId !== 'string') return
      const from = p.from as { x: number; z: number } | undefined
      playEmote(p.tokenId, 'hit_react', from ? headingTo(p, from) : null)
    }))
    return () => offs.forEach((off) => off())
  }, [])
}
