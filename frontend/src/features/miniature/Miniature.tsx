import { useFrame } from '@react-three/fiber'
import { useRef } from 'react'
import type { Group } from 'three'
import { normalizeAppearance, type Appearance } from './appearance'
import { Body } from './Body'
import { Gear } from './Gear'
import { Head } from './Head'
import { effectiveLevel } from './assimilationMarks'
import { Marks } from './Marks'

interface Props {
  appearance: Appearance
  /** `tugOfWar.assimilationLevel` (1-10); omit for NPCs. */
  assimilationLevel?: number | null
  /** Idle breathing sway; off in static renders. */
  animate?: boolean
}

/** Low-poly character miniature, about one unit tall with the feet at y = 0. */
export function Miniature({ appearance, assimilationLevel, animate = true }: Props) {
  const a = normalizeAppearance(appearance)
  const ref = useRef<Group>(null)
  useFrame(({ clock }) => {
    if (animate && ref.current) ref.current.scale.y = a.body.height * (1 + Math.sin(clock.elapsedTime * 2) * 0.012)
  })
  const level = effectiveLevel(assimilationLevel, a.infected)
  return (
    <group ref={ref} scale={[a.body.height, a.body.height, a.body.height]} position={[0, 0.1, 0]}>
      <Body a={a} />
      <Head a={a} />
      <Gear a={a} />
      <Marks level={level} build={a.body.build} />
    </group>
  )
}
