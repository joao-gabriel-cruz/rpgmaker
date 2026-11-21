import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../lib/api'
import { useMe } from '../auth/auth'
import { removeMember, roleLabel, type MemberInfo } from './campaigns'

function Members({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const { data: me } = useMe()
  const queryClient = useQueryClient()
  const key = ['campaigns', campaignId, 'members']
  const { data: members } = useQuery({
    queryKey: key,
    queryFn: () => api.get<MemberInfo[]>(`/campaigns/${campaignId}/members`),
  })
  const remove = useMutation({
    mutationFn: (userId: string) => removeMember(campaignId, userId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: key }),
  })

  return (
    <ul className="flex max-w-xl flex-col gap-2">
      {members?.map((m) => (
        <li key={m.userId} className="flex items-center gap-3 rounded-lg bg-zinc-800/60 px-4 py-2">
          <span className="font-medium">{m.displayName}</span>
          <span className="text-sm text-zinc-400">{roleLabel(m.role)}</span>
          {isGm && m.userId !== me?.id && m.role === 'player' && (
            <button onClick={() => remove.mutate(m.userId)} aria-label={`Remover ${m.displayName}`}
              className="ml-auto rounded-lg bg-zinc-800 px-3 py-1 text-sm transition-colors hover:bg-red-900 active:scale-[0.96]">
              Remover
            </button>
          )}
        </li>
      ))}
    </ul>
  )
}

export function MembersSection({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  return (
    <div className="flex flex-col gap-2">
      <h3 className="text-lg font-semibold">Membros</h3>
      <Members campaignId={campaignId} isGm={isGm} />
    </div>
  )
}
