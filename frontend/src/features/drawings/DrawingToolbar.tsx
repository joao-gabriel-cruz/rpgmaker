import { DRAW_COLORS, DRAW_WIDTHS, type DrawMode } from './drawings'
import type { SceneDrawings } from './useDrawings'

const btn = (active: boolean) =>
  `rounded px-2 py-1 text-sm transition ${active ? 'bg-emerald-600 text-white' : 'bg-zinc-800 text-zinc-200 hover:bg-zinc-700'}`

const MODES: { id: Exclude<DrawMode, 'none'>; label: string }[] = [
  { id: 'ruler', label: 'Régua' },
  { id: 'draw', label: 'Desenhar' },
]

function DrawOptions({ draw }: { draw: SceneDrawings }) {
  return (
    <>
      <span className="flex items-center gap-1" role="group" aria-label="Cor do desenho">
        {DRAW_COLORS.map((c) => (
          <button
            key={c}
            aria-label={`Cor ${c}`}
            aria-pressed={draw.color === c}
            onClick={() => draw.setColor(c)}
            className={`h-5 w-5 rounded-full border-2 ${draw.color === c ? 'border-white' : 'border-transparent'}`}
            style={{ background: c }}
          />
        ))}
        <input type="color" aria-label="Cor personalizada" value={draw.color} onChange={(e) => draw.setColor(e.target.value)}
          className="h-6 w-8 cursor-pointer rounded bg-zinc-800" />
      </span>
      <label className="flex items-center gap-1 text-sm text-zinc-300">
        Espessura
        <select value={draw.width} onChange={(e) => draw.setWidth(Number(e.target.value))} className="rounded bg-zinc-800 px-1 py-1">
          {DRAW_WIDTHS.map((w) => <option key={w} value={w}>{w}</option>)}
        </select>
      </label>
    </>
  )
}

/** Ruler / draw tools for every member; the Mestre also gets "Limpar desenhos". */
export function DrawingToolbar({ draw, isGm }: { draw: SceneDrawings; isGm: boolean }) {
  return (
    <div className="flex flex-wrap items-center gap-2" role="toolbar" aria-label="Régua e desenho">
      {MODES.map((m) => (
        <button key={m.id} className={btn(draw.mode === m.id)} aria-pressed={draw.mode === m.id}
          onClick={() => draw.setMode(draw.mode === m.id ? 'none' : m.id)}>
          {m.label}
        </button>
      ))}
      {draw.mode === 'draw' && <DrawOptions draw={draw} />}
      <button className={btn(false)} disabled={draw.mineCount === 0} onClick={draw.removeMine}>Apagar meus desenhos</button>
      {isGm && <button className={btn(false)} disabled={draw.drawings.length === 0} onClick={draw.clear}>Limpar desenhos</button>}
      {draw.error && <span role="alert" className="text-sm text-red-400">{draw.error}</span>}
    </div>
  )
}
