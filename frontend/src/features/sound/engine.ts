import { createBeds, NO_BEDS, type BedState } from './beds'
import { RECIPES, type SoundName } from './synth'

export interface SoundSettings {
  /** 0..1 */
  volume: number
  muted: boolean
}

export interface SoundEngine {
  /** Creates/resumes the AudioContext; call from a user gesture (browsers block audio before one). */
  unlock: () => void
  setSettings: (s: SoundSettings) => void
  play: (name: SoundName) => void
  setBeds: (b: BedState) => void
  dispose: () => void
}

type ContextFactory = () => AudioContext

const defaultFactory: ContextFactory = () => {
  const Ctor = typeof AudioContext === 'function' ? AudioContext : (globalThis as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
  if (!Ctor) throw new Error('WebAudio is not available')
  return new Ctor()
}

/** Output level for a slider value: squared so the low end of the slider is usable. */
export const levelOf = (s: SoundSettings) => (s.muted ? 0 : Math.max(0, Math.min(1, s.volume)) ** 2)

interface Graph {
  ctx: AudioContext
  master: GainNode
  beds: ReturnType<typeof createBeds>
}

function openGraph(factory: ContextFactory): Graph | null {
  try {
    const ctx = factory()
    const master = ctx.createGain()
    master.connect(ctx.destination)
    return { ctx, master, beds: createBeds(ctx, master) }
  } catch {
    return null
  }
}

class WebAudioEngine implements SoundEngine {
  private beds: BedState = NO_BEDS
  private graph: Graph | null = null
  private broken = false

  private factory: ContextFactory
  private settings: SoundSettings

  constructor(factory: ContextFactory, settings: SoundSettings) {
    this.factory = factory
    this.settings = settings
  }

  private apply() {
    this.graph?.master.gain.setTargetAtTime(levelOf(this.settings), this.graph.ctx.currentTime, 0.03)
  }

  unlock = () => {
    if (!this.graph && !this.broken) {
      this.graph = openGraph(this.factory)
      this.broken = !this.graph
      this.apply()
      this.graph?.beds.set(this.beds)
    }
    if (this.graph?.ctx.state === 'suspended') void this.graph.ctx.resume()
  }

  setSettings = (s: SoundSettings) => {
    this.settings = s
    this.apply()
  }

  play = (name: SoundName) => {
    const g = this.graph
    if (g && levelOf(this.settings) > 0) RECIPES[name](g.ctx, g.master, g.ctx.currentTime)
  }

  setBeds = (b: BedState) => {
    this.beds = b
    this.graph?.beds.set(b)
  }

  dispose = () => {
    this.graph?.beds.stop()
    void this.graph?.ctx.close?.()
    this.graph = null
  }
}

/**
 * WebAudio synth engine. Sounds are ignored until `unlock()` ran (first user gesture) and while muted;
 * the continuous beds keep running at zero gain when muted so unmuting resumes them at once.
 */
export function createSoundEngine(factory: ContextFactory = defaultFactory, initial: SoundSettings = { volume: 0.6, muted: false }): SoundEngine {
  return new WebAudioEngine(factory, initial)
}
