import { useFrame, useThree } from '@react-three/fiber'
import { useEffect, useRef } from 'react'
import { Vector3 } from 'three'
import type { CameraTarget } from './tokens'

const DURATION = 0.7

interface Orbit {
  target: Vector3
  update: () => void
}

interface Props {
  target: CameraTarget | null
  view: { current: (() => { x: number; z: number }) | null }
}

/** Slides the orbit target (and the camera with it) to the requested point; exposes the current target. */
export function CameraRig({ target, view }: Props) {
  const controls = useThree((s) => s.controls) as Orbit | null
  const camera = useThree((s) => s.camera)
  const anim = useRef<{ delta: Vector3; done: number } | null>(null)

  useEffect(() => {
    view.current = () => ({ x: controls?.target.x ?? 0, z: controls?.target.z ?? 0 })
    return () => { view.current = null }
  }, [controls, view])

  useEffect(() => {
    if (!target || !controls) return
    anim.current = { delta: new Vector3(target.x - controls.target.x, 0, target.z - controls.target.z), done: 0 }
  }, [target, controls])

  useFrame((_, dt) => {
    const a = anim.current
    if (!a || !controls) return
    const step = Math.min(DURATION - a.done, dt)
    const k = step / DURATION
    const move = a.delta.clone().multiplyScalar(k)
    controls.target.add(move)
    camera.position.add(move)
    controls.update()
    a.done += step
    if (a.done >= DURATION) anim.current = null
  })
  return null
}
