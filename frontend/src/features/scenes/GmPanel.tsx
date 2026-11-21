import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { deleteSnapshot, listSnapshots, saveSnapshot, type SnapshotDto } from './scenes'

const btn = 'rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700 disabled:opacity-50'

interface DraftProps {
  draft: boolean
  onCommand: (type: string, extra?: object) => void
}

function DraftBar({ draft, onCommand }: DraftProps) {
  return (
    <div className="flex flex-wrap items-center gap-3">
      <label className="flex items-center gap-2 text-sm">
        <input type="checkbox" checked={draft}
          onChange={(e) => onCommand(e.target.checked ? 'map.draft.start' : 'map.draft.discard')} />
        Rascunho
      </label>
      {draft && (
        <>
          <p role="status" className="rounded bg-amber-900/60 px-3 py-1 text-sm text-amber-200">
            Editando rascunho — jogadores não veem
          </p>
          <button type="button" className={btn} onClick={() => onCommand('map.draft.apply')}>Aplicar</button>
          <button type="button" className={btn} onClick={() => onCommand('map.draft.discard')}>Descartar</button>
        </>
      )}
    </div>
  )
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

function SnapshotRow({ s, draft, onCommand, onDelete }: DraftProps & { s: SnapshotDto; onDelete: () => void }) {
  return (
    <li className="flex items-center gap-2">
      <span>{s.name}</span>
      <button type="button" className={btn} disabled={draft} title={draft ? 'Aplique ou descarte o rascunho antes' : undefined}
        aria-label={`Disparar ${s.name}`} onClick={() => onCommand('map.snapshot.apply', { snapshotId: s.id })}>Disparar</button>
      <button type="button" className={btn} aria-label={`Apagar ${s.name}`} onClick={onDelete}>Apagar</button>
    </li>
  )
}

function Snapshots({ campaignId, sceneId, draft, onCommand }: Props) {
  const { list, save, del } = useSnapshots(campaignId, sceneId)
  const onSave = () => {
    const name = window.prompt('Nome do estado salvo')?.trim()
    if (name) save.mutate(name)
  }
  return (
    <section aria-label="Estados salvos" className="flex flex-col gap-1 text-sm">
      <div className="flex items-center gap-3">
        <h3 className="font-semibold">Estados salvos</h3>
        <button type="button" className={btn} onClick={onSave}>Salvar estado atual…</button>
      </div>
      {(save.error || del.error) && <p role="alert" className="text-red-400">{(save.error ?? del.error)?.message}</p>}
      <ul className="flex flex-col gap-1">
        {list.map((s) => <SnapshotRow key={s.id} s={s} draft={draft} onCommand={onCommand} onDelete={() => del.mutate(s.id)} />)}
      </ul>
    </section>
  )
}

interface Props extends DraftProps {
  campaignId: string
  sceneId: string
}

/** Mestre-only tools under the map: hidden draft mode and saved states. */
export function GmPanel(props: Props) {
  return (
    <div className="flex flex-col gap-2 rounded border border-zinc-800 p-2">
      <DraftBar draft={props.draft} onCommand={props.onCommand} />
      <Snapshots {...props} />
    </div>
  )
}
