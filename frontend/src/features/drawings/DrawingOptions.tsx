import { Eraser, Trash2 } from 'lucide-react'
import { DRAW_COLORS, DRAW_WIDTHS } from './drawings'
import type { SceneDrawings } from './useDrawings'

const btn = 'flex items-center gap-1.5 rounded-lg bg-zinc-800 px-2.5 py-1.5 text-sm text-zinc-200 transition-colors hover:bg-zinc-700 active:scale-[0.96] disabled:opacity-40 disabled:active:scale-100'

function Colors({ draw }: { draw: SceneDrawings }) {
  return (
    <div className="flex flex-wrap items-center gap-1.5" role="group" aria-label="Cor do desenho">
      {DRAW_COLORS.map((c) => (
        <button key={c} aria-label={`Cor ${c}`} aria-pressed={draw.color === c} onClick={() => draw.setColor(c)}
          className={`h-6 w-6 rounded-full border-2 transition-transform active:scale-[0.9] ${draw.color === c ? 'border-white' : 'border-transparent'}`}
          style={{ background: c }} />
      ))}
      <input type="color" aria-label="Cor personalizada" value={draw.color} onChange={(e) => draw.setColor(e.target.value)}
        className="h-6 w-8 cursor-pointer rounded bg-zinc-800" />
    </div>
  )
}

/** Options of the Draw tool: colour, thickness, erase mine, and (Mestre) clear all. */
export function DrawingOptions({ draw, isGm }: { draw: SceneDrawings; isGm: boolean }) {
  return (
    <div className="flex flex-col gap-3" role="group" aria-label="Opções do desenho">
      <Colors draw={draw} />
      <label className="flex items-center gap-2 text-sm text-zinc-300">
        Espessura
        <select value={draw.width} onChange={(e) => draw.setWidth(Number(e.target.value))}
          className="rounded-lg bg-zinc-800 px-2 py-1">
          {DRAW_WIDTHS.map((w) => <option key={w} value={w}>{w}</option>)}
        </select>
      </label>
      <div className="flex flex-wrap gap-2">
        <button className={btn} disabled={draw.mineCount === 0} onClick={draw.removeMine}>
          <Eraser size={16} aria-hidden /> Apagar meus desenhos
        </button>
        {isGm && (
          <button className={btn} disabled={draw.drawings.length === 0} onClick={draw.clear}>
            <Trash2 size={16} aria-hidden /> Limpar desenhos
          </button>
        )}
      </div>
      {draw.error && <p role="alert" className="text-sm text-red-400">{draw.error}</p>}
    </div>
  )
}
