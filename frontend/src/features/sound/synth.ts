/** Procedural sound recipes: each schedules a few nodes on `dest` starting at `t`; nothing is loaded from files. */
export type Recipe = (ctx: AudioContext, dest: AudioNode, t: number) => void

let noiseBuffer: AudioBuffer | null = null

/** One second of white noise, shared by every noise-based sound. */
export function whiteNoise(ctx: AudioContext): AudioBuffer {
  if (noiseBuffer && noiseBuffer.sampleRate === ctx.sampleRate) return noiseBuffer
  const buf = ctx.createBuffer(1, ctx.sampleRate, ctx.sampleRate)
  const data = buf.getChannelData(0)
  for (let i = 0; i < data.length; i++) data[i] = Math.random() * 2 - 1
  noiseBuffer = buf
  return buf
}

interface Tone {
  type?: OscillatorType
  from: number
  to?: number
  dur: number
  gain?: number
  delay?: number
}

/** A pitched blip with a fast attack and an exponential decay. */
export function tone(ctx: AudioContext, dest: AudioNode, t: number, o: Tone) {
  const at = t + (o.delay ?? 0)
  const osc = ctx.createOscillator()
  const g = ctx.createGain()
  osc.type = o.type ?? 'sine'
  osc.frequency.setValueAtTime(o.from, at)
  if (o.to) osc.frequency.exponentialRampToValueAtTime(o.to, at + o.dur)
  g.gain.setValueAtTime(0.0001, at)
  g.gain.exponentialRampToValueAtTime(o.gain ?? 0.3, at + 0.008)
  g.gain.exponentialRampToValueAtTime(0.0001, at + o.dur)
  osc.connect(g).connect(dest)
  osc.start(at)
  osc.stop(at + o.dur + 0.02)
}

interface Burst {
  kind: BiquadFilterType
  freq: number
  freqTo?: number
  q?: number
  dur: number
  gain?: number
  delay?: number
}

/** A filtered noise burst (clicks, whooshes, rumbles). */
export function burst(ctx: AudioContext, dest: AudioNode, t: number, o: Burst) {
  const at = t + (o.delay ?? 0)
  const src = ctx.createBufferSource()
  const f = ctx.createBiquadFilter()
  const g = ctx.createGain()
  src.buffer = whiteNoise(ctx)
  f.type = o.kind
  f.Q.value = o.q ?? 1
  f.frequency.setValueAtTime(o.freq, at)
  if (o.freqTo) f.frequency.exponentialRampToValueAtTime(o.freqTo, at + o.dur)
  g.gain.setValueAtTime(0.0001, at)
  g.gain.exponentialRampToValueAtTime(o.gain ?? 0.3, at + 0.005)
  g.gain.exponentialRampToValueAtTime(0.0001, at + o.dur)
  src.connect(f).connect(g).connect(dest)
  src.start(at)
  src.stop(at + o.dur + 0.02)
}

const dice: Recipe = (ctx, dest, t) => {
  for (let i = 0; i < 7; i++) {
    const delay = i * 0.06 + Math.random() * 0.04
    burst(ctx, dest, t, { kind: 'bandpass', freq: 2200 + Math.random() * 2200, q: 4, dur: 0.035, gain: 0.35 * (1 - i / 10), delay })
  }
}

const hit: Recipe = (ctx, dest, t) => {
  tone(ctx, dest, t, { from: 150, to: 45, dur: 0.2, gain: 0.7 })
  burst(ctx, dest, t, { kind: 'lowpass', freq: 1400, dur: 0.1, gain: 0.5 })
}

const shot: Recipe = (ctx, dest, t) => {
  burst(ctx, dest, t, { kind: 'highpass', freq: 900, dur: 0.14, gain: 0.7 })
  tone(ctx, dest, t, { from: 240, to: 50, dur: 0.18, gain: 0.5 })
}

const arrow: Recipe = (ctx, dest, t) => {
  burst(ctx, dest, t, { kind: 'bandpass', freq: 1200, freqTo: 3800, q: 2, dur: 0.28, gain: 0.35 })
  tone(ctx, dest, t, { from: 120, to: 70, dur: 0.08, gain: 0.35, delay: 0.3 })
}

const step: Recipe = (ctx, dest, t) => {
  burst(ctx, dest, t, { kind: 'lowpass', freq: 420, dur: 0.08, gain: 0.35 })
  burst(ctx, dest, t, { kind: 'lowpass', freq: 380, dur: 0.08, gain: 0.3, delay: 0.17 })
}

const thunder: Recipe = (ctx, dest, t) => {
  burst(ctx, dest, t, { kind: 'lowpass', freq: 260, freqTo: 70, q: 0.7, dur: 2.4, gain: 0.9 })
  burst(ctx, dest, t, { kind: 'lowpass', freq: 160, dur: 1.6, gain: 0.6, delay: 0.5 })
}

const cheer: Recipe = (ctx, dest, t) => {
  ;[523, 659, 784, 1047].forEach((f, i) => tone(ctx, dest, t, { type: 'triangle', from: f, dur: 0.22, gain: 0.25, delay: i * 0.1 }))
}

const zzz: Recipe = (ctx, dest, t) => {
  tone(ctx, dest, t, { from: 240, to: 190, dur: 0.8, gain: 0.12 })
  tone(ctx, dest, t, { from: 230, to: 180, dur: 0.8, gain: 0.1, delay: 0.9 })
}

const rustle: Recipe = (ctx, dest, t) => {
  for (let i = 0; i < 4; i++) burst(ctx, dest, t, { kind: 'bandpass', freq: 3000, q: 1.5, dur: 0.07, gain: 0.2, delay: i * 0.09 })
}

const clink: Recipe = (ctx, dest, t) => {
  tone(ctx, dest, t, { from: 1800, dur: 0.12, gain: 0.2 })
  tone(ctx, dest, t, { from: 2400, dur: 0.1, gain: 0.15, delay: 0.07 })
}

const chime: Recipe = (ctx, dest, t) => {
  tone(ctx, dest, t, { from: 880, dur: 0.5, gain: 0.2 })
  tone(ctx, dest, t, { from: 1320, dur: 0.45, gain: 0.12, delay: 0.05 })
}

const murmur: Recipe = (ctx, dest, t) => {
  tone(ctx, dest, t, { type: 'triangle', from: 170, to: 210, dur: 0.18, gain: 0.2 })
  tone(ctx, dest, t, { type: 'triangle', from: 200, to: 150, dur: 0.22, gain: 0.2, delay: 0.22 })
}

const whoosh: Recipe = (ctx, dest, t) => burst(ctx, dest, t, { kind: 'bandpass', freq: 300, freqTo: 900, q: 0.8, dur: 0.9, gain: 0.18 })

export const RECIPES = { dice, hit, shot, arrow, step, thunder, cheer, zzz, rustle, clink, chime, murmur, whoosh }

export type SoundName = keyof typeof RECIPES
