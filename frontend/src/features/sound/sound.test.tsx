import { act, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { fxBus } from '../../lib/fxBus'
import { fakeRealtime } from '../../test/fakeRealtime'
import { createSoundEngine, levelOf } from './engine'
import { loadSoundSettings } from './settings'
import { SoundControl } from './SoundControl'
import { SoundProvider } from './SoundContext'
import { useSceneSounds } from './useSceneSounds'

// Records the nodes the engine creates; every param is a spy-free stub.
function fakeAudio() {
  const param = () => ({ value: 0, setValueAtTime: vi.fn(), exponentialRampToValueAtTime: vi.fn(), linearRampToValueAtTime: vi.fn(), setTargetAtTime: vi.fn() })
  const node = (extra: object = {}) => {
    const n: Record<string, unknown> = { connect: (d: unknown) => d ?? n, start: vi.fn(), stop: vi.fn(), ...extra }
    return n
  }
  const made = { oscillators: 0, sources: 0, gains: [] as ReturnType<typeof param>[] }
  const ctx = {
    currentTime: 0, sampleRate: 8000, state: 'suspended', destination: node(),
    resume: vi.fn(() => { ctx.state = 'running'; return Promise.resolve() }),
    close: vi.fn(() => Promise.resolve()),
    createBuffer: (_c: number, len: number) => ({ sampleRate: 8000, getChannelData: () => new Float32Array(len) }),
    createGain: () => { const gain = param(); made.gains.push(gain); return node({ gain }) },
    createOscillator: () => { made.oscillators++; return node({ frequency: param() }) },
    createBufferSource: () => { made.sources++; return node({}) },
    createBiquadFilter: () => node({ frequency: param(), Q: param() }),
  }
  return { ctx: ctx as unknown as AudioContext, raw: ctx, made }
}

const total = (m: ReturnType<typeof fakeAudio>['made']) => m.oscillators + m.sources

beforeEach(() => localStorage.clear())
afterEach(() => vi.useRealTimers())

it('engine stays silent until unlocked by a gesture and while muted', () => {
  const a = fakeAudio()
  const engine = createSoundEngine(() => a.ctx)
  engine.play('dice')
  expect(total(a.made)).toBe(0)
  engine.unlock()
  expect(a.raw.resume).toHaveBeenCalled()
  const base = total(a.made) // the always-on beds
  engine.play('dice')
  expect(total(a.made)).toBeGreaterThan(base)
  const after = total(a.made)
  engine.setSettings({ volume: 0.5, muted: true })
  engine.play('hit')
  expect(total(a.made)).toBe(after)
  engine.setSettings({ volume: 0.5, muted: false })
  engine.play('hit')
  expect(total(a.made)).toBeGreaterThan(after)
})

it('engine survives a missing AudioContext', () => {
  const engine = createSoundEngine(() => { throw new Error('no audio') })
  expect(() => { engine.unlock(); engine.play('hit'); engine.setBeds({ rain: 'rain', fire: true, wind: true }) }).not.toThrow()
})

it('volume curve: muted is zero, and the slider is squared', () => {
  expect(levelOf({ volume: 0.8, muted: true })).toBe(0)
  expect(levelOf({ volume: 0.5, muted: false })).toBe(0.25)
  expect(levelOf({ volume: 4, muted: false })).toBe(1)
})

function setup() {
  const a = fakeAudio()
  const rt = fakeRealtime()
  render(<rt.Provider><SoundProvider createContext={() => a.ctx}><SoundControl /></SoundProvider></rt.Provider>)
  return { a, rt }
}

it('responds to realtime events and fx events once the first click unlocked audio', async () => {
  const { a, rt } = setup()
  act(() => rt.emit('roll.rolled', {}))
  expect(total(a.made)).toBe(0) // no gesture yet
  await userEvent.click(document.body)
  const base = total(a.made)
  act(() => rt.emit('roll.rolled', {}))
  const afterRoll = total(a.made)
  expect(afterRoll).toBeGreaterThan(base)
  act(() => rt.emit('token.moved', { tokenId: 't', x: 1, z: 1 }))
  const afterStep = total(a.made)
  expect(afterStep).toBeGreaterThan(afterRoll)
  for (const fx of ['fx:melee_strike', 'fx:ranged_shot', 'fx:hit_react', 'fx:dance', 'fx:rest']) {
    const before = total(a.made)
    act(() => fxBus.emit(fx, { x: 1, z: 1 }))
    expect(total(a.made), fx).toBeGreaterThan(before)
  }
  const beforeWeather = total(a.made)
  act(() => rt.emit('ambience.updated', { sceneId: 's', timeOfDay: 'night', weather: 'rain' }))
  expect(total(a.made)).toBeGreaterThan(beforeWeather)
})

it('plays thunder shortly after a lightning flash', async () => {
  const { a } = setup()
  await userEvent.click(document.body)
  vi.useFakeTimers()
  const before = total(a.made)
  act(() => fxBus.emit('ambience:lightning', {}))
  expect(total(a.made)).toBe(before)
  act(() => { vi.advanceTimersByTime(2000) })
  expect(total(a.made)).toBeGreaterThan(before)
})

it('volume and mute are per player: set from the top bar button and persisted in localStorage', async () => {
  setup()
  await userEvent.click(screen.getByRole('button', { name: 'Som' }))
  const slider = screen.getByLabelText(/Volume/)
  expect(slider).toHaveValue('60')
  fireEvent.change(slider, { target: { value: '40' } })
  expect(loadSoundSettings().volume).toBeCloseTo(0.4)
  await userEvent.click(screen.getByRole('button', { name: /Desligar o som/ }))
  expect(loadSoundSettings()).toEqual({ volume: 0.4, muted: true })
  expect(screen.getByRole('button', { name: 'Som' }).querySelector('svg')).toHaveClass('lucide-volume-x')
})

it('loads saved settings and tolerates broken or blocked storage', () => {
  localStorage.setItem('rpg.sound.v1', JSON.stringify({ volume: 0.2, muted: true }))
  expect(loadSoundSettings()).toEqual({ volume: 0.2, muted: true })
  localStorage.setItem('rpg.sound.v1', '{oops')
  expect(loadSoundSettings()).toEqual({ volume: 0.6, muted: false })
  const spy = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw new Error('blocked') })
  expect(loadSoundSettings()).toEqual({ volume: 0.6, muted: false })
  spy.mockRestore()
})

function Beds({ weather, fire }: { weather: 'rain' | 'clear'; fire: boolean }) {
  useSceneSounds({ timeOfDay: 'night', weather }, fire)
  return null
}

it('scene beds follow weather and campfire at night', async () => {
  const a = fakeAudio()
  const rt = fakeRealtime()
  const factory = () => a.ctx
  const { rerender } = render(<rt.Provider><SoundProvider createContext={factory}><Beds weather="clear" fire={false} /></SoundProvider></rt.Provider>)
  await userEvent.click(document.body)
  const rainGain = () => a.made.gains.at(1)?.setTargetAtTime.mock.calls.at(-1)?.[0]
  expect(rainGain()).toBe(0)
  rerender(<rt.Provider><SoundProvider createContext={factory}><Beds weather="rain" fire /></SoundProvider></rt.Provider>)
  expect(rainGain()).toBeGreaterThan(0)
})
