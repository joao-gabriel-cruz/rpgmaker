import { useMutation } from '@tanstack/react-query'
import { createInvite, inviteLink } from './campaigns'

export function InvitePanel({ campaignId }: { campaignId: string }) {
  const invite = useMutation({ mutationFn: () => createInvite(campaignId) })
  const link = invite.data ? inviteLink(invite.data.token) : ''

  return (
    <div className="flex flex-col gap-2">
      <button onClick={() => invite.mutate()} disabled={invite.isPending}
        className="w-fit rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500">
        Gerar convite
      </button>
      {invite.isError && <p role="alert" className="text-sm text-red-400">{invite.error.message}</p>}
      {invite.data && (
        <div className="flex max-w-xl items-center gap-2">
          <input readOnly aria-label="Link do convite" value={link} onFocus={(e) => e.target.select()}
            className="flex-1 rounded bg-zinc-800 px-3 py-2 text-sm" />
          <button onClick={() => navigator.clipboard?.writeText(link)}
            className="rounded bg-zinc-800 px-3 py-2 text-sm hover:bg-zinc-700">
            Copiar
          </button>
        </div>
      )}
      {invite.data && <p className="text-xs text-zinc-400">Vale por 7 dias e só pode ser usado uma vez.</p>}
    </div>
  )
}
