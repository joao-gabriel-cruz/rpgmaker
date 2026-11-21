import { Redo2, Undo2 } from 'lucide-react'
import type { EditHistory } from '../../map3d/useEditHistory'
import { Kbd, Tip } from '../Tip'

const B = 'group relative flex h-7 w-7 items-center justify-center rounded-md text-zinc-300 transition-[background-color,color,transform] hover:bg-zinc-800 hover:text-white active:scale-[0.96] disabled:opacity-30 disabled:hover:bg-transparent focus-visible:outline-2 focus-visible:outline-emerald-400 motion-reduce:transition-none'

/** Undo / redo of the Mestre's edits in this session (Ctrl/Cmd+Z and Shift+Ctrl/Cmd+Z). */
export function HistoryButtons({ history }: { history: EditHistory }) {
  return (
    <div className="flex gap-0.5">
      <button type="button" aria-label="Desfazer" disabled={!history.canUndo} onClick={history.undo} className={B}>
        <Undo2 size={15} strokeWidth={1.5} aria-hidden /><Tip side="bottom">Desfazer<Kbd>Ctrl+Z</Kbd></Tip>
      </button>
      <button type="button" aria-label="Refazer" disabled={!history.canRedo} onClick={history.redo} className={B}>
        <Redo2 size={15} strokeWidth={1.5} aria-hidden /><Tip side="bottom">Refazer<Kbd>Shift+Ctrl+Z</Kbd></Tip>
      </button>
    </div>
  )
}
