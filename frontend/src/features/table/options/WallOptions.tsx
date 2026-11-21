import { MousePointer2, Undo2 } from 'lucide-react'
import type { EditHistory } from '../../map3d/useEditHistory'
import { BTN, Hint } from './ui'

/** Wall tool: nothing to configure, just how it works and a quick way to take back the last stroke. */
export function WallOptions({ history }: { history: EditHistory }) {
  const canUndo = history.canUndo && history.lastKind === 'objects'
  return (
    <div className="flex flex-col gap-3">
      <p className="flex items-start gap-2 text-sm text-zinc-200">
        <MousePointer2 size={16} strokeWidth={1.5} aria-hidden className="mt-0.5 shrink-0 text-zinc-400" />
        Clique e arraste no mapa para erguer um muro em linha reta.
      </p>
      <Hint>Solte para confirmar. Ctrl+Z também desfaz o último traço.</Hint>
      <button type="button" disabled={!canUndo} onClick={history.undo} className={`${BTN} flex items-center justify-center gap-1.5 py-1.5`}>
        <Undo2 size={14} strokeWidth={1.5} aria-hidden />Desfazer último traço
      </button>
    </div>
  )
}
