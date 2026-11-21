import { useMutation } from '@tanstack/react-query'
import { useEffect, useState, type ReactNode } from 'react'
import { useMe } from '../auth/auth'
import type { Campaign } from '../campaigns/campaigns'
import { AppearanceEditor } from '../miniature/AppearanceEditor'
import { useRealtime } from '../realtime/realtimeContext'
import { RollableSheet } from '../rolls/RollableSheet'
import { NpcSheet } from './NpcSheet'
import { asNpc, asSheet } from './npc'
import { deleteCharacter, toSheetIssues, useCharacter, type CharacterView } from './characters'
import { useCharacterSync, type SaveStatus } from './useCharacterSync'

const STATUS_LABEL: Record<SaveStatus, string> = { saved: 'Salvo', saving: 'Salvando…', error: 'Erro ao salvar' }

function Toolbar({ links, character, canEdit, status, onRemove }: {
  links: ReactNode
  character: CharacterView
  canEdit: boolean
  status: SaveStatus
  onRemove: () => void
}) {
  return (
    <div className="flex flex-wrap items-center gap-3">
      {links}
      {!canEdit && <span className="text-sm text-zinc-400">Ficha de {character.ownerName || 'outro jogador'} (somente leitura)</span>}
      {canEdit && (
        <>
          <span role="status" aria-label="Estado do salvamento" className={`ml-auto text-sm ${status === 'error' ? 'text-red-400' : 'text-zinc-400'}`}>{STATUS_LABEL[status]}</span>
          <button onClick={() => window.confirm(character.kind === 'npc' ? 'Apagar este NPC?' : 'Apagar este personagem?') && onRemove()}
            className="rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-red-900">{character.kind === 'npc' ? 'Apagar NPC' : 'Apagar personagem'}</button>
        </>
      )}
    </div>
  )
}

function Banners({ status, error, hasIssues }: { status: SaveStatus; error: string; hasIssues: boolean }) {
  if (status === 'error') return <p role="alert" className="rounded bg-red-950 px-3 py-2 text-sm text-red-300">Não salvo: {error}</p>
  if (status === 'saved' && hasIssues) {
    return <p role="status" className="rounded bg-amber-950 px-3 py-2 text-sm text-amber-300">Ficha salva com avisos: há valores fora dos limites de criação.</p>
  }
  return null
}

function SheetBody({ campaign, character, canEdit, sync }: {
  campaign: Campaign
  character: CharacterView
  canEdit: boolean
  sync: ReturnType<typeof useCharacterSync>
}) {
  const { send } = useRealtime()
  const [depositError, setDepositError] = useState('')
  const { sheet, issues, onChange } = sync
  return (
    <>
      {depositError && <p role="alert" className="rounded bg-red-950 px-3 py-2 text-sm text-red-300">{depositError}</p>}
      <AppearanceEditor appearance={sheet.appearance} readOnly={!canEdit} assimilationLevel={sheet.tugOfWar?.assimilationLevel}
        onChange={(appearance) => onChange({ ...sheet, appearance })} />
      {character.kind === 'npc'
        ? <NpcSheet sheet={asNpc(sheet)} onChange={(n) => onChange(asSheet(n))} readOnly={!canEdit} issues={toSheetIssues(issues)} />
        : <RollableSheet campaignId={campaign.id} characterId={character.id} isGm={campaign.role === 'gm'}
          sheet={sheet} onChange={onChange} readOnly={!canEdit} issues={toSheetIssues(issues)}
          onDepositItem={(from, index) => {
            setDepositError('')
            send('refuge.item.deposit', { characterId: character.id, from, index }).catch((e: Error) => setDepositError(e.message))
          }}
          onUseItem={canEdit ? (place, index) => {
            setDepositError('')
            send('item.use', { characterId: character.id, place, index }).catch((e: Error) => setDepositError(e.message))
          } : undefined} />}
    </>
  )
}

interface EditorProps {
  campaign: Campaign
  character: CharacterView
  canEdit: boolean
  links: ReactNode
  onGone: () => void
}

function Editor({ campaign, character, canEdit, links, onGone }: EditorProps) {
  const { subscribe } = useRealtime()
  const sync = useCharacterSync(campaign.id, character, canEdit)
  const { issues, status, error } = sync
  const remove = useMutation({ mutationFn: () => deleteCharacter(campaign.id, character.id), onSuccess: onGone })
  useEffect(() => subscribe('character.deleted', (p) => {
    if ((p as { id: string }).id === character.id) onGone()
  }), [subscribe, character.id, onGone])

  return (
    <section className="@container flex flex-col gap-4">
      <Toolbar links={links} character={character} canEdit={canEdit} status={status} onRemove={() => remove.mutate()} />
      {canEdit && <Banners status={status} error={error} hasIssues={issues.length > 0} />}
      <SheetBody campaign={campaign} character={character} canEdit={canEdit} sync={sync} />
    </section>
  )
}

interface Props {
  campaign: Campaign
  characterId: string
  /** Navigation shown at the left of the toolbar (back link on the page, "open in page" in the drawer). */
  links: ReactNode
  /** The character was deleted (here or by someone else): leave the editor. */
  onGone: () => void
}

/** The full sheet (appearance + PC/NPC sheet) with autosave and permissions; used by the page and the table drawer. */
export function CharacterEditor({ campaign, characterId, links, onGone }: Props) {
  const { data: me } = useMe()
  const character = useCharacter(campaign.id, characterId)
  if (character.isError) return <div className="flex gap-3"><p role="alert" className="text-red-400">Personagem não encontrado.</p>{links}</div>
  if (!character.data || !me) return <p className="text-zinc-400">Carregando…</p>
  const canEdit = campaign.role === 'gm' || character.data.ownerId === me.id
  return <Editor key={character.data.id} campaign={campaign} character={character.data} canEdit={canEdit} links={links} onGone={onGone} />
}
