import { useFrame } from '@react-three/fiber'
import { useRef } from 'react'
import type { Group, PointLight } from 'three'
import { Box, Cone, Cyl } from './parts'
import { C } from './palette'

export function Car() {
  const wheels: [number, number][] = [[-0.6, -0.42], [0.6, -0.42], [-0.6, 0.42], [0.6, 0.42]]
  return (
    <>
      <Box size={[1.9, 0.4, 0.9]} pos={[0, 0.4, 0]} color={C.rust} />
      <Box size={[1, 0.35, 0.8]} pos={[-0.15, 0.78, 0]} color="#7d3a22" />
      <Box size={[0.9, 0.2, 0.82]} pos={[-0.15, 0.8, 0]} color={C.glass} />
      {wheels.map(([x, z]) => (
        <Cyl key={`${x}${z}`} r={0.22} h={0.14} seg={8} pos={[x, 0.22, z]} rot={[Math.PI / 2, 0, 0]} color={C.dark} />
      ))}
    </>
  )
}

export function Crate() {
  return (
    <>
      <Box size={[0.8, 0.8, 0.8]} pos={[0, 0.4, 0]} color={C.wood} />
      <Box size={[0.86, 0.1, 0.86]} pos={[0, 0.1, 0]} color={C.darkWood} />
      <Box size={[0.86, 0.1, 0.86]} pos={[0, 0.7, 0]} color={C.darkWood} />
    </>
  )
}

export function Campfire() {
  const light = useRef<PointLight>(null)
  const flame = useRef<Group>(null)
  useFrame(({ clock }) => {
    const t = clock.elapsedTime
    const flicker = Math.sin(t * 13) * 0.15 + Math.sin(t * 7.3) * 0.1
    if (light.current) light.current.intensity = 2.2 + flicker * 4
    if (flame.current) flame.current.scale.set(1, 1 + flicker, 1)
  })
  return (
    <>
      <Box size={[0.7, 0.12, 0.12]} pos={[0, 0.08, 0]} rot={[0, 0.5, 0]} color={C.darkWood} />
      <Box size={[0.7, 0.12, 0.12]} pos={[0, 0.08, 0]} rot={[0, -0.5, 0]} color={C.darkWood} />
      <Box size={[0.7, 0.12, 0.12]} pos={[0, 0.08, 0]} rot={[0, 1.57, 0]} color={C.trunk} />
      <group ref={flame}>
        <Cone r={0.22} h={0.6} seg={5} pos={[0, 0.4, 0]} color={C.fire} emissive="#ff6a00" />
        <Cone r={0.12} h={0.4} seg={5} pos={[0, 0.3, 0]} color="#fde047" emissive="#fbbf24" />
      </group>
      <pointLight ref={light} position={[0, 0.8, 0]} color="#ff8a3d" distance={7} decay={2} />
    </>
  )
}
