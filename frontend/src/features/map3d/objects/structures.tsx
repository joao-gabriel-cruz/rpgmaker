import { Box, Cone, Cyl } from './parts'
import { C } from './palette'
import type { PrefabProps } from './types'

export function Wall() {
  return (
    <>
      <Box size={[1, 2.9, 1]} pos={[0, 1.45, 0]} color={C.brick} />
      <Box size={[1.06, 0.18, 1.06]} pos={[0, 2.9, 0]} color={C.stone} />
    </>
  )
}

export function Barricade() {
  return (
    <>
      <Box size={[0.12, 1, 0.12]} pos={[-0.4, 0.5, 0]} color={C.darkWood} />
      <Box size={[0.12, 1, 0.12]} pos={[0.4, 0.5, 0]} color={C.darkWood} />
      <Box size={[1, 0.18, 0.12]} pos={[0, 0.3, 0.1]} rot={[0, 0, 0.12]} color={C.wood} />
      <Box size={[1, 0.18, 0.12]} pos={[0, 0.6, -0.1]} rot={[0, 0, -0.1]} color={C.wood} />
      <Box size={[1, 0.18, 0.12]} pos={[0, 0.85, 0.1]} rot={[0, 0, 0.08]} color={C.wood} />
    </>
  )
}

export function Fence() {
  return (
    <>
      <Box size={[0.1, 0.9, 0.1]} pos={[-0.45, 0.45, 0]} color={C.darkWood} />
      <Box size={[0.1, 0.9, 0.1]} pos={[0.45, 0.45, 0]} color={C.darkWood} />
      <Box size={[1, 0.08, 0.06]} pos={[0, 0.35, 0]} color={C.wood} />
      <Box size={[1, 0.08, 0.06]} pos={[0, 0.7, 0]} color={C.wood} />
    </>
  )
}

export function Door({ props }: PrefabProps) {
  const open = props.open === true
  return (
    <>
      <Box size={[0.12, 3, 0.24]} pos={[-0.44, 1.5, 0]} color={C.stone} />
      <Box size={[0.12, 3, 0.24]} pos={[0.44, 1.5, 0]} color={C.stone} />
      <Box size={[1, 0.14, 0.24]} pos={[0, 2.93, 0]} color={C.stone} />
      <group position={[-0.38, 0, 0]} rotation={[0, open ? -Math.PI / 2 : 0, 0]}>
        <Box size={[0.76, 2.85, 0.08]} pos={[0.38, 1.42, 0]} color={C.wood} />
        <Box size={[0.08, 0.08, 0.12]} pos={[0.66, 1.4, 0]} color={C.metal} />
      </group>
    </>
  )
}

export function Watchtower() {
  const legs: [number, number][] = [[-0.4, -0.4], [0.4, -0.4], [-0.4, 0.4], [0.4, 0.4]]
  return (
    <>
      {legs.map(([x, z]) => (
        <Box key={`${x}${z}`} size={[0.12, 3, 0.12]} pos={[x, 1.5, z]} color={C.darkWood} />
      ))}
      <Box size={[1, 0.15, 1]} pos={[0, 3.05, 0]} color={C.wood} />
      <Box size={[0.8, 0.9, 0.8]} pos={[0, 3.6, 0]} color={C.wood} />
      <Box size={[0.6, 0.12, 0.06]} pos={[0, 3.75, 0.41]} color={C.dark} />
      <Cone r={0.7} h={1} seg={4} pos={[0, 4.55, 0]} rot={[0, Math.PI / 4, 0]} color={C.rust} />
    </>
  )
}

export function Post() {
  return (
    <>
      <Cyl r={0.08} h={2} seg={6} pos={[0, 1, 0]} color={C.metal} />
      <Box size={[0.5, 0.06, 0.06]} pos={[0.2, 1.95, 0]} color={C.metal} />
      <Box size={[0.2, 0.1, 0.2]} pos={[0.4, 1.88, 0]} color="#fde68a" emissive="#fbbf24" />
    </>
  )
}
