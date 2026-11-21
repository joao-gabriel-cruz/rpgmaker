import { useState, type FormEvent } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { emptyConflict, newId, type Conflict, type Objective } from './conflict'
import { ActivationsEditor, ThreatsEditor } from './EditorThreats'
import { Group, NumberField, input, patchAt, smallBtn } from './editorParts'

function Conditions({ items, onChange }: { items: string[]; onChange: (c: string[]) => void }) {
  return (
    <Group title="Condicionantes" onAdd={() => onChange([...items, ''])}>
      {items.map((c, i) => (
        <div key={i} className="flex gap-2">
          <input aria-label="Condicionante" value={c} onChange={(e) => onChange(items.map((x, j) => (j === i ? e.target.value : x)))} className={`${input} flex-1`} />
          <button type="button" className={smallBtn} onClick={() => onChange(items.filter((_, j) => j !== i))}>Remover</button>
        </div>
      ))}
    </Group>
  )
}

function Objectives({ items, onChange }: { items: Objective[]; onChange: (o: Objective[]) => void }) {
  const add = () => onChange([...items, { id: newId(), title: '', kind: 'main', scope: 'collective', target: 5, progress: 0 }])
  return (
    <Group title="Objetivos" onAdd={add}>
      {items.map((o, i) => {
        const set = (p: Partial<Objective>) => onChange(patchAt(items, i, p))
        return (
          <div key={o.id} className="flex flex-wrap items-end gap-2">
            <input aria-label="Título do objetivo" placeholder="Título" value={o.title} onChange={(e) => set({ title: e.target.value })} className={input} />
            <select aria-label="Tipo" value={o.kind} onChange={(e) => set({ kind: e.target.value as Objective['kind'] })} className={input}>
              <option value="main">Principal</option><option value="secondary">Secundário</option>
            </select>
            <select aria-label="Alcance" value={o.scope} onChange={(e) => set({ scope: e.target.value as Objective['scope'] })} className={input}>
              <option value="collective">Coletivo</option><option value="individual">Individual</option>
            </select>
            <NumberField label="Meta" min={1} value={o.target} onChange={(n) => set({ target: n })} />
            <button type="button" className={smallBtn} onClick={() => onChange(items.filter((_, j) => j !== i))}>Remover</button>
          </div>
        )
      })}
    </Group>
  )
}

/** Form where the Mestre edits the whole score sheet; saved with conflict.save. */
export function ConflictEditor({ campaignId, initial, onDone }: { campaignId: string; initial: Conflict | null; onDone: () => void }) {
  const { send } = useRealtime()
  const [draft, setDraft] = useState<Conflict>(initial ?? emptyConflict())
  const [error, setError] = useState('')
  const set = (p: Partial<Conflict>) => setDraft((d) => ({ ...d, ...p }))
  const submit = (ev: FormEvent) => {
    ev.preventDefault()
    const { title, sceneId, conditions, objectives, threats, activations } = draft
    send('conflict.save', { title, sceneId, conditions, objectives, threats, activations: activations ?? [] }).then(onDone, (e: Error) => setError(e.message))
  }
  return (
    <form onSubmit={submit} className="flex flex-col gap-3 rounded bg-zinc-900/60 p-3">
      <input aria-label="Título da partitura" placeholder="Título da cena de conflito" value={draft.title} onChange={(e) => set({ title: e.target.value })} className={input} />
      <Conditions items={draft.conditions} onChange={(conditions) => set({ conditions })} />
      <Objectives items={draft.objectives} onChange={(objectives) => set({ objectives })} />
      <ThreatsEditor campaignId={campaignId} threats={draft.threats} onChange={(threats) => set({ threats })} />
      <ActivationsEditor activations={draft.activations ?? []} threats={draft.threats} onChange={(activations) => set({ activations })} />
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      <button type="submit" className="self-start rounded bg-emerald-600 px-4 py-1 text-sm font-medium text-white">Salvar partitura</button>
    </form>
  )
}
