import { useFrame } from '@react-three/fiber'
import { useRef } from 'react'
import type { PointLight } from 'three'
import type { LightPoint } from './lightPoints'

/** Radius of a carried light in map cells. */
export const LIGHT_RADIUS = 4.5

function Glow({ p, level, still }: { p: LightPoint; level: { current: number }; still: boolean }) {
  const ref = useRef<PointLight>(null)
  const base = p.fire ? 26 : 18
  useFrame(({ clock }) => {
    if (!ref.current) return
    const flicker = still ? 1 : 0.92 + 0.08 * Math.sin(clock.elapsedTime * 11 + p.x * 3) * Math.sin(clock.elapsedTime * 5.3 + p.z)
    ref.current.intensity = base * level.current * flicker
  })
  return <pointLight ref={ref} position={[p.x, p.y, p.z]} color={p.fire ? '#ff9340' : '#ffcf8a'} distance={LIGHT_RADIUS} decay={1.6} intensity={0} />
}

/** Warm point lights around torch bearers and campfires; `level` (0..1) follows how dark the scene is. */
export function LightSources({ points, level, still }: { points: LightPoint[]; level: { current: number }; still: boolean }) {
  return <>{points.map((p) => <Glow key={p.key} p={p} level={level} still={still} />)}</>
}
