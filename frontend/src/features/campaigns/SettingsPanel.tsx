import { useMutation, useQueryClient } from '@tanstack/react-query'
import { updateSettings, type Campaign, type CreationLimits } from './campaigns'

export function SettingsPanel({ campaign }: { campaign: Campaign }) {
  const queryClient = useQueryClient()
  const mutation = useMutation({
    mutationFn: (v: CreationLimits) => updateSettings(campaign.id, v),
    onSuccess: (c) => queryClient.setQueryData(['campaigns', campaign.id], c),
  })
  return (
    <div className="flex max-w-xl flex-col gap-1">
      <h2 className="text-xl font-semibold">Configurações</h2>
      <label className="flex flex-col gap-1 text-sm">
        Limites de criação de personagem
        <select value={campaign.settings?.creationLimits ?? 'warn'} disabled={mutation.isPending}
          onChange={(e) => mutation.mutate(e.target.value as CreationLimits)}
          className="w-fit rounded bg-zinc-800 px-3 py-2">
          <option value="warn">Avisar (a ficha é salva mesmo acima do limite)</option>
          <option value="block">Bloquear (a ficha acima do limite não é salva)</option>
        </select>
      </label>
      {mutation.isError && <p role="alert" className="text-sm text-red-400">{mutation.error.message}</p>}
    </div>
  )
}
