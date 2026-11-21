import { Billboard } from '@react-three/drei'
import type { ThreeEvent } from '@react-three/fiber'
import { useEffect, useMemo } from 'react'
import { nameTexture, portraitTexture } from './portraitTexture'
import { playerColor, type Token } from './tokens'

interface Props {
  token: Token
  y: number
  x: number
  z: number
  dragging?: boolean
  onPointerDown?: (e: ThreeEvent<PointerEvent>) => void
}

/** Vertical portrait plate on a round base; the plate always faces the camera around the Y axis. */
export function Standee({ token, x, z, y, dragging, onPointerDown }: Props) {
  const color = playerColor(token.ownerId)
  const portrait = useMemo(() => portraitTexture(token.name, color, token.portraitUrl), [token.name, color, token.portraitUrl])
  const label = useMemo(() => nameTexture(token.name), [token.name])
  useEffect(() => () => { portrait.dispose(); label.dispose() }, [portrait, label])
  const lift = dragging ? 0.4 : 0
  return (
    <group position={[x + 0.5, y + lift, z + 0.5]} onPointerDown={onPointerDown}>
      <mesh position={[0, 0.05, 0]}>
        <cylinderGeometry args={[0.38, 0.42, 0.1, 24]} />
        <meshStandardMaterial color={color} />
      </mesh>
      <Billboard lockX lockZ position={[0, 0.1, 0]}>
        <mesh position={[0, 0.5, 0]}>
          <planeGeometry args={[0.8, 0.9]} />
          <meshBasicMaterial map={portrait} transparent opacity={dragging ? 0.8 : 1} />
        </mesh>
        <mesh position={[0, 1.2, 0]}>
          <planeGeometry args={[1.2, 0.225]} />
          <meshBasicMaterial map={label} transparent />
        </mesh>
      </Billboard>
    </group>
  )
}
