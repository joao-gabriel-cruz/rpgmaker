import { Billboard } from '@react-three/drei'
import type { ThreeEvent } from '@react-three/fiber'
import { useFrame } from '@react-three/fiber'
import { useEffect, useMemo, useRef } from 'react'
import type { Group } from 'three'
import { emoteFrame } from '../fx/emoteState'
import { createTint } from '../fx/figureTint'
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

// Applies the running emote (if any) to the figure and its arms every frame.
function useEmotePose(token: Token) {
  const figure = useRef<Group>(null)
  const armL = useRef<Group>(null)
  const armR = useRef<Group>(null)
  const tint = useRef<ReturnType<typeof createTint> | null>(null)
  useEffect(() => {
    const t = figure.current && createTint(figure.current)
    tint.current = t
    return () => { t?.dispose(); tint.current = null }
  }, [])
  useFrame(() => {
    const g = figure.current
    if (!g) return
    const { pose: p, turn, active } = emoteFrame(token.id)
    g.position.set(Math.sin(turn) * p.dz, p.dy, Math.cos(turn) * p.dz)
    g.rotation.set(p.lean, turn + p.twist, p.tilt, 'YXZ')
    g.scale.y = p.squash
    armL.current?.rotation.set(p.armL[0], 0, p.armL[1])
    armR.current?.rotation.set(p.armR[0], 0, p.armR[1])
    if (active || g.userData.wasActive) tint.current?.apply(p.opacity, p.flash)
    g.userData.wasActive = active
  })
  return { figure, rig: { armL, armR } }
}

/** 3D miniature on a round base ringed with the player's color, name floating above. */
export function MiniatureToken({ token, x, y, z, dragging, onPointerDown }: Props) {
  const look = useMemo(() => normalizeAppearance(token.appearance), [token.appearance])
  const label = useMemo(() => nameTexture(token.name), [token.name])
  useEffect(() => () => label.dispose(), [label])
  const lift = dragging ? 0.4 : 0
  const { figure, rig } = useEmotePose(token)
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
      <group ref={figure}><Miniature appearance={look} assimilationLevel={token.assimilationLevel} rig={rig} /></group>
      <Billboard lockX lockZ position={[0, 1.35 * look.body.height, 0]}>
        <mesh>
          <planeGeometry args={[1.2, 0.225]} />
          <meshBasicMaterial map={label} transparent opacity={dragging ? 0.8 : 1} />
        </mesh>
      </Billboard>
    </group>
  )
}
