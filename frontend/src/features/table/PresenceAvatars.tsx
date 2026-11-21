import { roleLabel } from '../campaigns/campaigns'
import { useCharacters } from '../characters/characters'
import { CharacterAvatar } from '../miniature/CharacterAvatar'
import type { ConnectionStatus } from '../realtime/CampaignSocket'
import { useRealtime } from '../realtime/realtimeContext'
import { usePresence, type OnlineUser } from '../realtime/usePresence'
import { Tip } from './Tip'

const STATUS_LABEL: Record<ConnectionStatus, string> = {
  connecting: 'Conectando…',
  open: 'Conectado',
  reconnecting: 'Reconectando…',
  closed: 'Desconectado',
}

const initials = (name: string) => name.trim().split(/\s+/).slice(0, 2).map((w) => w[0]?.toUpperCase()).join('')

function PresenceFace({ user, look }: { user: OnlineUser; look: unknown }) {
  const label = `${user.displayName} (${roleLabel(user.role)})`
  if (look) return <span role="img" aria-label={label} className="block rounded-full ring-2 ring-zinc-900"><CharacterAvatar name={user.displayName} appearance={look} ring={user.role === 'gm' ? '#d97706' : undefined} size="h-8 w-8" shape="circle" /></span>
  return (
    <span role="img" aria-label={label}
      className={`flex h-8 w-8 items-center justify-center rounded-full text-xs font-semibold ring-2 ring-zinc-900 ${user.role === 'gm' ? 'bg-amber-600 text-white' : 'bg-emerald-700 text-white'}`}>
      {initials(user.displayName)}
    </span>
  )
}

/** Stacked initials of who is online (tooltip: name and role) plus a discreet connection status. */
export function PresenceAvatars({ campaignId }: { campaignId: string }) {
  const { status } = useRealtime()
  const users = usePresence()
  const chars = useCharacters(campaignId).data
  const withLook = (id: string) => (Array.isArray(chars) ? chars : []).find((c) => c.ownerId === id && c.sheet?.appearance)
  return (
    <div className="flex items-center gap-2">
      <span role="status" className={status === 'open' ? 'sr-only' : 'text-xs text-amber-400'}>{STATUS_LABEL[status]}</span>
      <ul aria-label="Online agora" className="flex -space-x-2">
        {users.map((u) => (
          <li key={u.userId} className="group relative">
            <PresenceFace user={u} look={withLook(u.userId)?.sheet.appearance} />
            <Tip side="bottom">{u.displayName} · {roleLabel(u.role)}</Tip>
          </li>
        ))}
      </ul>
    </div>
  )
}
