import { useState } from 'react'
import { fieldCls } from './EffectsFields'
import type { CompendiumItem } from './items'
import { lootPlacement } from './lootPlacement'

/** "Esconder no mapa": set the difficulty (Sucessos), then click a cell or object on the map. */
export function HideForm({ item, onClose }: { item: CompendiumItem; onClose: () => void }) {
  const [difficulty, setDifficulty] = useState(1)
  const start = () => {
    lootPlacement.start({ itemId: item.id, name: item.name, difficulty })
    onClose()
  }
  return (
    <div role="group" aria-label={`Esconder ${item.name} no mapa`} className="flex flex-col gap-2 rounded-lg bg-zinc-800 p-2 text-sm">
      <label className="flex items-center gap-2">Dificuldade (Sucessos para achar)
        <input type="number" min={0} max={20} className={`${fieldCls} w-16`} value={difficulty} onChange={(e) => setDifficulty(Math.min(20, Math.max(0, Number(e.target.value) || 0)))} />
      </label>
      <div className="flex gap-2">
        <button type="button" onClick={start} className="rounded bg-emerald-700 px-3 py-1 hover:bg-emerald-600">Escolher no mapa</button>
        <button type="button" onClick={onClose} className="rounded bg-zinc-700 px-3 py-1 hover:bg-zinc-600">Cancelar</button>
      </div>
    </div>
  )
}
