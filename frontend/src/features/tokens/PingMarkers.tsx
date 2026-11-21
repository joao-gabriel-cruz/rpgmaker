import { useFrame } from '@react-three/fiber'
import { useRef } from 'react'
import type { Mesh, MeshBasicMaterial } from 'three'
import type { MapGrid } from '../map3d/grid'
import { getCell, surfaceHeight } from '../map3d/grid'
import type { PingMarker } from './tokens'

function Ring({ ping, y }: { ping: PingMarker; y: number }) {
  const mesh = useRef<Mesh>(null)
  const born = useRef<number | null>(null)
  useFrame((state) => {
    born.current ??= state.clock.elapsedTime
    const t = (state.clock.elapsedTime - born.current) % 1
    mesh.current?.scale.setScalar(0.5 + t * 1.5)
    ;(mesh.current?.material as MeshBasicMaterial | undefined)?.setValues({ opacity: 1 - t })
  })
  return (
    <mesh ref={mesh} position={[ping.x, y + 0.08, ping.z]} rotation-x={-Math.PI / 2} raycast={() => null}>
      <ringGeometry args={[0.8, 1, 32]} />
      <meshBasicMaterial color="#fbbf24" transparent depthTest={false} />
    </mesh>
  )
}

/** Animated rings on the map for each active ping. */
export function PingMarkers({ grid, pings }: { grid: MapGrid; pings: readonly PingMarker[] }) {
  return (
    <>
      {pings.map((p) => (
        <Ring key={p.id} ping={p} y={surfaceHeight(getCell(grid, Math.floor(p.x), Math.floor(p.z))?.height ?? 1)} />
      ))}
    </>
  )
}
