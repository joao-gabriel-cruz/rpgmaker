import { useState } from 'react'
import { EffectsFields, fieldCls } from './EffectsFields'
import { IconPicker } from './IconPicker'
import { ItemPreview } from './ItemPreview'
import { isWeapon, KIND_LABEL, type ItemDraft, type ItemKind } from './items'

interface Props {
  initial: ItemDraft
  onSave: (draft: ItemDraft) => Promise<void> | void
  onCancel: () => void
}

const STATS = [['quality', 'Qualidade'], ['scarcity', 'Escassez'], ['pressure', 'Pressão']] as const

function Stats({ draft, onChange }: { draft: ItemDraft; onChange: (d: ItemDraft) => void }) {
  return (
    <div className="flex gap-2">
      {STATS.map(([k, label]) => (
        <label key={k} className="flex flex-col gap-1 text-xs text-zinc-400">{label}
          <input type="number" min={0} max={10} className={`${fieldCls} w-16`} value={draft[k]}
            onChange={(e) => onChange({ ...draft, [k]: Math.min(10, Math.max(0, Number(e.target.value) || 0)) })} />
        </label>
      ))}
    </div>
  )
}

function Identity({ draft, onChange }: { draft: ItemDraft; onChange: (d: ItemDraft) => void }) {
  const setKind = (kind: ItemKind) => onChange({ ...draft, kind, attackAnimation: isWeapon(kind) ? draft.attackAnimation : undefined })
  return (
    <>
      <label className="flex flex-col gap-1 text-xs text-zinc-400">Nome
        <input className={fieldCls} value={draft.name} maxLength={80} onChange={(e) => onChange({ ...draft, name: e.target.value })} />
      </label>
      <label className="flex flex-col gap-1 text-xs text-zinc-400">Descrição
        <textarea rows={2} className={fieldCls} value={draft.description} maxLength={500} onChange={(e) => onChange({ ...draft, description: e.target.value })} />
      </label>
      <div className="flex gap-2">
        <label className="flex flex-1 flex-col gap-1 text-xs text-zinc-400">Tipo
          <select className={fieldCls} value={draft.kind} onChange={(e) => setKind(e.target.value as ItemKind)}>
            {Object.entries(KIND_LABEL).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-xs text-zinc-400">Cor
          <input type="color" aria-label="Cor do item" className="h-8 w-12 rounded bg-zinc-900" value={draft.color} onChange={(e) => onChange({ ...draft, color: e.target.value })} />
        </label>
      </div>
    </>
  )
}

function useSubmit(current: ItemDraft, onSave: Props['onSave']) {
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setError('')
    try {
      await onSave(current)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Não foi possível salvar o item')
    } finally {
      setBusy(false)
    }
  }
  return { error, busy, submit }
}

/** Create/edit form of a compendium item with icon grid, color and a live card preview. */
export function ItemForm({ initial, onSave, onCancel }: Props) {
  const [draft, setDraft] = useState(initial)
  const [traits, setTraits] = useState(initial.traits.join(', '))
  const current = { ...draft, traits: traits.split(',').map((t) => t.trim()).filter(Boolean) }
  const { error, busy, submit } = useSubmit(current, onSave)
  return (
    <form onSubmit={submit} aria-label="Formulário do item" className="flex flex-col gap-3">
      <Identity draft={draft} onChange={setDraft} />
      <IconPicker value={draft.icon} color={draft.color} onChange={(icon) => setDraft({ ...draft, icon })} />
      <Stats draft={draft} onChange={setDraft} />
      <label className="flex flex-col gap-1 text-xs text-zinc-400">Características (separadas por vírgula)
        <input className={fieldCls} value={traits} onChange={(e) => setTraits(e.target.value)} />
      </label>
      <EffectsFields draft={draft} onChange={setDraft} />
      <ItemPreview item={current} label="Prévia do cartão" />
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      <div className="flex gap-2">
        <button type="submit" disabled={busy || !draft.name.trim()} className="rounded bg-emerald-700 px-3 py-1 text-sm font-medium hover:bg-emerald-600 disabled:opacity-40">Salvar item</button>
        <button type="button" onClick={onCancel} className="rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700">Cancelar</button>
      </div>
    </form>
  )
}
