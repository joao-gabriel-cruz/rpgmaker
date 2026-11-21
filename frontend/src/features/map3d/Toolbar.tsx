import type { ReactNode } from 'react'
import { MAX_HEIGHT, MIN_HEIGHT } from './grid'
import { TERRAIN_LIST } from './terrainTypes'
import { PrefabPalette } from './objects/PrefabPalette'
import { BRUSH_SIZES, OBJECT_TOOLS, TERRAIN_TOOLS, type ToolState } from './tools'

interface Props {
  state: ToolState
  onChange: (next: ToolState) => void
  showGrid: boolean
  onToggleGrid: () => void
}

const btn = (active: boolean) =>
  `rounded px-2 py-1 text-sm transition ${
    active ? 'bg-emerald-600 text-white' : 'bg-zinc-800 text-zinc-200 hover:bg-zinc-700'
  }`

type SetFn = (p: Partial<ToolState>) => void

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="flex flex-col gap-1">
      <h3 className="text-xs uppercase text-zinc-400">{title}</h3>
      {children}
    </section>
  )
}

function ToolButtons({ state, set, tools }: { state: ToolState; set: SetFn; tools: typeof TERRAIN_TOOLS | typeof OBJECT_TOOLS }) {
  return tools.map((t) => (
    <button key={t.id} className={btn(state.tool === t.id)} onClick={() => set({ tool: t.id })}>
      {t.label} <kbd className="float-right text-xs opacity-60">{t.key.toUpperCase()}</kbd>
    </button>
  ))
}

function TerrainButtons({ state, set }: { state: ToolState; set: SetFn }) {
  return (
    <div className="grid grid-cols-2 gap-1">
      {TERRAIN_LIST.map((t) => (
        <button key={t.id} className={btn(state.terrain === t.id)} onClick={() => set({ terrain: t.id })}>
          <span className="mr-1 inline-block h-3 w-3 rounded-sm align-middle" style={{ background: t.color }} />
          {t.name}
        </button>
      ))}
    </div>
  )
}

function BrushButtons({ state, set }: { state: ToolState; set: SetFn }) {
  return (
    <div className="flex gap-1">
      {BRUSH_SIZES.map((n) => (
        <button key={n} className={btn(state.brushSize === n)} onClick={() => set({ brushSize: n })}>
          {n}
        </button>
      ))}
    </div>
  )
}

function HeightInput({ state, set }: { state: ToolState; set: SetFn }) {
  return (
    <label className="flex flex-col gap-1 text-sm">
      <span className="text-xs uppercase text-zinc-400">Altura ({MIN_HEIGHT}–{MAX_HEIGHT})</span>
      <input
        type="number"
        min={MIN_HEIGHT}
        max={MAX_HEIGHT}
        value={state.height}
        onChange={(e) => set({ height: Number(e.target.value) })}
        className="rounded bg-zinc-800 px-2 py-1"
      />
    </label>
  )
}

export function Toolbar({ state, onChange, showGrid, onToggleGrid }: Props) {
  const set: SetFn = (p) => onChange({ ...state, ...p })
  return (
    <div className="flex w-52 flex-col gap-4 overflow-y-auto bg-zinc-900 p-3 text-zinc-100">
      <Section title="Ferramentas"><ToolButtons state={state} set={set} tools={TERRAIN_TOOLS} /></Section>
      <Section title="Objetos"><ToolButtons state={state} set={set} tools={OBJECT_TOOLS} /></Section>
      {state.tool === 'object' && (
        <Section title="Paleta (R gira)">
          <PrefabPalette
            value={state.prefab}
            variant={state.variant}
            onChange={(prefab) => set({ prefab })}
            onVariant={(variant) => set({ variant })}
            color={state.color}
            onColor={(color) => set({ color })}
            wallHeight={state.wallHeight}
            onWallHeight={(wallHeight) => set({ wallHeight })}
          />
        </Section>
      )}
      <Section title="Terreno"><TerrainButtons state={state} set={set} /></Section>
      <Section title="Tamanho do pincel"><BrushButtons state={state} set={set} /></Section>
      <HeightInput state={state} set={set} />
      <label className="flex items-center gap-2 text-sm">
        <input type="checkbox" checked={showGrid} onChange={onToggleGrid} /> Mostrar grade
      </label>
    </div>
  )
}
