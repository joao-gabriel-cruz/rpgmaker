import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Check, FilePenLine, Play, Save, Trash2, X } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { deleteSnapshot, listSnapshots, saveSnapshot, type SnapshotDto } from '../../scenes/scenes'
import { BTN, Hint, INPUT, Section, Switch } from './ui'

interface Props {
  campaignId: string
  sceneId: string
  draft: boolean
  onCommand: (type: string, extra?: object) => void
}

function useSnapshots(campaignId: string, sceneId: string) {
  const queryClient = useQueryClient()
  const key = ['campaigns', campaignId, 'scenes', sceneId, 'snapshots']
  const query = useQuery({ queryKey: key, queryFn: () => listSnapshots(campaignId, sceneId) })
  const refresh = () => queryClient.invalidateQueries({ queryKey: key })
  const save = useMutation({ mutationFn: (name: string) => saveSnapshot(campaignId, sceneId, name), onSuccess: refresh })
  const del = useMutation({ mutationFn: (id: string) => deleteSnapshot(campaignId, sceneId, id), onSuccess: refresh })
  return { list: Array.isArray(query.data) ? query.data : [], save, del }
}

function SaveForm({ onSave }: { onSave: (name: string) => void }) {
  const [name, setName] = useState('')
  const submit = (e: FormEvent) => {
    e.preventDefault()
    const n = name.trim()
    if (!n) return
    onSave(n)
    setName('')
  }
  return (
    <form onSubmit={submit} className="flex gap-1">
      <input aria-label="Nome do estado salvo" placeholder="Nome do estado" value={name} onChange={(e) => setName(e.target.value)} className={`${INPUT} min-w-0 flex-1`} />
      <button type="submit" aria-label="Salvar estado atual" disabled={!name.trim()} className={`${BTN} flex items-center gap-1`}>
        <Save size={14} strokeWidth={1.5} aria-hidden />Salvar
      </button>
    </form>
  )
}

function SnapshotRow({ s, draft, onFire, onDelete }: { s: SnapshotDto; draft: boolean; onFire: () => void; onDelete: () => void }) {
  return (
    <li className="flex items-center gap-1 rounded-lg bg-zinc-800/60 p-1 pl-2">
      <span className="min-w-0 flex-1 truncate text-sm text-zinc-100">{s.name}</span>
      <button type="button" className={`${BTN} flex items-center gap-1`} disabled={draft} aria-label={`Disparar ${s.name}`}
        title={draft ? 'Aplique ou descarte o rascunho antes' : undefined} onClick={onFire}>
        <Play size={13} strokeWidth={1.5} aria-hidden />Disparar
      </button>
      <button type="button" className={`${BTN} px-1.5 text-red-300`} aria-label={`Apagar ${s.name}`} onClick={onDelete}>
        <Trash2 size={13} strokeWidth={1.5} aria-hidden />
      </button>
    </li>
  )
}

/** Draft mode (edit without players seeing) and saved scene states; the banner lives over the map. */
export function DraftOptions({ campaignId, sceneId, draft, onCommand }: Props) {
  const { list, save, del } = useSnapshots(campaignId, sceneId)
  const err = save.error ?? del.error
  return (
    <div className="flex flex-col gap-3">
      <Section title="Rascunho">
        <Switch label="Rascunho" icon={FilePenLine} on={draft} onChange={() => onCommand(draft ? 'map.draft.discard' : 'map.draft.start')} />
        {draft ? (
          <div className="flex gap-1">
            <button type="button" className={`${BTN} flex flex-1 items-center justify-center gap-1 bg-emerald-700 hover:bg-emerald-600`} onClick={() => onCommand('map.draft.apply')}>
              <Check size={14} strokeWidth={1.5} aria-hidden />Aplicar
            </button>
            <button type="button" className={`${BTN} flex flex-1 items-center justify-center gap-1`} onClick={() => onCommand('map.draft.discard')}>
              <X size={14} strokeWidth={1.5} aria-hidden />Descartar
            </button>
          </div>
        ) : <Hint>Edite o mapa sem que os jogadores vejam; aplique quando estiver pronto.</Hint>}
      </Section>
      <section aria-label="Estados salvos" className="flex flex-col gap-1.5">
        <h3 className="text-[11px] font-medium uppercase tracking-wide text-zinc-400">Estados salvos</h3>
        <SaveForm onSave={(n) => save.mutate(n)} />
        {err && <p role="alert" className="text-xs text-red-400">{err.message}</p>}
        <ul className="flex flex-col gap-1">
          {list.map((s) => <SnapshotRow key={s.id} s={s} draft={draft} onFire={() => onCommand('map.snapshot.apply', { snapshotId: s.id })} onDelete={() => del.mutate(s.id)} />)}
        </ul>
        {list.length === 0 && <Hint>Nenhum estado salvo nesta cena.</Hint>}
      </section>
    </div>
  )
}
