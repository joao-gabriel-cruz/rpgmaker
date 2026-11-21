import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useMe } from '../auth/auth'
import { useCharacters } from '../characters/characters'
import { useRealtime } from '../realtime/realtimeContext'
import { RefugeItems, type Place } from './RefugeItems'
import { RefugeStats } from './RefugeStats'
import { detailsOf, refugeKey, saveRefuge, useRefuge, type NumericField, type Refuge } from './refuge'

function useLiveRefuge(campaignId: string) {
  const queryClient = useQueryClient()
  const { subscribe } = useRealtime()
  useEffect(() => subscribe('refuge.updated', (p) => {
    queryClient.setQueryData(refugeKey(campaignId), (p as { refuge: Refuge }).refuge)
  }), [subscribe, queryClient, campaignId])
}

function Notes({ refuge, isGm, onSave }: { refuge: Refuge; isGm: boolean; onSave: (notes: string) => void }) {
  const [draft, setDraft] = useState<string | null>(null)
  const value = draft ?? refuge.notes
  return (
    <label className="flex max-w-3xl flex-col gap-1 text-sm text-zinc-400">Notas do Refúgio
      <textarea rows={3} readOnly={!isGm} value={value} onChange={(e) => setDraft(e.target.value)}
        onBlur={() => { if (draft !== null && draft !== refuge.notes) onSave(draft); setDraft(null) }}
        className="rounded bg-zinc-900 px-3 py-2 text-zinc-100" />
    </label>
  )
}

export function RefugeSection({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const { data: refuge } = useRefuge(campaignId)
  const { data: chars } = useCharacters(campaignId)
  const { data: me } = useMe()
  const { send } = useRealtime()
  const queryClient = useQueryClient()
  const [error, setError] = useState('')
  useLiveRefuge(campaignId)
  const save = useMutation({
    mutationFn: (d: ReturnType<typeof detailsOf>) => saveRefuge(campaignId, d),
    onSuccess: (r) => queryClient.setQueryData(refugeKey(campaignId), r),
  })
  if (!refuge || !Array.isArray(refuge.items)) return null
  const mine = (Array.isArray(chars) ? chars : []).filter((c) => c.kind === 'pc' && (isGm || c.ownerId === me?.id))
  const change = (patch: Partial<Refuge>) => save.mutate(detailsOf({ ...refuge, ...patch }))
  const take = (characterId: string, to: Place, refugeIndex: number) => {
    setError('')
    send('refuge.item.withdraw', { characterId, to, refugeIndex }).catch((e: Error) => setError(e.message))
  }
  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-xl font-semibold">Refúgio</h2>
      <p className="text-sm text-zinc-400">{refuge.name}{refuge.description && ` — ${refuge.description}`}</p>
      <RefugeStats refuge={refuge} isGm={isGm} onChange={(f: NumericField, v) => change({ [f]: v })} />
      <Notes refuge={refuge} isGm={isGm} onSave={(notes) => change({ notes })} />
      {save.isError && <p role="alert" className="text-sm text-red-400">{save.error.message}</p>}
      <RefugeItems items={refuge.items} characters={mine} error={error} onTake={take} />
    </div>
  )
}
