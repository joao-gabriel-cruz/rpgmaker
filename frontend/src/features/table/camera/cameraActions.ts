import type { Vec3 } from './cameraMath'
import { isTopDown, rotateOffset, topDownOffset, zoomOffset, ZOOM_STEP } from './cameraMath'

/** What the camera buttons and shortcuts can do; the in-canvas bridge fills in the implementation. */
export interface CameraActions {
  zoomIn: () => void
  zoomOut: () => void
  rotate: (degrees: number) => void
  toggleTopDown: () => boolean
}

export type CameraHandle = { current: CameraActions | null }

interface Orbit {
  target: Vec3
  update: () => void
}

interface Rig {
  controls: Orbit
  camera: { position: Vec3 & { set: (x: number, y: number, z: number) => void } }
}

const offsetOf = ({ controls, camera }: Rig): Vec3 => ({
  x: camera.position.x - controls.target.x,
  y: camera.position.y - controls.target.y,
  z: camera.position.z - controls.target.z,
})

function place(rig: Rig, o: Vec3) {
  const t = rig.controls.target
  rig.camera.position.set(t.x + o.x, t.y + o.y, t.z + o.z)
  rig.controls.update()
}

/** Camera actions over an OrbitControls rig; `saved` remembers the angle to return to from top-down. */
export function createCameraActions(rig: Rig): CameraActions {
  let saved: Vec3 | null = null
  const apply = (f: (o: Vec3) => Vec3) => place(rig, f(offsetOf(rig)))
  return {
    zoomIn: () => apply((o) => zoomOffset(o, ZOOM_STEP)),
    zoomOut: () => apply((o) => zoomOffset(o, 1 / ZOOM_STEP)),
    rotate: (deg) => apply((o) => rotateOffset(o, deg)),
    toggleTopDown: () => {
      const cur = offsetOf(rig)
      const back = isTopDown(cur) && saved
      place(rig, back ? saved! : topDownOffset(cur))
      saved = back ? null : cur
      return !back
    },
  }
}
