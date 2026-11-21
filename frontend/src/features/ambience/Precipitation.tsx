import { useFrame } from '@react-three/fiber'
import { useLayoutEffect, useMemo, useRef } from 'react'
import { Object3D, type InstancedMesh } from 'three'
import type { Weather } from './ambience'
import { useCanvasActive } from './useCanvasActive'

interface Kind {
  count: number
  speed: number
  drift: number
  size: [number, number, number]
  color: string
  opacity: number
}

// A few hundred instances only: enough to read as weather without costing the frame rate.
export const PRECIPITATION: Partial<Record<Weather, Kind>> = {
  rain: { count: 420, speed: 16, drift: 1.5, size: [0.02, 0.55, 0.02], color: '#9ec5ff', opacity: 0.35 },
  storm: { count: 560, speed: 22, drift: 4, size: [0.02, 0.7, 0.02], color: '#9ec5ff', opacity: 0.4 },
  snow: { count: 280, speed: 1.4, drift: 0.5, size: [0.07, 0.07, 0.07], color: '#ffffff', opacity: 0.85 },
  ash: { count: 220, speed: 0.8, drift: 0.9, size: [0.06, 0.06, 0.06], color: '#c9b9a8', opacity: 0.7 },
}

const TOP = 12
const MAX_SPAN = 40

function seed(count: number, w: number, d: number) {
  return Float32Array.from({ length: count * 3 }, (_, i) => (i % 3 === 0 ? Math.random() * w : i % 3 === 1 ? Math.random() * TOP : Math.random() * d))
}

interface Area {
  w: number
  d: number
  offX: number
  offZ: number
}

// Moves every particle one step (wrapping at the bottom and the right edge) and writes its matrix.
function advance(m: InstancedMesh, pos: Float32Array, kind: Kind, dt: number, a: Area, dummy: Object3D) {
  for (let i = 0; i < kind.count; i++) {
    pos[i * 3 + 1] -= kind.speed * dt
    pos[i * 3] += kind.drift * dt * 0.3
    if (pos[i * 3 + 1] < 0) pos[i * 3 + 1] += TOP
    if (pos[i * 3] > a.w) pos[i * 3] -= a.w
    dummy.position.set(pos[i * 3] + a.offX, pos[i * 3 + 1], pos[i * 3 + 2] + a.offZ)
    dummy.rotation.z = kind.drift > 1 ? 0.12 : 0
    dummy.updateMatrix()
    m.setMatrixAt(i, dummy.matrix)
  }
  m.instanceMatrix.needsUpdate = true
}

function Falling({ kind, width, depth, still }: { kind: Kind; width: number; depth: number; still: boolean }) {
  const mesh = useRef<InstancedMesh>(null)
  const active = useCanvasActive()
  const area = useMemo<Area>(() => {
    const w = Math.min(width, MAX_SPAN)
    const d = Math.min(depth, MAX_SPAN)
    return { w, d, offX: (width - w) / 2, offZ: (depth - d) / 2 }
  }, [width, depth])
  const pos = useMemo(() => seed(kind.count, area.w, area.d), [kind.count, area])
  const dummy = useMemo(() => new Object3D(), [])
  const step = (dt: number) => mesh.current && advance(mesh.current, pos, kind, dt, area, dummy)
  useLayoutEffect(() => { step(0) })
  useFrame((_, dt) => { if (!still && active.current) step(Math.min(dt, 0.05)) })
  return (
    <instancedMesh ref={mesh} args={[undefined, undefined, kind.count]} frustumCulled={false} raycast={() => null}>
      <boxGeometry args={kind.size} />
      <meshBasicMaterial color={kind.color} transparent opacity={kind.opacity} fog={false} depthWrite={false} />
    </instancedMesh>
  )
}

/** Instanced rain, snow or ash over the map; `still` freezes it (reduced motion keeps a static snapshot). */
export function Precipitation({ weather, width, depth, still }: { weather: Weather; width: number; depth: number; still: boolean }) {
  const kind = PRECIPITATION[weather]
  if (!kind) return null
  return <Falling key={weather} kind={kind} width={width} depth={depth} still={still} />
}
