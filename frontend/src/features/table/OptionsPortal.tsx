import type { ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { useTableTool } from './useTableTool'

/** Renders children inside the floating options panel (the map owns the state those controls drive). */
export function OptionsPortal({ children }: { children: ReactNode }) {
  const { optionsHost } = useTableTool()
  return optionsHost ? createPortal(children, optionsHost) : null
}
