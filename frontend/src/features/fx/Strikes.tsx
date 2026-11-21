import { useFrame } from '@react-three/fiber'
import { useRef } from 'react'
import type { Group, Mesh } from 'three'
import type { Fx } from './effects'
import { useLife } from './useLife'

interface P {
  fx: Fx
  y: number
  onDone: (id: number) => void
}

const heading = (fx: Fx) => Math.atan2((fx.toX ?? fx.x) - fx.x, (fx.toZ ?? fx.z) - fx.z)

/** Sweeping arc of a melee blow (blunt weapons get a heavier orange ring). */
export function Arc({ fx, y, onDone }: P) {
  const life = useLife(fx, onDone)
  const group = useRef<Group>(null)
  const sweep = useRef<Mesh>(null)
  const blunt = fx.variant === 'blunt'
  useFrame(() => {
    const t = life.current
    if (!group.current) return
    group.current.visible = t >= 0
    if (t < 0 || !sweep.current) return
    sweep.current.rotation.x = -1.6 + t * 3.2
    sweep.current.scale.setScalar(0.8 + t * 0.4)
    ;(sweep.current.material as { opacity: number }).opacity = 1 - t * t
  })
  const mid = { x: (fx.x + (fx.toX ?? fx.x)) / 2, z: (fx.z + (fx.toZ ?? fx.z)) / 2 }
  return (
    <group ref={group} position={[mid.x + 0.5, y + 0.6, mid.z + 0.5]} rotation={[0, heading(fx), 0]} visible={false}>
      <mesh ref={sweep} rotation={[0, Math.PI / 2, 0]}>
        <ringGeometry args={[0.35, blunt ? 0.62 : 0.5, 20, 1, 0, Math.PI * 0.85]} />
        <meshBasicMaterial color={blunt ? '#fb923c' : '#f4f4f5'} transparent side={2} depthWrite={false} />
      </mesh>
    </group>
  )
}

/** Projectile flying from the shooter to the target: an arrow or a bright shot, with a short trail. */
export function Projectile({ fx, y, onDone }: P) {
  const life = useLife(fx, onDone)
  const group = useRef<Group>(null)
  const arrow = fx.variant !== 'shot'
  useFrame(() => {
    const t = life.current
    if (!group.current) return
    group.current.visible = t >= 0
    if (t < 0) return
    const to = { x: fx.toX ?? fx.x, z: fx.toZ ?? fx.z }
    group.current.position.set(fx.x + (to.x - fx.x) * t + 0.5, y + 0.7 - 0.15 * Math.sin(t * Math.PI), fx.z + (to.z - fx.z) * t + 0.5)
  })
  return (
    <group ref={group} rotation={[0, heading(fx), 0]} visible={false}>
      {[0, 1, 2, 3].map((i) => (
        <mesh key={i} position={[0, 0, -0.12 * i]}>
          <boxGeometry args={arrow ? [0.025, 0.025, 0.28] : [0.05, 0.05, 0.1]} />
          <meshBasicMaterial color={arrow ? '#e7e5e4' : '#fde047'} transparent opacity={1 - i * 0.25} />
        </mesh>
      ))}
    </group>
  )
}
