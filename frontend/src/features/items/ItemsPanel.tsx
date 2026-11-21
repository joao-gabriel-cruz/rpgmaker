import { useState } from 'react'
import { useCharacters } from '../characters/characters'
import { GiveForm } from './GiveForm'
import { HideForm } from './HideForm'
import { ItemForm } from './ItemForm'
import { ItemIcon } from './ItemIcon'
import { ItemPreview } from './ItemPreview'
import { blankItem, deleteItem, filterItems, KIND_LABEL, saveItem, useItems, type CompendiumItem, type ItemDraft, type ItemKind } from './items'
import { fieldCls } from './EffectsFields'
import { lootPlacement, usePlacement } from './lootPlacement'

type Mode = { type: 'edit'; item?: CompendiumItem } | { type: 'give' | 'hide'; item: CompendiumItem } | null

const btn = 'rounded bg-zinc-800 px-2 py-0.5 text-xs hover:bg-zinc-700'

function PlacementBanner() {
  const placing = usePlacement()
  if (!placing) return null
  return (
    <p role="status" className="flex items-center justify-between gap-2 rounded-lg bg-amber-900/70 px-3 py-2 text-sm text-amber-100">
      <span>Clique numa célula ou objeto do mapa para esconder “{placing.name}” (dificuldade {placing.difficulty}).</span>
      <button type="button" className={btn} onClick={lootPlacement.cancel}>Cancelar</button>
    </p>
  )
}

function Row({ item, onMode, onDelete }: { item: CompendiumItem; onMode: (m: Mode) => void; onDelete: () => void }) {
  return (
    <li className="flex flex-col gap-1">
      <ItemPreview item={item} />
      <div className="flex flex-wrap gap-1">
        <button type="button" className={btn} aria-label={`Entregar ${item.name}`} onClick={() => onMode({ type: 'give', item })}>Entregar…</button>
        <button type="button" className={btn} aria-label={`Esconder ${item.name} no mapa`} onClick={() => onMode({ type: 'hide', item })}>Esconder no mapa</button>
        <button type="button" className={btn} aria-label={`Editar ${item.name}`} onClick={() => onMode({ type: 'edit', item })}>Editar</button>
        <button type="button" className={btn} aria-label={`Excluir ${item.name}`} onClick={onDelete}>Excluir</button>
      </div>
    </li>
  )
}

function Filters({ query, kind, set }: { query: string; kind: ItemKind | ''; set: (q: string, k: ItemKind | '') => void }) {
  return (
    <div className="flex gap-2">
      <input type="search" aria-label="Buscar item" placeholder="Buscar…" className={`${fieldCls} min-w-0 flex-1`} value={query} onChange={(e) => set(e.target.value, kind)} />
      <select aria-label="Filtrar por tipo" className={fieldCls} value={kind} onChange={(e) => set(query, e.target.value as ItemKind | '')}>
        <option value="">Todos</option>
        {Object.entries(KIND_LABEL).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
      </select>
    </div>
  )
}

function useCompendium(campaignId: string, mode: Mode, setMode: (m: Mode) => void) {
  const { items, error, reload } = useItems(campaignId)
  const [failure, setFailure] = useState('')
  const save = async (draft: ItemDraft) => {
    await saveItem(campaignId, draft, mode?.type === 'edit' ? mode.item?.id : undefined)
    reload()
    setMode(null)
  }
  const remove = (item: CompendiumItem) => {
    setFailure('')
    deleteItem(campaignId, item.id).then(reload, (e: Error) => setFailure(e.message))
  }
  return { items, error: error || failure, save, remove }
}

const draftOf = (mode: Mode): ItemDraft => {
  if (mode?.type !== 'edit' || !mode.item) return blankItem()
  const { id: _id, ...rest } = mode.item
  return rest
}

/** The Mestre's item compendium: search, create/edit, give to a character or the Refúgio, hide on the map. */
export function ItemsPanel({ campaignId }: { campaignId: string }) {
  const [mode, setMode] = useState<Mode>(null)
  const [filter, setFilter] = useState<{ query: string; kind: ItemKind | '' }>({ query: '', kind: '' })
  const { items, error, save, remove } = useCompendium(campaignId, mode, setMode)
  const { data: chars } = useCharacters(campaignId)
  const close = () => setMode(null)
  return (
    <section aria-label="Itens" className="flex flex-col gap-3 rounded-xl bg-zinc-900 px-3 py-3">
      <h2 className="flex items-center justify-between text-lg font-semibold">
        <span className="flex items-center gap-2"><ItemIcon name="backpack" size={18} />Itens</span>
        {mode?.type !== 'edit' && <button type="button" className="rounded bg-emerald-700 px-2 py-1 text-sm hover:bg-emerald-600" onClick={() => setMode({ type: 'edit' })}>Novo item</button>}
      </h2>
      <PlacementBanner />
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      {mode?.type === 'edit' ? <ItemForm key={mode.item?.id ?? 'new'} initial={draftOf(mode)} onSave={save} onCancel={close} /> : (
        <>
          <Filters query={filter.query} kind={filter.kind} set={(query, kind) => setFilter({ query, kind })} />
          {mode?.type === 'give' && <GiveForm item={mode.item} characters={Array.isArray(chars) ? chars : []} onClose={close} />}
          {mode?.type === 'hide' && <HideForm item={mode.item} onClose={close} />}
          {items.length === 0 && <p className="text-sm text-zinc-400">Nenhum item no compêndio. Crie o primeiro.</p>}
          <ul className="flex flex-col gap-3">{filterItems(items, filter.query, filter.kind).map((i) => <Row key={i.id} item={i} onMode={setMode} onDelete={() => remove(i)} />)}</ul>
        </>
      )}
    </section>
  )
}
