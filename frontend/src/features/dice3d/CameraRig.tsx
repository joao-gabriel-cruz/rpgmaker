import { useThree } from '@react-three/fiber'
import { useLayoutEffect } from 'react'
import { PerspectiveCamera } from 'three'
import { gridHalfExtents } from './trayLayout'

const ELEVATION = (68 * Math.PI) / 180

// Frames the dice grid from above, fitting both width and depth for the canvas aspect.
export function CameraRig({ total }: { total: number }) {
  const camera = useThree((s) => s.camera) as PerspectiveCamera
  const { width, height } = useThree((s) => s.size)
  useLayoutEffect(() => {
    const { x, z } = gridHalfExtents(total)
    const tan = Math.tan((camera.fov * Math.PI) / 360)
    const aspect = width / height
    const byDepth = (z * Math.sin(ELEVATION) + 1.5 * Math.cos(ELEVATION)) / tan
    const byWidth = x / (aspect * tan)
    const dist = Math.max(byDepth, byWidth) * 1.15 + z * Math.cos(ELEVATION)
    camera.position.set(0, dist * Math.sin(ELEVATION), dist * Math.cos(ELEVATION))
    camera.lookAt(0, 0, 0)
    camera.updateProjectionMatrix()
  }, [camera, total, width, height])
  return null
}
