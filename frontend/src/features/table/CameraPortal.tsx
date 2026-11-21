import type { ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { useTableTool } from './useTableTool'

/** Renders the camera controls in the bottom-right zone of the table; free-floating when there is no zone. */
export function CameraPortal({ children }: { children: ReactNode }) {
  const { cameraHost } = useTableTool()
  if (cameraHost) return createPortal(children, cameraHost)
  return <div className="absolute bottom-3 right-3 z-10">{children}</div>
}
