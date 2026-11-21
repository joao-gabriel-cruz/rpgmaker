import { Billboard } from '@react-three/drei'
import { useFrame } from '@react-three/fiber'
import { useMemo, useRef } from 'react'
import type { Group, Mesh } from 'three'
import type { Fx } from './effects'
import { glyphTexture } from './glyph'
import { rand, useLife } from './useLife'

interface P {
  fx: Fx
  y: number
  onDone: (id: number) => void
}

/** A glyph floating up and fading: notes, zzz, glints, the healing cross, the damage number. */
export function Rise({ fx, y, onDone }: P) {
  const life = useLife(fx, onDone)
  const group = useRef<Group>(null)
  const tex = useMemo(() => glyphTexture(fx.glyph ?? '✦', fx.color), [fx.glyph, fx.color])
  const side = (rand(fx.seed, 1) - 0.5) * 0.5
  const number = fx.kind === 'number'
  useFrame(() => {
    const t = life.current
    const g = group.current
    if (!g) return
    g.visible = t >= 0
    if (t < 0) return
    g.position.set(fx.x + 0.5 + side * t, y + 1.2 + t * (number ? 0.9 : 0.7), fx.z + 0.5)
    const mat = (g.children[0] as Mesh | undefined)?.material as { opacity: number } | undefined
    if (mat) mat.opacity = Math.min(1, (1 - t) * 3)
  })
  const size = number ? 0.7 : 0.45
  return (
    <group ref={group} visible={false}>
      <Billboard>
        <mesh>
          <planeGeometry args={[size * (number ? 1.6 : 1), size]} />
          <meshBasicMaterial map={tex} transparent depthWrite={false} />
        </mesh>
      </Billboard>
    </group>
  )
}

/** Speech balloon with a gesture icon above the actor; shakes slightly while it lasts. */
export function Bubble({ fx, y, onDone }: P) {
  const life = useLife(fx, onDone)
  const group = useRef<Group>(null)
  const tex = useMemo(() => glyphTexture(fx.glyph ?? '💬', fx.color ?? '#18181b', true), [fx.glyph, fx.color])
  useFrame(() => {
    const t = life.current
    const g = group.current
    if (!g) return
    g.visible = t >= 0 && t < 1
    if (t < 0) return
    const pop = Math.min(1, t * 8)
    g.scale.setScalar(pop * (1 + (t > 0.8 ? (t - 0.8) * -2.5 : 0)))
    g.position.set(fx.x + 0.5 + Math.sin(t * 40) * 0.01, y + 1.95, fx.z + 0.5)
  })
  return (
    <group ref={group} visible={false}>
      <Billboard>
        <mesh>
          <planeGeometry args={[0.8, 0.8]} />
          <meshBasicMaterial map={tex} transparent depthWrite={false} />
        </mesh>
      </Billboard>
    </group>
  )
}

/** Expanding ground ring: the perception wave. */
export function Wave({ fx, y, onDone }: P) {
  const life = useLife(fx, onDone)
  const ring = useRef<Mesh>(null)
  useFrame(() => {
    const t = life.current
    const m = ring.current
    if (!m) return
    m.visible = t >= 0
    if (t < 0) return
    m.scale.setScalar(0.4 + t * 5)
    ;(m.material as { opacity: number }).opacity = 0.8 * (1 - t)
  })
  return (
    <mesh ref={ring} position={[fx.x + 0.5, y + 0.08, fx.z + 0.5]} rotation={[-Math.PI / 2, 0, 0]} visible={false}>
      <ringGeometry args={[0.9, 1, 40]} />
      <meshBasicMaterial color={fx.color ?? '#38bdf8'} transparent depthWrite={false} />
    </mesh>
  )
}

/** Reduced motion: one static icon that appears and disappears without moving. */
export function StaticFlash({ fx, y, onDone }: P) {
  const life = useLife(fx, onDone)
  const group = useRef<Group>(null)
  const tex = useMemo(() => glyphTexture(fx.glyph ?? '✦', fx.color), [fx.glyph, fx.color])
  useFrame(() => {
    if (group.current) group.current.visible = life.current >= 0 && life.current < 1
  })
  return (
    <group ref={group} position={[fx.x + 0.5, y + 1.95, fx.z + 0.5]} visible={false}>
      <Billboard>
        <mesh>
          <planeGeometry args={[0.55, 0.55]} />
          <meshBasicMaterial map={tex} transparent depthWrite={false} />
        </mesh>
      </Billboard>
    </group>
  )
}
