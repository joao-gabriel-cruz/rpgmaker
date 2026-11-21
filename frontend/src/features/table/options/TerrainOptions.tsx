import { Grid3x3 } from 'lucide-react'
import { MAX_HEIGHT, MIN_HEIGHT } from '../../map3d/grid'
import { TERRAIN_LIST } from '../../map3d/terrainTypes'
import { BRUSH_SIZES, type ToolState } from '../../map3d/tools'
import type { EditHistory } from '../../map3d/useEditHistory'
import { HistoryButtons } from './HistoryButtons'
import { Tip } from '../Tip'
import { TERRAIN_SUBTOOLS } from './subTools'
import { Section, Segmented, Stepper, Switch } from './ui'

interface Props {
  editor: ToolState
  setEditor: (s: ToolState) => void
  showGrid: boolean
  toggleGrid: () => void
  history: EditHistory
}

function Swatches({ editor, setEditor }: Pick<Props, 'editor' | 'setEditor'>) {
  return (
    <div role="group" aria-label="Terreno" className="grid grid-cols-5 gap-1.5">
      {TERRAIN_LIST.map((t) => {
        const on = editor.terrain === t.id
        return (
          <button key={t.id} type="button" aria-label={t.name} aria-pressed={on} onClick={() => setEditor({ ...editor, terrain: t.id })}
            style={{ background: t.color }}
            className={`group relative h-9 rounded-md shadow-sm shadow-black/40 transition-[box-shadow,transform] active:scale-[0.96] motion-reduce:transition-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-400 ${
              on ? 'ring-2 ring-white ring-offset-2 ring-offset-zinc-900' : 'ring-1 ring-black/40 hover:ring-white/60'}`}>
            <Tip side="bottom">{t.name}</Tip>
          </button>
        )
      })}
    </div>
  )
}

/** Terrain tool: sub-tools, terrain swatches and brush; only what the chosen sub-tool uses is shown. */
export function TerrainOptions({ editor, setEditor, showGrid, toggleGrid, history }: Props) {
  const t = editor.tool
  const set = (p: Partial<ToolState>) => setEditor({ ...editor, ...p })
  return (
    <div className="flex flex-col gap-3">
      <Section title="Ferramenta" aside={<HistoryButtons history={history} />}>
        <Segmented label="Ferramenta de terreno" value={t} options={TERRAIN_SUBTOOLS} onChange={(tool) => set({ tool })} />
      </Section>
      {(t === 'paint' || t === 'fill') && <Section title="Terreno"><Swatches editor={editor} setEditor={setEditor} /></Section>}
      {t !== 'fill' && (
        <Section title="Pincel">
          <Segmented label="Tamanho do pincel" value={editor.brushSize}
            options={BRUSH_SIZES.map((n) => ({ id: n, label: `Pincel ${n}×${n}`, text: `${n}×${n}` }))}
            onChange={(brushSize) => set({ brushSize })} />
        </Section>
      )}
      {t === 'height' && (
        <Section title={`Altura (${MIN_HEIGHT}–${MAX_HEIGHT})`}>
          <Stepper label="Altura" value={editor.height} min={MIN_HEIGHT} max={MAX_HEIGHT} onChange={(height) => set({ height })} />
        </Section>
      )}
      <Switch label="Mostrar grade" icon={Grid3x3} on={showGrid} onChange={toggleGrid} />
    </div>
  )
}
