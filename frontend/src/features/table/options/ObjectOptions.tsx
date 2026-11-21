import { RotateCw, Search } from 'lucide-react'
import { createElement, useState } from 'react'
import type { MapGrid } from '../../map3d/grid'
import { CATEGORY_LABELS, type Category, type PrefabMeta } from '../../map3d/objects/catalogData'
import { AWNING_OPTIONS, WALL_HEIGHT_OPTIONS } from '../../map3d/objects/toolProps'
import type { MapObject, ObjectPatch } from '../../map3d/objects/model'
import { nextRotation } from '../../map3d/objects/model'
import type { ObjectToolId, ToolState } from '../../map3d/tools'
import type { EditHistory } from '../../map3d/useEditHistory'
import { HistoryButtons } from './HistoryButtons'
import { CATEGORY_TINT, filterPrefabs, prefabIcon } from './prefabCatalog'
import { SelectedObject } from './SelectedObject'
import { OBJECT_MODES } from './subTools'
import { Hint, INPUT, Section, Segmented, type SegOption } from './ui'

type CategoryFilter = Category | 'all'
const FILTERS: SegOption<CategoryFilter>[] = [
  { id: 'all', label: 'Todas as categorias', text: 'Todos' },
  ...(Object.keys(CATEGORY_LABELS) as Category[]).map((c) => ({ id: c, label: CATEGORY_LABELS[c], text: CATEGORY_LABELS[c] })),
]

export interface ObjectOptionsProps {
  editor: ToolState
  setEditor: (s: ToolState) => void
  grid: MapGrid
  objects: readonly MapObject[]
  onObjectPatch: (patch: ObjectPatch) => void
  selectedId: string | null
  onSelect: (id: string | null) => void
  history: EditHistory
}

const card = (on: boolean) =>
  `group flex flex-col items-center gap-1 rounded-lg p-1.5 text-center transition-[background-color,box-shadow,transform] active:scale-[0.96] motion-reduce:transition-none focus-visible:outline-2 focus-visible:outline-emerald-400 ${
    on ? 'bg-emerald-600/25 ring-2 ring-emerald-500' : 'bg-zinc-800/70 ring-1 ring-white/5 hover:bg-zinc-800'}`

function PrefabCard({ meta: p, on, onPick }: { meta: PrefabMeta; on: boolean; onPick: () => void }) {
  return (
    <li>
      <button type="button" aria-label={p.name} aria-pressed={on} className={`${card(on)} w-full`} onClick={onPick}>
        {createElement(prefabIcon(p.id), { size: 22, strokeWidth: 1.5, 'aria-hidden': true, className: CATEGORY_TINT[p.category] })}
        <span className="text-[11px] leading-tight text-zinc-200">{p.name}</span>
        <span className="text-[10px] text-zinc-500">{p.width}×{p.depth}</span>
      </button>
    </li>
  )
}

function Palette({ editor, setEditor }: Pick<ObjectOptionsProps, 'editor' | 'setEditor'>) {
  const [query, setQuery] = useState('')
  const [cat, setCat] = useState<CategoryFilter>('all')
  const list = filterPrefabs(query, cat)
  return (
    <>
      <label className="relative block">
        <Search size={14} strokeWidth={1.5} aria-hidden className="pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 text-zinc-500" />
        <input type="search" aria-label="Buscar objeto" placeholder="Buscar objeto" value={query} onChange={(e) => setQuery(e.target.value)}
          className={`${INPUT} w-full pl-7`} />
      </label>
      <Segmented label="Categoria" value={cat} options={FILTERS} onChange={setCat} />
      <ul aria-label="Catálogo de objetos" className="grid max-h-56 grid-cols-3 gap-1.5 overflow-y-auto pr-0.5">
        {list.map((p) => <PrefabCard key={p.id} meta={p} on={editor.prefab === p.id} onPick={() => setEditor({ ...editor, prefab: p.id })} />)}
      </ul>
      {list.length === 0 && <Hint>Nenhum objeto encontrado.</Hint>}
    </>
  )
}

function PlaceControls({ editor, setEditor }: Pick<ObjectOptionsProps, 'editor' | 'setEditor'>) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <button type="button" aria-label="Girar objeto" aria-keyshortcuts="Shift+R" onClick={() => setEditor({ ...editor, rotation: nextRotation(editor.rotation) })}
        className="flex items-center gap-1.5 rounded-md bg-zinc-800 px-2 py-1 text-xs text-zinc-200 transition-[background-color,transform] hover:bg-zinc-700 active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-emerald-400 motion-reduce:transition-none">
        <RotateCw size={14} strokeWidth={1.5} aria-hidden />Girar {editor.rotation}°
        <kbd className="rounded bg-zinc-900 px-1 font-mono text-[10px] text-zinc-400">Shift+R</kbd>
      </button>
      {editor.prefab === 'market_stall' && (
        <Segmented label="Cor do toldo" value={editor.color ?? 'red'} className="w-full"
          options={AWNING_OPTIONS.map((o) => ({ id: o.id, label: o.label, text: o.short }))}
          onChange={(color) => setEditor({ ...editor, color })} />
      )}
      {editor.prefab === 'ruined_wall' && (
        <Segmented label="Altura do muro" value={String(editor.wallHeight ?? 1)} className="w-full"
          options={WALL_HEIGHT_OPTIONS.map((o) => ({ id: o.id, label: o.label, text: o.short }))}
          onChange={(h) => setEditor({ ...editor, wallHeight: h === '2' ? 2 : 1 })} />
      )}
      {editor.prefab === 'tree' && (
        <Segmented label="Tipo de árvore" value={editor.variant} className="flex-1"
          options={[{ id: 'pine', label: 'Pinheiro', text: 'Pinheiro' }, { id: 'round', label: 'Frondosa', text: 'Frondosa' }]}
          onChange={(variant) => setEditor({ ...editor, variant })} />
      )}
    </div>
  )
}

/** Objects tool: Place / Select / Remove modes; the palette only exists while placing. */
export function ObjectOptions(p: ObjectOptionsProps) {
  const { editor, setEditor } = p
  const selected = p.objects.find((o) => o.id === p.selectedId) ?? null
  return (
    <div className="flex flex-col gap-3">
      <Section title="Modo" aside={<HistoryButtons history={p.history} />}>
        <Segmented label="Modo dos objetos" value={editor.tool as ObjectToolId} options={OBJECT_MODES}
          onChange={(tool) => setEditor({ ...editor, tool })} />
      </Section>
      {editor.tool === 'object' && (
        <Section title="Paleta"><Palette editor={editor} setEditor={setEditor} /><PlaceControls editor={editor} setEditor={setEditor} /></Section>
      )}
      {editor.tool === 'select' && (
        <Section title="Objeto selecionado">
          <SelectedObject obj={selected} objects={p.objects} grid={p.grid} emit={p.onObjectPatch} onDone={() => p.onSelect(null)} />
        </Section>
      )}
      {editor.tool === 'removeObject' && <Hint>Clique em um objeto do mapa para removê-lo. Ctrl+Z desfaz.</Hint>}
    </div>
  )
}
