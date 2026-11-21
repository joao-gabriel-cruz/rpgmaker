import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { api } from '../../lib/api'
import { ActionForm, BLANK_ACTION } from './ActionForm'
import { actionsKey, CATEGORY_LABELS, iconOf, useActions, type GameAction } from './actions'

const btn = 'rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700'

function ActionRow({ a, onToggle, onEdit, onRemove }: { a: GameAction; onToggle: (on: boolean) => void; onEdit: () => void; onRemove: () => void }) {
  const Icon = iconOf(a.icon)
  return (
    <li className="flex flex-wrap items-center gap-3 py-2">
      <Icon size={18} aria-hidden className={a.enabled ? 'text-zinc-200' : 'text-zinc-600'} />
      <span className={`min-w-40 flex-1 text-sm ${a.enabled ? '' : 'text-zinc-500 line-through'}`}>{a.name}
        <span className="ml-2 text-xs text-zinc-500">{CATEGORY_LABELS[a.category]}{a.builtin ? '' : ' · própria'}</span></span>
      <label className="flex items-center gap-1.5 text-xs text-zinc-300">
        <input type="checkbox" aria-label={`Ativar ${a.name}`} checked={a.enabled} onChange={(e) => onToggle(e.target.checked)} /> Ativa
      </label>
      <button className={btn} aria-label={`Editar ${a.name}`} onClick={onEdit}>Editar</button>
      <button className={btn} aria-label={a.builtin ? `Restaurar ${a.name}` : `Excluir ${a.name}`} onClick={onRemove}>
        {a.builtin ? 'Restaurar padrão' : 'Excluir'}</button>
    </li>
  )
}

/** Mestre tab "Ações": enable/disable, edit, restore defaults and create actions of the campaign. */
export function ActionsEditor({ campaignId }: { campaignId: string }) {
  const { actions } = useActions(campaignId)
  const qc = useQueryClient()
  const [editing, setEditing] = useState<GameAction | null>(null)
  const [error, setError] = useState('')
  const base = `/campaigns/${campaignId}/actions`
  const refresh = () => qc.invalidateQueries({ queryKey: actionsKey(campaignId) })
  const run = (p: Promise<unknown>) => p.then(() => { setError(''); return refresh() }, (e: Error) => setError(e.message))
  const save = (a: GameAction) => (a.id ? api.put(`${base}/${a.id}`, a) : api.post(base, a)).then(refresh)
  return (
    <div className="flex flex-col gap-3">
      <p className="text-sm text-zinc-400">Ações são os gestos que os personagens fazem na mesa: alvo, rolagem, animação e a consequência sugerida a você.</p>
      <ul className="flex flex-col divide-y divide-zinc-800">
        {actions.map((a) => (
          <ActionRow key={a.id} a={a} onToggle={(on) => run(save({ ...a, enabled: on }))} onEdit={() => setEditing(a)}
            onRemove={() => run(api.delete(`${base}/${a.id}`))} />
        ))}
      </ul>
      {editing ? <ActionForm key={editing.id} initial={editing} onSave={save} onCancel={() => setEditing(null)} />
        : <button className={`${btn} self-start`} onClick={() => setEditing(BLANK_ACTION)}>Nova ação</button>}
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
    </div>
  )
}
