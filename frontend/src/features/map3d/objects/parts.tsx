import type { Material } from 'three'
import type { V3 } from './palette'

interface PartProps {
  pos?: V3
  rot?: V3
  scale?: V3
  color: string
  emissive?: string
  material?: Material
}

function Mat({ color, emissive, material }: Pick<PartProps, 'color' | 'emissive' | 'material'>) {
  if (material) return null
  return (
    <meshStandardMaterial
      color={color}
      emissive={emissive ?? '#000000'}
      emissiveIntensity={emissive ? 1 : 0}
      flatShading
      roughness={0.9}
    />
  )
}

export function Box({ size, pos, rot, scale, ...m }: PartProps & { size: V3 }) {
  return (
    <mesh position={pos} rotation={rot} scale={scale} material={m.material}>
      <boxGeometry args={size} />
      <Mat {...m} />
    </mesh>
  )
}

export function Cyl({ r, rTop, h, seg = 8, pos, rot, scale, ...m }: PartProps & { r: number; rTop?: number; h: number; seg?: number }) {
  return (
    <mesh position={pos} rotation={rot} scale={scale} material={m.material}>
      <cylinderGeometry args={[rTop ?? r, r, h, seg]} />
      <Mat {...m} />
    </mesh>
  )
}

export function Cone({ r, h, seg = 7, pos, rot, scale, ...m }: PartProps & { r: number; h: number; seg?: number }) {
  return (
    <mesh position={pos} rotation={rot} scale={scale} material={m.material}>
      <coneGeometry args={[r, h, seg]} />
      <Mat {...m} />
    </mesh>
  )
}

export function Rock({ r, pos, rot, scale, ...m }: PartProps & { r: number }) {
  return (
    <mesh position={pos} rotation={rot} scale={scale} material={m.material}>
      <dodecahedronGeometry args={[r, 0]} />
      <Mat {...m} />
    </mesh>
  )
}
