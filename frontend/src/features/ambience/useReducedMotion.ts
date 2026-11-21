import { useEffect, useState } from 'react'

const QUERY = '(prefers-reduced-motion: reduce)'
const read = () => typeof window !== 'undefined' && typeof window.matchMedia === 'function' && window.matchMedia(QUERY).matches

/** Whether the user asked the system for less motion (live). */
export function useReducedMotion(): boolean {
  const [reduced, setReduced] = useState(read)
  useEffect(() => {
    if (typeof window.matchMedia !== 'function') return
    const mq = window.matchMedia(QUERY)
    const onChange = () => setReduced(mq.matches)
    mq.addEventListener?.('change', onChange)
    return () => mq.removeEventListener?.('change', onChange)
  }, [])
  return reduced
}
