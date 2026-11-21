import { burst, whiteNoise } from './synth'

export type Rain = 'none' | 'rain' | 'storm'

export interface BedState {
  rain: Rain
  fire: boolean
  wind: boolean
}

export const NO_BEDS: BedState = { rain: 'none', fire: false, wind: false }

const RAIN_GAIN = { none: 0, rain: 0.2, storm: 0.32 }

interface Loop {
  gain: GainNode
  stop: () => void
}

// Looping filtered noise whose level is eased with setTargetAtTime (no clicks when it starts or stops).
function noiseLoop(ctx: AudioContext, dest: AudioNode, filters: [BiquadFilterType, number][]): Loop {
  const src = ctx.createBufferSource()
  src.buffer = whiteNoise(ctx)
  src.loop = true
  const gain = ctx.createGain()
  gain.gain.value = 0
  let node: AudioNode = src
  for (const [type, freq] of filters) {
    const f = ctx.createBiquadFilter()
    f.type = type
    f.frequency.value = freq
    node = node.connect(f)
  }
  node.connect(gain).connect(dest)
  src.start()
  return { gain, stop: () => src.stop() }
}

// Random tiny clicks while the fire burns; returns a handle to start/stop the crackle.
function crackler(ctx: AudioContext, dest: AudioNode) {
  let timer: ReturnType<typeof setTimeout> | undefined
  let on = false
  const tick = () => {
    if (!on) return
    burst(ctx, dest, ctx.currentTime, { kind: 'highpass', freq: 1800 + Math.random() * 2500, dur: 0.025, gain: 0.12 + Math.random() * 0.1 })
    timer = setTimeout(tick, 60 + Math.random() * 260)
  }
  return {
    set(next: boolean) {
      if (next && !on) {
        on = true
        tick()
      } else if (!next) {
        on = false
        clearTimeout(timer)
      }
    },
  }
}

/** Continuous sound beds (rain, campfire crackle, wind) fading in and out with the scene's state. */
export function createBeds(ctx: AudioContext, dest: AudioNode) {
  const rain = noiseLoop(ctx, dest, [['highpass', 700], ['lowpass', 7000]])
  const wind = noiseLoop(ctx, dest, [['bandpass', 420]])
  const fire = noiseLoop(ctx, dest, [['lowpass', 260]])
  const crackle = crackler(ctx, dest)
  const ease = (l: Loop, v: number) => l.gain.gain.setTargetAtTime(v, ctx.currentTime, 0.6)
  return {
    set(s: BedState) {
      ease(rain, RAIN_GAIN[s.rain])
      ease(wind, s.wind ? 0.07 : 0)
      ease(fire, s.fire ? 0.08 : 0)
      crackle.set(s.fire)
    },
    stop() {
      crackle.set(false)
      ;[rain, wind, fire].forEach((l) => l.stop())
    },
  }
}
