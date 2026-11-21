import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router'
import { campaignsKey, createCampaign, useSystems } from './campaigns'

function useCreateCampaign(name: string, systemId: string, onDone: () => void) {
  const { data: systems } = useSystems()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  return useMutation({
    mutationFn: () => createCampaign(name, systemId || systems?.[0]?.id || ''),
    onSuccess: (c) => {
      queryClient.invalidateQueries({ queryKey: campaignsKey })
      onDone()
      navigate(`/campaigns/${c.id}`)
    },
  })
}

function SystemSelect({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const { data: systems } = useSystems()
  return (
    <label className="flex flex-col gap-1 text-sm">
      Sistema
      <select value={value} onChange={(e) => onChange(e.target.value)} className="rounded bg-zinc-800 px-3 py-2">
        {systems?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
      </select>
    </label>
  )
}

export function NewCampaignForm({ onDone }: { onDone: () => void }) {
  const [name, setName] = useState('')
  const [systemId, setSystemId] = useState('')
  const create = useCreateCampaign(name, systemId, onDone)
  const submit = (e: FormEvent) => {
    e.preventDefault()
    create.mutate()
  }

  return (
    <form onSubmit={submit} className="flex max-w-md flex-col gap-3 rounded-lg bg-zinc-900 p-4">
      <label className="flex flex-col gap-1 text-sm">
        Nome da campanha
        <input required value={name} onChange={(e) => setName(e.target.value)}
          className="rounded bg-zinc-800 px-3 py-2" />
      </label>
      <SystemSelect value={systemId} onChange={setSystemId} />
      {create.isError && <p role="alert" className="text-sm text-red-400">{create.error.message}</p>}
      <button type="submit" disabled={create.isPending}
        className="rounded bg-emerald-600 px-3 py-2 font-medium hover:bg-emerald-500 disabled:opacity-50">
        Criar campanha
      </button>
    </form>
  )
}
