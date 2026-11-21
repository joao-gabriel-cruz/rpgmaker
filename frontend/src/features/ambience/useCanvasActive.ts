import { useThree } from '@react-three/fiber'
import { useEffect, useRef } from 'react'

/** Ref that is true while the canvas is on screen and the tab visible; effects pause their animation otherwise. */
export function useCanvasActive() {
  const dom = useThree((s) => s.gl.domElement)
  const active = useRef(true)
  useEffect(() => {
    let inView = true
    const update = () => { active.current = inView && !document.hidden }
    const io = typeof IntersectionObserver === 'function'
      ? new IntersectionObserver(([e]) => { inView = e.isIntersecting; update() })
      : null
    io?.observe(dom)
    document.addEventListener('visibilitychange', update)
    return () => {
      io?.disconnect()
      document.removeEventListener('visibilitychange', update)
    }
  }, [dom])
  return active
}
