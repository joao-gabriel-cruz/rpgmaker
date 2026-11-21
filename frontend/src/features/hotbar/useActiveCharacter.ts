import { useMe } from '../auth/auth'
import { useCharacters, type CharacterView } from '../characters/characters'
import { useLiveCharacters } from '../characters/useLiveCharacters'
import { useTableCharacterOptional } from './TableCharacterContext'

/** The hotbar's character: the player's own (switchable) or, for the Mestre, whichever is selected. */
export function useActiveCharacter(campaignId: string, isGm: boolean) {
  const { data: me } = useMe()
  const { data } = useCharacters(campaignId)
  const selectedId = useTableCharacterOptional()?.selectedId ?? null
  useLiveCharacters(campaignId)
  const all = Array.isArray(data) ? data : []
  const options: CharacterView[] = isGm ? all : all.filter((c) => c.ownerId === me?.id && c.kind !== 'npc')
  const active = options.find((c) => c.id === selectedId) ?? (isGm ? undefined : options[0])
  return { options, active, userId: me?.id ?? '' }
}
