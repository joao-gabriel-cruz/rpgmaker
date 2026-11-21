import { Billboard } from '@react-three/drei'
import type { ThreeEvent } from '@react-three/fiber'
import { useEffect, useMemo } from 'react'
import { normalizeAppearance } from '../miniature/appearance'
import { Miniature } from '../miniature/Miniature'
import { nameTexture } from './portraitTexture'
import type { Token } from './tokens'

interface Props {
  token: Token
  x: number
  y: number
  z: number
  dragging?: boolean
  onPointerDown?: (e: ThreeEvent<PointerEvent>) => void
}

/** 3D miniature on a round base ringed with the player's color, name floating above. */
export function MiniatureToken({ token, x, y, z, dragging, onPointerDown }: Props) {
  const look = useMemo(() => normalizeAppearance(token.appearance), [token.appearance])
  const label = useMemo(() => nameTexture(token.name), [token.name])
  useEffect(() => () => label.dispose(), [label])
  const lift = dragging ? 0.4 : 0
  return (
    <group position={[x + 0.5, y + lift, z + 0.5]} onPointerDown={onPointerDown}>
      <mesh position={[0, 0.05, 0]}>
        <cylinderGeometry args={[0.38, 0.42, 0.1, 24]} />
        <meshStandardMaterial color={look.tokenColor} />
      </mesh>
      <mesh position={[0, 0.052, 0]}>
        <cylinderGeometry args={[0.3, 0.3, 0.1, 24]} />
        <meshStandardMaterial color="#27272a" />
      </mesh>
      <Miniature appearance={look} assimilationLevel={token.assimilationLevel} />
      <Billboard lockX lockZ position={[0, 1.35 * look.body.height, 0]}>
        <mesh>
          <planeGeometry args={[1.2, 0.225]} />
          <meshBasicMaterial map={label} transparent opacity={dragging ? 0.8 : 1} />
        </mesh>
      </Billboard>
    </group>
  )
}
