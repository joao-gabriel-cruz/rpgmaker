import type { ReactNode } from 'react'

/** CSS tooltip shown on hover/focus of the parent `group`; the button keeps its own aria-label. */
export function Tip({ children, side = 'right' }: { children: ReactNode; side?: 'right' | 'bottom' | 'left' }) {
  const pos = side === 'right' ? 'left-full top-1/2 ml-2 -translate-y-1/2'
    : side === 'left' ? 'right-full top-1/2 mr-2 -translate-y-1/2'
    : 'left-1/2 top-full mt-2 -translate-x-1/2'
  return (
    <span role="tooltip" aria-hidden
      className={`pointer-events-none absolute z-40 whitespace-nowrap rounded-md bg-zinc-950 px-2 py-1 text-xs text-zinc-100 opacity-0 shadow-lg ring-1 ring-white/10 transition-opacity duration-100 group-hover:opacity-100 group-focus-visible:opacity-100 motion-reduce:transition-none ${pos}`}>
      {children}
    </span>
  )
}

export function Kbd({ children }: { children: ReactNode }) {
  return <kbd className="ml-1.5 rounded bg-zinc-800 px-1 font-mono text-[10px] text-zinc-300">{children}</kbd>
}
