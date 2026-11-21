import { useFrame } from '@react-three/fiber'
import { useEffect, useMemo } from 'react'
import { MeshStandardMaterial } from 'three'
import { useReducedMotion } from '../../ambience/useReducedMotion'
import { buildBody, buildVein, CRATER_HEIGHT, SEGMENTS, VEIN_SEGMENTS, volcanoRings } from './volcanoGeometry'

const BASE_GLOW = 1.1

function useLavaMaterial() {
  const reduced = useReducedMotion()
  const material = useMemo(() => new MeshStandardMaterial({ color: '#e8531c', emissive: '#ff4a00', emissiveIntensity: BASE_GLOW, flatShading: true }), [])
  useEffect(() => { if (reduced) material.emissiveIntensity = BASE_GLOW }, [reduced, material])
  useFrame(({ clock }) => {
    if (!reduced) material.emissiveIntensity = BASE_GLOW + Math.sin(clock.elapsedTime * 2.2) * 0.4 + Math.sin(clock.elapsedTime * 5.3) * 0.12
  })
  return material
}

/** Wide low-poly cone filling the 3x3 footprint: irregular slope, crater with glowing lava and lava veins. */
export function Volcano() {
  const lava = useLavaMaterial()
  const rings = useMemo(volcanoRings, [])
  const body = useMemo(() => buildBody(rings), [rings])
  const veins = useMemo(() => VEIN_SEGMENTS.map((s) => buildVein(s % SEGMENTS, rings)), [rings])
  return (
    <group>
      <mesh geometry={body}><meshStandardMaterial vertexColors flatShading roughness={0.95} /></mesh>
      <mesh position={[0, CRATER_HEIGHT + 0.04, 0]} rotation={[-Math.PI / 2, 0, 0]} material={lava}><circleGeometry args={[0.46, 12]} /></mesh>
      {veins.map((g, i) => <mesh key={VEIN_SEGMENTS[i]} geometry={g} material={lava} />)}
    </group>
  )
}
