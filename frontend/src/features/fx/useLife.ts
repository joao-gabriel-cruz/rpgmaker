import { useFrame } from '@react-three/fiber'
import { useRef } from 'react'
import type { Fx } from './effects'

/** Progress of an effect (0..1 after its delay); -1 while waiting. Calls onDone once when it ends. */
export function useLife(fx: Fx, onDone: (id: number) => void) {
  const born = useRef<number | null>(null)
  const done = useRef(false)
  const t = useRef(-1)
  useFrame(() => {
    born.current ??= performance.now()
    const age = performance.now() - born.current - fx.delay
    t.current = age < 0 ? -1 : Math.min(1, age / fx.ms)
    if (age >= fx.ms && !done.current) {
      done.current = true
      onDone(fx.id)
    }
  })
  return t
}

/** Deterministic pseudo-random in [0,1) from a seed and an index (same on every client). */
export function rand(seed: number, i: number): number {
  const x = Math.sin(seed * 12.9898 + i * 78.233) * 43758.5453
  return x - Math.floor(x)
}
