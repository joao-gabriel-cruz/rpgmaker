import { useEffect } from 'react'
import { EMOJIS } from './presence'

const RADIUS = 58

interface Props {
  /** Screen point the wheel is centered on. */
  x: number
  y: number
  onPick: (emoji: string) => void
  onClose: () => void
}

/** Radial menu of the six quick reactions; Esc or a click outside closes it. */
export function ReactionWheel({ x, y, onPick, onClose }: Props) {
  useEffect(() => {
    const key = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    window.addEventListener('keydown', key)
    return () => window.removeEventListener('keydown', key)
  }, [onClose])
  return (
    <div className="fixed inset-0 z-40" onPointerDown={(e) => e.target === e.currentTarget && onClose()}>
      <div role="menu" aria-label="Reações" className="absolute h-0 w-0" style={{ left: x, top: y }}>
        <span aria-hidden className="absolute -left-[4.5rem] -top-[4.5rem] h-36 w-36 rounded-full bg-zinc-950/60 ring-1 ring-white/10 backdrop-blur-sm" />
        {EMOJIS.map((emoji, i) => {
          const a = (i / EMOJIS.length) * Math.PI * 2 - Math.PI / 2
          return (
            <button key={emoji} type="button" role="menuitem" aria-label={`Reagir com ${emoji}`} onClick={() => onPick(emoji)}
              className="absolute flex h-10 w-10 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full bg-zinc-800 text-xl transition-transform hover:scale-125 active:scale-95 focus-visible:outline-2 focus-visible:outline-emerald-400 motion-reduce:transition-none"
              style={{ left: Math.cos(a) * RADIUS, top: Math.sin(a) * RADIUS }}>
              {emoji}
            </button>
          )
        })}
      </div>
    </div>
  )
}
