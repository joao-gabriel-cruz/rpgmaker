import { X } from 'lucide-react'
import { useEffect, useRef, type ReactNode } from 'react'

interface Props {
  title: string
  onClose: () => void
  children: ReactNode
  wide?: boolean
}

/** Minimal accessible dialog: Esc and backdrop click close it; focus moves into it on open. */
export function Modal({ title, onClose, children, wide }: Props) {
  const box = useRef<HTMLDivElement>(null)
  useEffect(() => {
    box.current?.focus()
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return
      e.stopPropagation()
      onClose()
    }
    window.addEventListener('keydown', onKey, true)
    return () => window.removeEventListener('keydown', onKey, true)
  }, [onClose])
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div ref={box} role="dialog" aria-modal="true" aria-label={title} tabIndex={-1}
        className={`flex max-h-[85dvh] w-full flex-col rounded-2xl bg-zinc-900 p-2 shadow-2xl ring-1 ring-white/10 outline-none ${wide ? 'max-w-3xl' : 'max-w-md'}`}>
        <header className="flex items-center justify-between px-3 py-2">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button type="button" aria-label="Fechar" onClick={onClose}
            className="rounded-lg p-1.5 text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-white active:scale-[0.96]">
            <X size={18} aria-hidden />
          </button>
        </header>
        <div className="min-h-0 overflow-y-auto rounded-xl p-3">{children}</div>
      </div>
    </div>
  )
}
