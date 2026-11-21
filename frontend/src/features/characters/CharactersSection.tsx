import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { Link, useNavigate } from 'react-router'
import { useRealtime } from '../realtime/realtimeContext'
import { charactersKey, createCharacter, useCharacters, type CharacterView } from './characters'

const EVENTS = ['character.created', 'character.updated', 'character.deleted']

function useLiveList(campaignId: string) {
  const queryClient = useQueryClient()
  const { subscribe } = useRealtime()
  useEffect(() => {
    const refresh = () => void queryClient.invalidateQueries({ queryKey: charactersKey(campaignId) })
    const offs = EVENTS.map((t) => subscribe(t, refresh))
    return () => offs.forEach((off) => off())
  }, [subscribe, queryClient, campaignId])
}

function useCreate(campaignId: string, kind: 'pc' | 'npc') {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => createCharacter(campaignId, kind),
    onSuccess: (c) => {
      void queryClient.invalidateQueries({ queryKey: charactersKey(campaignId) })
      navigate(`/campaigns/${campaignId}/characters/${c.id}`)
    },
  })
}

function CharacterList({ campaignId, list, empty, showOwner }: { campaignId: string; list: CharacterView[]; empty: string; showOwner: boolean }) {
  return (
    <ul className="flex max-w-xl flex-col gap-2">
      {list.map((c) => (
        <li key={c.id}>
          <Link to={`/campaigns/${campaignId}/characters/${c.id}`}
            className="flex items-center gap-3 rounded bg-zinc-900 px-4 py-2 hover:bg-zinc-800">
            <span className="font-medium">{c.name}</span>
            {showOwner && <span className="text-sm text-zinc-400">{c.ownerName || 'Sem dono'}</span>}
            {c.derived && <span className="ml-auto text-sm text-emerald-400">{c.derived.levelName}</span>}
          </Link>
        </li>
      ))}
      {list.length === 0 && <li className="text-sm text-zinc-400">{empty}</li>}
    </ul>
  )
}

function CreateButton({ campaignId, kind, label }: { campaignId: string; kind: 'pc' | 'npc'; label: string }) {
  const create = useCreate(campaignId, kind)
  return (
    <>
      <button onClick={() => create.mutate()} disabled={create.isPending}
        className="w-fit rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500">
        {label}
      </button>
      {create.isError && <p role="alert" className="text-sm text-red-400">{create.error.message}</p>}
    </>
  )
}

export function CharactersSection({ campaignId, isGm = false }: { campaignId: string; isGm?: boolean }) {
  const { data } = useCharacters(campaignId)
  useLiveList(campaignId)
  const list = Array.isArray(data) ? data : []
  const pcs = list.filter((c) => c.kind !== 'npc')
  const npcs = list.filter((c) => c.kind === 'npc')

  return (
    <div className="flex flex-col gap-2">
      <h2 className="text-xl font-semibold">Personagens</h2>
      <CharacterList campaignId={campaignId} list={pcs} empty="Nenhum personagem ainda." showOwner />
      <CreateButton campaignId={campaignId} kind="pc" label="Criar personagem" />
      {isGm && (
        <>
          <h3 className="mt-4 text-lg font-semibold">NPCs</h3>
          <CharacterList campaignId={campaignId} list={npcs} empty="Nenhum NPC ainda." showOwner={false} />
          <CreateButton campaignId={campaignId} kind="npc" label="Criar NPC" />
        </>
      )}
    </div>
  )
}
