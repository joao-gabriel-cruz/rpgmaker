import { useCallback, useEffect, useRef, useState } from 'react'
import { fxBus, type FxPayload } from '../../lib/fxBus'
import { ANIMATIONS, type Animation } from '../actions/actions'
import { getCell, surfaceHeight, type MapGrid } from '../map3d/grid'
import type { Token } from '../tokens/tokens'
import { START_EFFECTS, STATIC_GLYPH, type EffectSpec, type Fx } from './effects'
import { Confetti, Sparks } from './Particles'
import { prefersReducedMotion } from './reducedMotion'
import { Bubble, Rise, StaticFlash, Wave } from './Signs'
import { Arc, Projectile } from './Strikes'
import { useDamageNumbers } from './useDamageNumbers'

interface Props {
  campaignId: string
  grid: MapGrid
  tokens: readonly Token[]
}

type Point = { x: number; z: number }

const pointOf = (p: FxPayload): Point | null => (p.target && p.target.x !== undefined && p.target.z !== undefined ? { x: p.target.x, z: p.target.z } : null)

function build(spec: EffectSpec, p: FxPayload, id: number): Fx | null {
  const target = pointOf(p)
  if (spec.toTarget && !target) return null
  const at = spec.atTarget && target ? target : { x: p.x, z: p.z }
  const { atTarget: _a, toTarget: _t, ...rest } = spec
  return { ...rest, id, x: at.x, z: at.z, toX: target?.x, toZ: target?.z, variant: p.variant as string | undefined, seed: id * 7 + 3 }
}

function useSpawner() {
  const [fxs, setFxs] = useState<Fx[]>([])
  const seq = useRef(0)
  const spawn = useCallback((specs: EffectSpec[], p: FxPayload) => {
    const made = specs.map((s) => build(s, p, ++seq.current)).filter((f): f is Fx => !!f)
    if (made.length) setFxs((l) => [...l, ...made])
  }, [])
  const done = useCallback((id: number) => setFxs((l) => l.filter((f) => f.id !== id)), [])
  return { fxs, spawn, done }
}

const SPARKS: EffectSpec = { kind: 'sparks', ms: 600, delay: 0 }
const staticSpec = (glyph: string): EffectSpec => ({ kind: 'flash', glyph, ms: 900, delay: 0 })

// fx:<animation>, fx:hit_react, fx:loot_found and fx:number -> map effects (the emotes live in useEmoteListener).
function useFxEvents(spawn: (specs: EffectSpec[], p: FxPayload) => void) {
  useEffect(() => {
    const reduced = prefersReducedMotion
    const offs = ANIMATIONS.map((a: Animation) => fxBus.on(`fx:${a}`, (p) => spawn(reduced() ? [staticSpec(STATIC_GLYPH[a])] : START_EFFECTS[a], p)))
    offs.push(fxBus.on('fx:hit_react', (p) => spawn(reduced() ? [staticSpec('💥')] : [{ ...SPARKS, color: p.variant === 'blunt' ? '#fb923c' : '#fde047' }], p)))
    offs.push(fxBus.on('fx:loot_found', (p) => spawn(reduced()
      ? [staticSpec('✨')]
      : [{ kind: 'wave', ms: 1200, delay: 0, color: '#facc15' }, ...[0, 250, 500].map((delay): EffectSpec => ({ kind: 'rise', glyph: '✨', ms: 1500, delay }))], p)))
    offs.push(fxBus.on('fx:number', (p) => {
      const n = p.amount as number
      spawn([{ kind: 'number', glyph: n > 0 ? `-${n}` : `+${-n}`, color: n > 0 ? '#f87171' : '#4ade80', ms: 1400, delay: 0 }], p)
    }))
    return () => offs.forEach((off) => off())
  }, [spawn])
}

function Item({ fx, grid, onDone }: { fx: Fx; grid: MapGrid; onDone: (id: number) => void }) {
  const y = surfaceHeight(getCell(grid, Math.floor(fx.x), Math.floor(fx.z))?.height ?? 1)
  const p = { fx, y, onDone }
  switch (fx.kind) {
    case 'arc': return <Arc {...p} />
    case 'projectile': return <Projectile {...p} />
    case 'sparks': return <Sparks {...p} />
    case 'confetti': return <Confetti {...p} />
    case 'wave': return <Wave {...p} />
    case 'bubble': return <Bubble {...p} />
    case 'flash': return <StaticFlash {...p} />
    default: return <Rise {...p} />
  }
}

/** In-canvas effects driven by the fx bus: strike arcs, projectiles, sparks, waves, signs and floating numbers. */
export function FxLayer({ campaignId, grid, tokens }: Props) {
  const { fxs, spawn, done } = useSpawner()
  useFxEvents(spawn)
  useDamageNumbers(campaignId, tokens)
  return <>{fxs.map((fx) => <Item key={fx.id} fx={fx} grid={grid} onDone={done} />)}</>
}
