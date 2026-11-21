import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { characterKey, charactersKey, type CharacterView } from './characters'

/** Keeps the characters list (and each character's own cache entry) live from the socket. */
export function useLiveCharacters(campaignId: string) {
  const queryClient = useQueryClient()
  const { subscribe } = useRealtime()
  useEffect(() => {
    const refresh = () => void queryClient.invalidateQueries({ queryKey: charactersKey(campaignId) })
    const updated = (p: unknown) => {
      const c = p as CharacterView
      if (!c?.id || !c.sheet) return refresh()
      queryClient.setQueryData(characterKey(campaignId, c.id), c)
      queryClient.setQueryData<CharacterView[]>(charactersKey(campaignId), (list) =>
        Array.isArray(list) && list.some((x) => x.id === c.id) ? list.map((x) => (x.id === c.id ? c : x)) : list)
    }
    const offs = [subscribe('character.created', refresh), subscribe('character.deleted', refresh), subscribe('character.updated', updated)]
    return () => offs.forEach((off) => off())
  }, [subscribe, queryClient, campaignId])
}
