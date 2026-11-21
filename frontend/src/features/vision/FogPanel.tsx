import { useState } from 'react'
import { FOG_MODES, type FogMode } from './vision'
import type { Vision } from './useVision'

const btn = 'rounded bg-zinc-800 px-2 py-1 text-sm hover:bg-zinc-700'
const on = 'rounded bg-emerald-700 px-2 py-1 text-sm'

/** Mestre controls: fog mode, vision radius and the reveal/hide brush. */
export function FogPanel({ vision }: { vision: Vision }) {
  const { state, brush, setBrush, configure, all } = vision
  const [radius, setRadius] = useState<number | null>(null)
  const shown = radius ?? state.visionRadius
  return (
    <div className="flex flex-wrap items-center gap-2 text-sm">
      <label className="flex items-center gap-1">Névoa
        <select className="rounded bg-zinc-800 px-1 py-1" value={state.mode} onChange={(e) => configure(e.target.value as FogMode, shown)}>
          {FOG_MODES.map((m) => <option key={m.id} value={m.id}>{m.label}</option>)}
        </select>
      </label>
      <label className="flex items-center gap-1">Raio de visão
        <input type="number" min={1} max={40} className="w-16 rounded bg-zinc-800 px-1 py-1" value={shown}
          onChange={(e) => setRadius(Number(e.target.value))}
          onBlur={() => { if (radius !== null) { configure(state.mode, radius); setRadius(null) } }} />
      </label>
      <button type="button" className={brush === 'reveal' ? on : btn} aria-pressed={brush === 'reveal'} onClick={() => setBrush(brush === 'reveal' ? null : 'reveal')}>Revelar área</button>
      <button type="button" className={brush === 'hide' ? on : btn} aria-pressed={brush === 'hide'} onClick={() => setBrush(brush === 'hide' ? null : 'hide')}>Esconder área</button>
      <button type="button" className={btn} onClick={() => all(true)}>Revelar tudo</button>
      <button type="button" className={btn} onClick={() => all(false)}>Esconder tudo</button>
      {vision.error && <span role="alert" className="text-red-400">{vision.error}</span>}
    </div>
  )
}
