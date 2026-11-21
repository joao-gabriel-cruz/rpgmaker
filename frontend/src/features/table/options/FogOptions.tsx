import { Cloud, CloudFog, Eye, EyeOff, Sun } from 'lucide-react'
import { useEffect, useState } from 'react'
import type { Vision } from '../../vision/useVision'
import { FOG_MODES, type FogMode } from '../../vision/vision'
import { BTN, Hint, Section, Segmented, type SegOption } from './ui'

const MODE_ICONS: Record<FogMode, SegOption<FogMode>['icon']> = { off: Sun, manual: Cloud, dynamic: CloudFog }
const SHORT: Record<FogMode, string> = { off: 'Sem', manual: 'Manual', dynamic: 'Dinâmica' }
const MODES: SegOption<FogMode>[] = FOG_MODES.map((m) => ({ id: m.id, label: m.label, icon: MODE_ICONS[m.id], text: SHORT[m.id] }))

const BRUSHES: SegOption<'reveal' | 'hide'>[] = [
  { id: 'reveal', label: 'Revelar área', icon: Eye, text: 'Revelar' },
  { id: 'hide', label: 'Esconder área', icon: EyeOff, text: 'Esconder' },
]

function Radius({ vision }: { vision: Vision }) {
  const { state, configure } = vision
  const [local, setLocal] = useState<number | null>(null)
  const shown = local ?? state.visionRadius
  return (
    <label className="flex items-center gap-2 text-sm text-zinc-200">
      <span className="shrink-0">Raio</span>
      <input type="range" aria-label="Raio de visão" min={1} max={40} value={shown} className="min-w-0 flex-1 accent-emerald-500"
        onChange={(e) => setLocal(Number(e.target.value))}
        onPointerUp={() => { if (local !== null) { configure(state.mode, local); setLocal(null) } }}
        onKeyUp={() => { if (local !== null) { configure(state.mode, local); setLocal(null) } }} />
      <output className="w-6 text-right tabular-nums text-zinc-300">{shown}</output>
    </label>
  )
}

/** Fog tool: mode, vision radius and the reveal/hide area brush. The brush starts on "Revelar". */
export function FogOptions({ vision }: { vision: Vision }) {
  const { state, brush, setBrush, configure, all } = vision
  useEffect(() => { if (brush === null) setBrush('reveal') }, [brush, setBrush])
  return (
    <div className="flex flex-col gap-3">
      <Section title="Modo">
        <Segmented label="Modo da névoa" value={state.mode} options={MODES} onChange={(m) => configure(m, state.visionRadius)} />
      </Section>
      {state.mode === 'dynamic' && <Section title="Visão"><Radius vision={vision} /></Section>}
      <Section title="Pincel de área">
        <Segmented label="Pincel da névoa" value={brush} options={BRUSHES} onChange={setBrush} />
        <Hint>Arraste um retângulo no mapa para revelar ou esconder.</Hint>
        <div className="flex gap-1">
          <button type="button" className={`${BTN} flex-1`} onClick={() => all(true)}>Revelar tudo</button>
          <button type="button" className={`${BTN} flex-1`} onClick={() => all(false)}>Esconder tudo</button>
        </div>
      </Section>
      {vision.error && <p role="alert" className="text-xs text-red-400">{vision.error}</p>}
    </div>
  )
}
