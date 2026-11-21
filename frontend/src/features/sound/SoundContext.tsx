import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { fxBus } from '../../lib/fxBus'
import { useRealtime } from '../realtime/realtimeContext'
import { createSoundEngine, type SoundEngine, type SoundSettings } from './engine'
import { FX_SOUNDS } from './fxSounds'
import { loadSoundSettings, saveSoundSettings } from './settings'

export interface SoundApi {
  engine: SoundEngine
  settings: SoundSettings
  setVolume: (v: number) => void
  setMuted: (m: boolean) => void
}

const Ctx = createContext<SoundApi | null>(null)

/** The table's sound engine and the player's settings; null outside the provider. */
export const useSound = () => useContext(Ctx)

const STEP_GAP_MS = 220
const THUNDER_DELAY_MS = [350, 1400]

// Realtime and fx events become sounds; the engine ignores them until the first user gesture.
function useSoundEvents(engine: SoundEngine) {
  const { subscribe } = useRealtime()
  useEffect(() => {
    let lastStep = 0
    const offs = [
      subscribe('roll.rolled', () => engine.play('dice')),
      subscribe('ambience.updated', () => engine.play('whoosh')),
      subscribe('token.moved', () => {
        const now = Date.now()
        if (now - lastStep < STEP_GAP_MS) return
        lastStep = now
        engine.play('step')
      }),
      ...Object.entries(FX_SOUNDS).map(([name, sound]) => fxBus.on(name, () => engine.play(sound))),
    ]
    const timers: ReturnType<typeof setTimeout>[] = []
    offs.push(fxBus.on('ambience:lightning', () => {
      timers.push(setTimeout(() => engine.play('thunder'), THUNDER_DELAY_MS[0] + Math.random() * (THUNDER_DELAY_MS[1] - THUNDER_DELAY_MS[0])))
    }))
    return () => {
      offs.forEach((off) => off())
      timers.forEach(clearTimeout)
    }
  }, [subscribe, engine])
}

// Browsers only allow audio after a gesture: the first click or key press unlocks the engine.
function useUnlockOnGesture(engine: SoundEngine) {
  useEffect(() => {
    const unlock = () => {
      engine.unlock()
      off()
    }
    const off = () => {
      window.removeEventListener('pointerdown', unlock)
      window.removeEventListener('keydown', unlock)
    }
    window.addEventListener('pointerdown', unlock)
    window.addEventListener('keydown', unlock)
    return off
  }, [engine])
}

export function SoundProvider({ children, createContext: factory }: { children: ReactNode; createContext?: () => AudioContext }) {
  const [settings, setSettings] = useState<SoundSettings>(loadSoundSettings)
  const first = useRef(settings)
  const engine = useMemo(() => createSoundEngine(factory, first.current), [factory])
  useEffect(() => () => engine.dispose(), [engine])
  useEffect(() => {
    engine.setSettings(settings)
    saveSoundSettings(settings)
  }, [engine, settings])
  useSoundEvents(engine)
  useUnlockOnGesture(engine)
  const setVolume = useCallback((volume: number) => setSettings((s) => ({ ...s, volume, muted: false })), [])
  const setMuted = useCallback((muted: boolean) => setSettings((s) => ({ ...s, muted })), [])
  const value = useMemo(() => ({ engine, settings, setVolume, setMuted }), [engine, settings, setVolume, setMuted])
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}
