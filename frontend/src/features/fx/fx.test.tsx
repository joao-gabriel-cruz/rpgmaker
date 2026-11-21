import { act, render } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { fxBus } from '../../lib/fxBus'
import { EMOTE_MS, poseAt, REST_POSE } from '../miniature/emotes'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { ActionFxDriver } from './ActionFxDriver'
import { clearEmotes, emoteFrame, headingTo, playEmote } from './emoteState'
import { spotlight } from './spotlight'

const handlers: Record<string, ((p: unknown) => void)[]> = {}
const send = vi.fn().mockResolvedValue(undefined)

function mount(props: { isGm?: boolean; sceneId?: string | null } = {}) {
  const api = { status: 'open', send, subscribe: (t: string, h: (p: unknown) => void) => { (handlers[t] ??= []).push(h); return () => { handlers[t] = handlers[t].filter((x) => x !== h) } } } as unknown as CampaignSocketApi
  render(<RealtimeContext.Provider value={api}><ActionFxDriver {...props} /></RealtimeContext.Provider>)
}
const emit = (t: string, p: unknown) => act(() => handlers[t].forEach((h) => h(p)))

const started = {
  actionRunId: 'run1', rollId: 'roll1', actionId: 'melee_attack', animation: 'melee_strike', actorTokenId: 'tk-a', actorName: 'Rita', variant: 'slash',
  actor: { x: 3, z: 3 }, target: { kind: 'token', tokenId: 'tk-b', x: 4, z: 3 },
}

beforeEach(() => { clearEmotes(); send.mockClear(); spotlight.set(false) })
afterEach(() => { vi.unstubAllGlobals() })

it('plays fx:<animation> once the action resolves and the tray is closed, with the emote on the actor facing the target', () => {
  const seen = vi.fn()
  const off = fxBus.on('fx:melee_strike', seen)
  mount()
  emit('action.started', started)
  expect(seen).not.toHaveBeenCalled()
  emit('action.resolved', { actionRunId: 'run1', target: started.target, outcome: { kind: 'none', text: 'falhou' } })
  expect(seen).toHaveBeenCalledWith({ x: 3, z: 3, target: started.target, tokenId: 'tk-a', variant: 'slash' })
  expect(emoteFrame('tk-a').active).toBe(true)
  expect(emoteFrame('tk-b').active).toBe(false)
  expect(headingTo({ x: 3, z: 3 }, { x: 4, z: 3 })).toBeCloseTo(Math.PI / 2)
  off()
})

it('waits for the dice tray of that roll to close before playing (the tray covers the map)', () => {
  const seen = vi.fn()
  const off = fxBus.on('fx:melee_strike', seen)
  mount()
  fxBus.emit('ui:tray', { x: 0, z: 0, rollId: 'roll1' })
  emit('action.started', started)
  emit('action.resolved', { actionRunId: 'run1', target: started.target, outcome: { kind: 'none', text: 'falhou' } })
  expect(seen).not.toHaveBeenCalled()
  fxBus.emit('ui:tray', { x: 0, z: 0, rollId: null })
  expect(seen).toHaveBeenCalledTimes(1)
  off()
})

it('plays hit_react on the target (facing the attacker) only when the hit lands', () => {
  vi.useFakeTimers()
  const hit = vi.fn()
  const off = fxBus.on('fx:hit_react', hit)
  mount()
  emit('action.started', started)
  emit('action.resolved', { actionRunId: 'run1', target: started.target, outcome: { kind: 'none', text: 'falhou' } })
  vi.advanceTimersByTime(2000)
  expect(hit).not.toHaveBeenCalled()
  emit('action.started', { ...started, actionRunId: 'run2' })
  emit('action.resolved', { actionRunId: 'run2', target: started.target, outcome: { kind: 'damage_suggested', text: 'acertou' } })
  vi.advanceTimersByTime(600)
  expect(hit).toHaveBeenCalledWith(expect.objectContaining({ x: 4, z: 3, tokenId: 'tk-b', from: { x: 3, z: 3 } }))
  expect(emoteFrame('tk-b').active).toBe(true)
  off()
  vi.useRealTimers()
})

it('announces found loot at its cell', () => {
  const seen = vi.fn()
  const off = fxBus.on('fx:loot_found', seen)
  mount()
  emit('loot.found', { loot: { x: 7, z: 2 } })
  expect(seen).toHaveBeenCalledWith({ x: 7, z: 2 })
  off()
})

it('keeps emotes off under prefers-reduced-motion but still emits the bus event for sounds and static icons', () => {
  vi.stubGlobal('matchMedia', (q: string) => ({ matches: q.includes('reduce') }))
  const seen = vi.fn()
  const off = fxBus.on('fx:dance', seen)
  mount()
  emit('action.started', { ...started, animation: 'dance', target: { kind: 'none' } })
  emit('action.resolved', { actionRunId: 'run1', target: { kind: 'none' }, outcome: { kind: 'morale', text: 'x' } })
  expect(seen).toHaveBeenCalled()
  expect(emoteFrame('tk-a').active).toBe(false)
  off()
})

it('spotlight on: the Mestre relays each action.started as camera.focusAll on the active scene', () => {
  spotlight.set(true)
  mount({ isGm: true, sceneId: 'sc1' })
  emit('action.started', started)
  expect(send).toHaveBeenCalledWith('camera.focusAll', { sceneId: 'sc1', x: 3.5, z: 3.5 })
})

it('spotlight off, or a player client: no camera command', () => {
  mount({ isGm: true, sceneId: 'sc1' })
  emit('action.started', started)
  expect(send).not.toHaveBeenCalled()
  spotlight.set(true)
  document.body.innerHTML = ''
})

it('emotes end by themselves after their duration and a new emote replaces the running one', () => {
  playEmote('t', 'dance', null, 1000)
  expect(emoteFrame('t', 1500).active).toBe(true)
  playEmote('t', 'rest', null, 1600)
  expect(emoteFrame('t', 1700).pose.squash).toBeLessThan(1)
  expect(emoteFrame('t', 1600 + EMOTE_MS.rest + 1).active).toBe(false)
  playEmote('t', 'not-an-emote', null, 0)
  expect(emoteFrame('t', 5).active).toBe(false)
})

it('procedural poses start and end at rest and show what each emote promises', () => {
  for (const name of Object.keys(EMOTE_MS) as (keyof typeof EMOTE_MS)[]) {
    expect(poseAt(name, 0).dy).toBeCloseTo(REST_POSE.dy)
    expect(poseAt(name, 1).squash).toBeCloseTo(1)
    expect(poseAt(name, 1).opacity).toBeCloseTo(1)
  }
  expect(poseAt('melee_strike', 0.45).dz).toBeGreaterThan(0.2)
  expect(poseAt('sneak', 0.5).opacity).toBeLessThan(0.6)
  expect(poseAt('rest', 0.5).squash).toBeLessThan(0.7)
  expect(poseAt('dance', 0.1).armL[1]).toBeGreaterThan(0)
  expect(poseAt('hit_react', 0.1).dz).toBeLessThan(0)
  expect(Math.max(...[0.05, 0.2, 0.4].map((t) => poseAt('hit_react', t).flash))).toBeGreaterThan(0)
})
