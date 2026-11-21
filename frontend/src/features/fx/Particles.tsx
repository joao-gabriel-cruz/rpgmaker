import { useFrame } from '@react-three/fiber'
import { useRef } from 'react'
import type { Group } from 'three'
import type { Fx } from './effects'
import { rand, useLife } from './useLife'

interface P {
  fx: Fx
  y: number
  onDone: (id: number) => void
}

const CONFETTI = ['#f472b6', '#facc15', '#60a5fa', '#4ade80', '#fb923c', '#a78bfa']

interface BurstProps extends P {
  count: number
  size: number
  speed: number
  lift: number
  colors: string[]
}

function Burst({ fx, y, onDone, count, size, speed, lift, colors }: BurstProps) {
  const life = useLife(fx, onDone)
  const group = useRef<Group>(null)
  useFrame(() => {
    const t = life.current
    const g = group.current
    if (!g) return
    g.visible = t >= 0
    if (t < 0) return
    g.children.forEach((c, i) => {
      const a = rand(fx.seed, i) * Math.PI * 2
      const v = speed * (0.5 + rand(fx.seed, i + 50))
      c.position.set(Math.cos(a) * v * t, lift * t - 1.2 * t * t + 0.4, Math.sin(a) * v * t)
      c.rotation.set(t * 8 * rand(fx.seed, i + 9), t * 6, 0)
      c.scale.setScalar(1 - t * 0.7)
    })
  })
  return (
    <group ref={group} position={[fx.x + 0.5, y + 0.2, fx.z + 0.5]} visible={false}>
      {Array.from({ length: count }, (_, i) => (
        <mesh key={i}>
          <boxGeometry args={[size, size, size * 0.4]} />
          <meshBasicMaterial color={colors[i % colors.length]} />
        </mesh>
      ))}
    </group>
  )
}

/** Impact or work sparks flying outwards and falling. */
export const Sparks = (p: P) => (
  <Burst {...p} count={10} size={0.06} speed={0.9} lift={1.4} colors={[p.fx.color ?? '#fbbf24', '#fef08a']} />
)

/** Party confetti. */
export const Confetti = (p: P) => <Burst {...p} count={26} size={0.07} speed={1.3} lift={2.4} colors={CONFETTI} />
