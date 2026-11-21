import { ChevronDown, Clapperboard } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { SceneSelector } from '../scenes/SceneSelector'
import type { SceneControls } from '../scenes/useSceneControls'
import { SURFACE } from './surface'

function useDismiss(open: boolean, close: () => void) {
  const box = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDown = (e: MouseEvent) => !box.current?.contains(e.target as Node) && close()
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && close()
    document.addEventListener('mousedown', onDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [open, close])
  return box
}

/** Top bar scene control: the Mestre gets a dropdown (create/rename/activate/delete); players see the active scene. */
export function SceneMenu({ scenes, isGm }: { scenes: SceneControls; isGm: boolean }) {
  const [open, setOpen] = useState(false)
  const box = useDismiss(open, () => setOpen(false))
  const name = scenes.selected?.name ?? 'Sem cena'
  if (!isGm) {
    return <span title="Cena atual" className="flex items-center gap-1.5 text-sm text-zinc-300"><Clapperboard size={16} aria-hidden />{scenes.loading ? 'Carregando…' : scenes.selected?.name ?? 'Nenhuma cena ativa'}</span>
  }
  return (
    <div ref={box} className="relative">
      <button type="button" aria-haspopup="dialog" aria-expanded={open} aria-label="Cenas" onClick={() => setOpen((o) => !o)}
        className="flex h-8 items-center gap-1.5 rounded-lg px-2 text-sm text-zinc-200 transition-[background-color,transform] hover:bg-zinc-800 active:scale-[0.96]">
        <Clapperboard size={16} aria-hidden />{name}<ChevronDown size={14} aria-hidden />
      </button>
      {open && (
        <div role="dialog" aria-label="Seletor de cenas" className={`absolute left-0 top-full z-40 mt-2 w-[26rem] max-w-[90vw] rounded-xl p-3 ${SURFACE} bg-zinc-900/95`}>
          <SceneSelector scenes={scenes.scenes} selectedId={scenes.selectedId} onSelect={(id) => { scenes.select(id); setOpen(false) }}
            onCreate={scenes.create} onRename={scenes.rename} onDelete={scenes.remove} onActivate={scenes.activate} />
          {scenes.error && <p role="alert" className="mt-2 text-sm text-red-400">{scenes.error}</p>}
        </div>
      )}
    </div>
  )
}
