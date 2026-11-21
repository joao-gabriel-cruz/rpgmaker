import { useThree } from '@react-three/fiber'
import { useEffect } from 'react'
import { createCameraActions, type CameraHandle } from './cameraActions'

/** In-canvas half of the camera controls: hands the DOM buttons actions bound to the orbit rig. */
export function CameraBridge({ handleRef }: { handleRef: CameraHandle }) {
  const controls = useThree((s) => s.controls) as { target: { x: number; y: number; z: number }; update: () => void } | null
  const camera = useThree((s) => s.camera)
  useEffect(() => {
    if (!controls) return
    handleRef.current = createCameraActions({ controls, camera })
    return () => { handleRef.current = null }
  }, [controls, camera, handleRef])
  return null
}
