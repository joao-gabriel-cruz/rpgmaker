import { useFrame } from '@react-three/fiber'
import { useMemo } from 'react'
import { MeshStandardMaterial } from 'three'
import { Box, Cone, Cyl, Rock } from './parts'
import { C } from './palette'
import type { PrefabProps } from './types'

function TreePine() {
  return (
    <>
      <Cyl r={0.12} h={0.9} seg={6} pos={[0, 0.45, 0]} color={C.trunk} />
      <Cone r={0.75} h={1.1} pos={[0, 1.2, 0]} color={C.leafDark} />
      <Cone r={0.58} h={1} pos={[0, 1.9, 0]} color={C.leaf} />
      <Cone r={0.4} h={0.9} pos={[0, 2.55, 0]} color={C.leafDark} />
    </>
  )
}

function TreeRound() {
  return (
    <>
      <Cyl r={0.14} rTop={0.1} h={1.3} seg={6} pos={[0, 0.65, 0]} color={C.trunk} />
      <Rock r={0.7} pos={[0, 1.7, 0]} scale={[1, 0.9, 1]} color={C.leaf} />
      <Rock r={0.45} pos={[0.4, 1.35, 0.2]} color={C.leafDark} />
      <Rock r={0.4} pos={[-0.3, 2.1, -0.2]} color={C.leaf} />
    </>
  )
}

export function Bush() {
  return (
    <>
      <Rock r={0.4} pos={[0, 0.3, 0]} scale={[1.2, 0.8, 1.1]} color={C.leafDark} />
      <Rock r={0.3} pos={[0.25, 0.35, 0.2]} color={C.leaf} />
    </>
  )
}

export function BigRock() {
  return (
    <group scale={0.5}>
      <Rock r={0.95} pos={[0, 0.7, 0]} scale={[1, 0.8, 1]} color={C.stone} />
      <Rock r={0.6} pos={[0.55, 0.4, -0.4]} color="#6b727b" />
      <Rock r={0.4} pos={[-0.6, 0.3, 0.5]} color="#8a919b" />
    </group>
  )
}

function useLavaMaterial() {
  const material = useMemo(
    () => new MeshStandardMaterial({ color: C.lava, emissive: '#ff4a00', flatShading: true }),
    [],
  )
  useFrame(({ clock }) => {
    material.emissiveIntensity = 0.9 + Math.sin(clock.elapsedTime * 2.2) * 0.35
  })
  return material
}

export function Volcano() {
  const lava = useLavaMaterial()
  return (
    <group scale={[1, 0.88, 1]}>
      <Cyl r={1.5} rTop={0.85} h={3} seg={8} pos={[0, 1.5, 0]} color={C.ash} />
      <Cyl r={1.1} rTop={0.7} h={0.5} seg={8} pos={[0, 3.2, 0]} color="#3b2f2a" />
      <Cyl r={0.6} h={0.08} seg={8} pos={[0, 3.42, 0]} color={C.lava} material={lava} />
      <Box size={[0.28, 0.07, 3.1]} pos={[0, 1.55, 1.1]} rot={[0.3, 0, 0]} color={C.lava} material={lava} />
      <Box size={[3.1, 0.07, 0.24]} pos={[-1.1, 1.3, 0.2]} rot={[0, 0, -0.3]} color={C.lava} material={lava} />
    </group>
  )
}

export function Tree({ props }: PrefabProps) {
  return props.variant === 'round' ? <TreeRound /> : <TreePine />
}
