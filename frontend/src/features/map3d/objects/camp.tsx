import { Box, Cone, Cyl } from './parts'
import { C, awningColor } from './palette'
import type { PrefabProps } from './types'

export function Tent() {
  const stakes: [number, number][] = [[-1, -1], [1, -1], [-1, 1], [1, 1]]
  return (
    <>
      <Cone r={1.25} h={1.7} seg={4} pos={[0, 0.85, 0]} rot={[0, Math.PI / 4, 0]} color={C.canvas} />
      <Box size={[0.5, 0.9, 0.05]} pos={[0, 0.45, 0.88]} rot={[-0.18, 0, 0]} color={C.dark} />
      <Cone r={1.25} h={0.18} seg={4} pos={[0.05, 0.08, 0.02]} rot={[0, Math.PI / 4, 0]} color={C.darkWood} />
      {stakes.map(([x, z]) => (
        <Box key={`${x}${z}`} size={[0.07, 0.35, 0.07]} pos={[x * 0.95, 0.17, z * 0.95]} rot={[0.2 * z, 0, -0.2 * x]} color={C.darkWood} />
      ))}
    </>
  )
}

export function Barrel() {
  return (
    <>
      <Cyl r={0.3} rTop={0.27} h={0.85} seg={8} pos={[0, 0.43, 0]} color={C.oldRust} />
      <Cyl r={0.325} rTop={0.325} h={0.07} seg={8} pos={[0, 0.2, 0]} color={C.metal} />
      <Cyl r={0.325} rTop={0.325} h={0.07} seg={8} pos={[0, 0.65, 0]} color={C.metal} />
      <Cyl r={0.25} h={0.03} seg={8} pos={[0, 0.86, 0]} color={C.dark} />
    </>
  )
}

export function Bench() {
  const legs: [number, number][] = [[-0.8, -0.14], [0.8, -0.14], [-0.8, 0.14], [0.8, 0.14]]
  return (
    <>
      <Box size={[1.9, 0.07, 0.38]} pos={[0, 0.45, 0.04]} color={C.wood} />
      <Box size={[1.9, 0.3, 0.06]} pos={[0, 0.78, -0.17]} rot={[-0.12, 0, 0]} color={C.wood} />
      {legs.map(([x, z]) => (
        <Box key={`${x}${z}`} size={[0.08, 0.45, 0.07]} pos={[x, 0.22, z]} color={C.metal} />
      ))}
    </>
  )
}

export function MarketStall({ props }: PrefabProps) {
  const awning = awningColor(props.color)
  const poles: [number, number][] = [[-0.9, -0.85], [0.9, -0.85], [-0.9, 0.85], [0.9, 0.85]]
  return (
    <>
      {poles.map(([x, z]) => (
        <Box key={`${x}${z}`} size={[0.08, z < 0 ? 1.9 : 1.65, 0.08]} pos={[x, z < 0 ? 0.95 : 0.82, z]} color={C.darkWood} />
      ))}
      {Array.from({ length: 6 }, (_, i) => (
        <Box key={i} size={[0.32, 0.05, 2.15]} pos={[-0.8 + i * 0.32, 1.84, 0]} rot={[0.14, 0, 0]} color={i % 2 ? '#e8e0cc' : awning} />
      ))}
      <Box size={[1.7, 0.55, 0.6]} pos={[0, 0.28, 0.45]} color={C.wood} />
      <Box size={[1.8, 0.06, 0.7]} pos={[0, 0.58, 0.45]} color={C.darkWood} />
      <Box size={[0.3, 0.2, 0.25]} pos={[-0.5, 0.71, 0.45]} color={C.leafDark} />
      <Box size={[0.3, 0.2, 0.25]} pos={[0.05, 0.71, 0.45]} color={C.rust} />
      <Box size={[0.25, 0.2, 0.25]} pos={[0.55, 0.71, 0.45]} color={C.sand} />
    </>
  )
}

export function Well() {
  return (
    <>
      <Cyl r={0.42} h={0.55} seg={8} pos={[0, 0.27, 0]} color={C.stone} />
      <Cyl r={0.3} h={0.02} seg={8} pos={[0, 0.56, 0]} color="#2d4a5e" />
      <Box size={[0.08, 1.1, 0.08]} pos={[-0.4, 0.85, 0]} color={C.darkWood} />
      <Box size={[0.08, 1.1, 0.08]} pos={[0.4, 0.85, 0]} color={C.darkWood} />
      <Box size={[0.9, 0.06, 0.06]} pos={[0, 1.0, 0]} color={C.darkWood} />
      <Cone r={0.62} h={0.4} seg={4} pos={[0, 1.55, 0]} rot={[0, Math.PI / 4, 0]} color={C.rust} />
    </>
  )
}

export function Sign() {
  return (
    <group rotation={[0, 0.25, 0.08]}>
      <Box size={[0.08, 0.95, 0.08]} pos={[0, 0.47, 0]} color={C.darkWood} />
      <Box size={[0.62, 0.34, 0.05]} pos={[0.04, 0.8, 0.06]} color={C.wood} />
      <Box size={[0.44, 0.04, 0.06]} pos={[0.04, 0.84, 0.09]} color={C.dark} />
      <Box size={[0.34, 0.04, 0.06]} pos={[0.0, 0.76, 0.09]} color={C.dark} />
    </group>
  )
}

export function Sandbags() {
  const bags: [number, number, number][] = [[-0.3, 0.12, -0.15], [0.3, 0.12, -0.15], [-0.3, 0.12, 0.2], [0.3, 0.12, 0.2], [0, 0.34, 0.02]]
  return (
    <>
      {bags.map(([x, y, z], i) => (
        <Box key={i} size={[0.52, 0.22, 0.34]} pos={[x, y, z]} rot={[0, (i % 2 ? 0.1 : -0.08), 0]} color={i % 2 ? C.sand : '#a89368'} />
      ))}
    </>
  )
}
