import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ExternalLink } from 'lucide-react'
import { Link } from 'react-router'
import { ConditionIcons } from '../ambience/ConditionIcons'
import { conditionsOf } from '../ambience/conditions'
import { HealthBar, Portrait, StateBadges, TugBar } from '../hotbar/HotbarStatus'
import { useTableCharacter, useTableCharacterOptional } from '../hotbar/TableCharacterContext'
import { charactersKey, createCharacter, useCharacters, type CharacterView } from './characters'
import { vitalsOf } from '../hotbar/vitals'
import { useLiveCharacters } from './useLiveCharacters'

function useCreate(campaignId: string, kind: 'pc' | 'npc') {
  const table = useTableCharacterOptional()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => createCharacter(campaignId, kind),
    onSuccess: (c) => {
      void queryClient.invalidateQueries({ queryKey: charactersKey(campaignId) })
      table?.openSheet(c.id)
    },
  })
}

function Card({ campaignId, c, showOwner }: { campaignId: string; c: CharacterView; showOwner: boolean }) {
  const table = useTableCharacter()
  const v = vitalsOf(c)
  return (
    <li className="relative rounded-xl bg-zinc-800/60 transition-colors hover:bg-zinc-800">
      <button type="button" aria-pressed={table.sheetId === c.id} onClick={() => table.openSheet(c.id)}
        className="flex w-full flex-col gap-2 rounded-xl px-3 py-2 text-left active:scale-[0.99] focus-visible:outline-2 focus-visible:outline-emerald-400">
        <span className="flex items-center gap-3 pr-7">
          <Portrait character={c} />
          <span className="min-w-0 flex-1">
            <span className="block truncate text-sm font-medium">{c.name}</span>
            {showOwner && <span className="block truncate text-xs text-zinc-400">{c.ownerName || 'Sem dono'}</span>}
            <span className="flex flex-wrap items-center gap-1"><StateBadges v={v} /><ConditionIcons ids={conditionsOf(c.sheet)} status={c.sheet} /></span>
          </span>
        </span>
        <HealthBar v={v} />
        <TugBar v={v} />
      </button>
      <Link to={`/campaigns/${campaignId}/characters/${c.id}`} aria-label={`Abrir ficha de ${c.name} em página`} title="Abrir em página"
        className="absolute right-2 top-2 rounded-md p-1 text-zinc-400 transition-colors hover:bg-zinc-700 hover:text-white focus-visible:outline-2 focus-visible:outline-emerald-400">
        <ExternalLink size={14} aria-hidden />
      </Link>
    </li>
  )
}

function CharacterList({ campaignId, list, empty, showOwner }: { campaignId: string; list: CharacterView[]; empty: string; showOwner: boolean }) {
  return (
    <ul className="flex flex-col gap-1.5">
      {list.map((c) => <Card key={c.id} campaignId={campaignId} c={c} showOwner={showOwner} />)}
      {list.length === 0 && <li className="text-sm text-zinc-400">{empty}</li>}
    </ul>
  )
}

function CreateButton({ campaignId, kind, label }: { campaignId: string; kind: 'pc' | 'npc'; label: string }) {
  const create = useCreate(campaignId, kind)
  return (
    <>
      <button onClick={() => create.mutate()} disabled={create.isPending}
        className="w-fit rounded-lg bg-emerald-600 px-3 py-1.5 text-sm font-medium transition-[background-color,transform] hover:bg-emerald-500 active:scale-[0.96]">
        {label}
      </button>
      {create.isError && <p role="alert" className="text-sm text-red-400">{create.error.message}</p>}
    </>
  )
}

export function CharactersSection({ campaignId, isGm = false }: { campaignId: string; isGm?: boolean }) {
  const { data } = useCharacters(campaignId)
  useLiveCharacters(campaignId)
  const list = Array.isArray(data) ? data : []
  const pcs = list.filter((c) => c.kind !== 'npc')
  const npcs = list.filter((c) => c.kind === 'npc')

  return (
    <div className="flex flex-col gap-2">
      <h2 className="text-lg font-semibold">Personagens</h2>
      <CharacterList campaignId={campaignId} list={pcs} empty="Nenhum personagem ainda." showOwner />
      <CreateButton campaignId={campaignId} kind="pc" label="Criar personagem" />
      {isGm && (
        <>
          <h3 className="mt-4 text-base font-semibold">NPCs</h3>
          <CharacterList campaignId={campaignId} list={npcs} empty="Nenhum NPC ainda." showOwner={false} />
          <CreateButton campaignId={campaignId} kind="npc" label="Criar NPC" />
        </>
      )}
    </div>
  )
}
