import { Box, Cone, Cyl } from './parts'
import { C } from './palette'
import type { PrefabProps } from './types'

export function LampPost() {
  return (
    <>
      <Cyl r={0.12} h={0.2} seg={6} pos={[0, 0.1, 0]} color={C.dark} />
      <Cyl r={0.06} h={2.8} seg={6} pos={[0, 1.4, 0]} color={C.metal} />
      <Box size={[0.3, 0.32, 0.3]} pos={[0, 2.95, 0]} color="#fde68a" emissive="#fbbf24" />
      <Box size={[0.36, 0.05, 0.36]} pos={[0, 2.75, 0]} color={C.dark} />
      <Cone r={0.3} h={0.22} seg={4} pos={[0, 3.2, 0]} rot={[0, Math.PI / 4, 0]} color={C.dark} />
    </>
  )
}

const JAGGED: [number, number, number, number][] = [
  // x, height, width, rotation
  [-0.35, 0.7, 0.3, 0.05], [-0.05, 1.05, 0.28, -0.08], [0.25, 0.55, 0.3, 0.1], [0.42, 0.85, 0.16, -0.05],
]

export function RuinedWall({ props }: PrefabProps) {
  const tall = props.height === 2 || props.height === '2'
  const scale = tall ? 1.7 : 1
  return (
    <>
      <Box size={[1, 0.35, 0.42]} pos={[0, 0.18, 0]} color={C.brick} />
      {JAGGED.map(([x, h, w, r], i) => (
        <Box key={i} size={[w, h * scale, 0.4]} pos={[x, 0.3 + (h * scale) / 2, 0]} rot={[0, 0, r]} color={i % 2 ? C.brick : '#7b7d82'} />
      ))}
      <Box size={[0.12, 0.1, 0.1]} pos={[0.1, 0.07, 0.35]} color={C.stone} />
    </>
  )
}

const FLAG_COLORS = ['#d9534f', '#e8c547', '#4f8fd9', '#5cb85c', '#f0f0f0', '#d9772b']
const SEGMENTS = 8

// String sagging between two poles; local x runs along the 2-cell footprint.
export function Bunting() {
  const sag = (x: number) => 1.55 - 0.3 * (1 - (x / 0.9) ** 2)
  const pts = Array.from({ length: SEGMENTS + 1 }, (_, i) => -0.9 + (1.8 * i) / SEGMENTS)
  return (
    <>
      {[-0.9, 0.9].map((x) => (
        <Cyl key={x} r={0.04} h={1.7} seg={5} pos={[x, 0.85, 0]} color={C.darkWood} />
      ))}
      {pts.slice(1).map((x, i) => {
        const a = pts[i]
        const len = Math.hypot(x - a, sag(x) - sag(a))
        return <Box key={x} size={[len, 0.02, 0.02]} pos={[(a + x) / 2, (sag(a) + sag(x)) / 2, 0]} rot={[0, 0, Math.atan2(sag(x) - sag(a), x - a)]} color={C.dark} />
      })}
      {pts.slice(1, -1).map((x, i) => (
        <Cone key={x} r={0.1} h={0.24} seg={3} pos={[x, sag(x) - 0.13, 0]} rot={[Math.PI, 0, 0]} color={FLAG_COLORS[i % FLAG_COLORS.length]} />
      ))}
    </>
  )
}
