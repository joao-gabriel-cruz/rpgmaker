import { useCallback, useEffect, useRef } from 'react'

export const HOLD_MS = 450
const SLOP_PX = 8

export interface HoldPoint {
  x: number
  y: number
}

/** Long press: `begin` arms a timer that fires `onHold` unless the pointer is released or moves away. */
export function useHold<T>(onHold: ((item: T, at: HoldPoint) => void) | undefined) {
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const origin = useRef<HoldPoint>({ x: 0, y: 0 })
  const cancel = useCallback(() => clearTimeout(timer.current), [])
  useEffect(() => {
    const move = (e: PointerEvent) => {
      if (Math.hypot(e.clientX - origin.current.x, e.clientY - origin.current.y) > SLOP_PX) cancel()
    }
    window.addEventListener('pointerup', cancel)
    window.addEventListener('pointermove', move)
    return () => {
      cancel()
      window.removeEventListener('pointerup', cancel)
      window.removeEventListener('pointermove', move)
    }
  }, [cancel])
  return useCallback((item: T, at: HoldPoint) => {
    if (!onHold) return
    cancel()
    origin.current = at
    timer.current = setTimeout(() => onHold(item, at), HOLD_MS)
  }, [onHold, cancel])
}
