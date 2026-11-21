import { OrbitControls } from '@react-three/drei'
import { Canvas } from '@react-three/fiber'
import type { Appearance } from './appearance'
import { Miniature } from './Miniature'

/** Large rotating 3D preview used by the appearance editor. */
export default function MiniaturePreview({ appearance, assimilationLevel }: { appearance: Appearance; assimilationLevel?: number | null }) {
  return (
    <Canvas camera={{ position: [0, 0.8, 2.9], fov: 35 }} dpr={[1, 2]}>
      <color attach="background" args={['#18181b']} />
      <ambientLight intensity={0.9} />
      <directionalLight position={[3, 5, 4]} intensity={1.8} />
      <mesh position={[0, 0.05, 0]}>
        <cylinderGeometry args={[0.42, 0.45, 0.1, 24]} />
        <meshStandardMaterial color={appearance.tokenColor} />
      </mesh>
      <Miniature appearance={appearance} assimilationLevel={assimilationLevel} />
      <OrbitControls target={[0, 0.65, 0]} autoRotate autoRotateSpeed={3} enablePan={false} minDistance={1.4} maxDistance={4} />
    </Canvas>
  )
}
