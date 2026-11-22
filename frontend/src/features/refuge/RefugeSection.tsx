import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useMe } from '../auth/auth'
import { useCharacters } from '../characters/characters'
import { useRealtime } from '../realtime/realtimeContext'
import { RefugeItems, type Place } from './RefugeItems'
import { RefugeStats } from './RefugeStats'
import { BuildingsSection } from './BuildingsSection'
import { RegionPanel } from './RegionPanel'
import { ReservesPanel } from './ReservesPanel'
import { WeekPanel } from './WeekPanel'
import { WorkAllocation } from './WorkAllocation'
import { editableOf, refugeKey, saveRefuge, useRefuge, type Refuge, type RefugeEditable } from './refuge'

function useLiveRefuge(campaignId: string) {
  const queryClient = useQueryClient()
  const { subscribe } = useRealtime()
  useEffect(() => subscribe('refuge.updated', (p) => {
    queryClient.setQueryData(refugeKey(campaignId), (p as { refuge: Refuge }).refuge)
  }), [subscribe, queryClient, campaignId])
}

function useSaveRefuge(campaignId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (d: RefugeEditable) => saveRefuge(campaignId, d),
    onSuccess: (r) => queryClient.setQueryData(refugeKey(campaignId), r),
  })
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

function Header({ refuge, isGm, onChange }: { refuge: Refuge; isGm: boolean; onChange: (p: Partial<RefugeEditable>) => void }) {
  const field = (key: 'name' | 'description', label: string, cls: string) => (
    <input aria-label={label} readOnly={!isGm} defaultValue={refuge[key]} placeholder={isGm ? label : ''} key={refuge[key]}
      onBlur={(e) => { if ((key === 'description' || e.target.value.trim()) && e.target.value !== refuge[key]) onChange({ [key]: e.target.value }) }}
      className={`w-full max-w-3xl bg-transparent ${isGm ? 'rounded px-1 hover:bg-zinc-900 focus:bg-zinc-900' : ''} ${cls}`} />
  )
  return (
    <div className="flex flex-col gap-1">
      {field('name', 'Nome do Refúgio', 'text-xl font-semibold')}
      {(isGm || refuge.description) && field('description', 'Descrição do Refúgio', 'text-sm text-zinc-400')}
      <p className="px-1 text-xs text-zinc-500">Semana {refuge.week}</p>
    </div>
  )
}

export function RefugeSection({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const { data: refuge } = useRefuge(campaignId)
  const { data: chars } = useCharacters(campaignId)
  const { data: me } = useMe()
  const { send } = useRealtime()
  const [error, setError] = useState('')
  useLiveRefuge(campaignId)
  const save = useSaveRefuge(campaignId)
  if (!refuge || !Array.isArray(refuge.items)) return null
  const mine = (Array.isArray(chars) ? chars : []).filter((c) => c.kind === 'pc' && (isGm || c.ownerId === me?.id))
  const change = (patch: Partial<RefugeEditable>) => save.mutate({ ...editableOf(refuge), ...patch })
  const take = (characterId: string, to: Place, refugeIndex: number) => {
    setError('')
    send('refuge.item.withdraw', { characterId, to, refugeIndex }).catch((e: Error) => setError(e.message))
  }
  return (
    <div className="flex flex-col gap-3">
      <Header refuge={refuge} isGm={isGm} onChange={change} />
      <RefugeStats refuge={refuge} isGm={isGm} onChange={change} />
      <ReservesPanel refuge={refuge} isGm={isGm} onChange={change} />
      <RegionPanel refuge={refuge} isGm={isGm} onChange={change} />
      <WeekPanel campaignId={campaignId} refuge={refuge} isGm={isGm} />
      <BuildingsSection refuge={refuge} isGm={isGm} />
      <WorkAllocation refuge={refuge} isGm={isGm} />
      <Notes refuge={refuge} isGm={isGm} onSave={(notes) => change({ notes })} />
      {save.isError && <p role="alert" className="text-sm text-red-400">{save.error.message}</p>}
      <RefugeItems items={refuge.items} characters={mine} error={error} onTake={take} />
    </div>
  )
}
