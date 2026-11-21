import { ExternalLink, X } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { Link } from 'react-router'
import type { Campaign } from '../campaigns/campaigns'
import { CharacterEditor } from '../characters/CharacterEditor'
import { SURFACE } from '../table/surface'
import { useReturnFocus } from '../table/useReturnFocus'
import { useTableCharacter } from './TableCharacterContext'

function Sheet({ campaign, id, onClose }: { campaign: Campaign; id: string; onClose: () => void }) {
  const box = useRef<HTMLElement>(null)
  useReturnFocus()
  useEffect(() => {
    box.current?.focus()
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose, id])
  const page = (
    <Link to={`/campaigns/${campaign.id}/characters/${id}`} className="flex items-center gap-1 text-sm text-zinc-400 underline hover:text-white">
      <ExternalLink size={14} aria-hidden /> Abrir em página
    </Link>
  )
  return (
    <aside ref={box} tabIndex={-1} aria-label="Ficha do personagem"
      className={`pointer-events-auto flex h-full w-[36rem] max-w-full flex-col rounded-2xl p-1 outline-none ${SURFACE}`}>
      <header className="flex items-center justify-between px-3 py-1.5">
        <h2 className="text-sm font-semibold">Ficha</h2>
        <button type="button" aria-label="Fechar ficha" onClick={onClose}
          className="rounded-lg p-1.5 text-zinc-400 transition-[background-color,color,transform] hover:bg-zinc-800 hover:text-white active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-emerald-400">
          <X size={18} aria-hidden />
        </button>
      </header>
      <div className="min-h-0 flex-1 overflow-y-auto rounded-xl p-3">
        <CharacterEditor key={id} campaign={campaign} characterId={id} links={page} onGone={onClose} />
      </div>
    </aside>
  )
}

/** Side drawer over the map with the full character sheet; the map stays usable beside it. */
export function SheetDrawer({ campaign }: { campaign: Campaign }) {
  const { sheetId, closeSheet } = useTableCharacter()
  return sheetId ? <Sheet campaign={campaign} id={sheetId} onClose={closeSheet} /> : null
}
