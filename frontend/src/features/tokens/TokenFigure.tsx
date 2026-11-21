import type { ThreeEvent } from '@react-three/fiber'
import { MiniatureToken } from './MiniatureToken'
import { Standee } from './Standee'
import type { Token } from './tokens'

interface Props {
  token: Token
  x: number
  y: number
  z: number
  dragging?: boolean
  onPointerDown?: (e: ThreeEvent<PointerEvent>) => void
}

/** The miniature when the character has an appearance; the portrait standee otherwise. */
export function TokenFigure(p: Props) {
  return p.token.appearance ? <MiniatureToken {...p} /> : <Standee {...p} />
}
