import { useEffect } from 'react'

/** Remembers the element focused when the panel opened and gives focus back when it closes. */
export function useReturnFocus() {
  useEffect(() => {
    const prev = document.activeElement
    return () => {
      if (prev instanceof HTMLElement && prev.isConnected) prev.focus()
    }
  }, [])
}
