import { roleLabel } from '../campaigns/campaigns'
import type { ConnectionStatus } from '../realtime/CampaignSocket'
import { useRealtime } from '../realtime/realtimeContext'
import { usePresence } from '../realtime/usePresence'
import { Tip } from './Tip'

const STATUS_LABEL: Record<ConnectionStatus, string> = {
  connecting: 'Conectando…',
  open: 'Conectado',
  reconnecting: 'Reconectando…',
  closed: 'Desconectado',
}

const initials = (name: string) => name.trim().split(/\s+/).slice(0, 2).map((w) => w[0]?.toUpperCase()).join('')

/** Stacked initials of who is online (tooltip: name and role) plus a discreet connection status. */
export function PresenceAvatars() {
  const { status } = useRealtime()
  const users = usePresence()
  return (
    <div className="flex items-center gap-2">
      <span role="status" className={status === 'open' ? 'sr-only' : 'text-xs text-amber-400'}>{STATUS_LABEL[status]}</span>
      <ul aria-label="Online agora" className="flex -space-x-2">
        {users.map((u) => (
          <li key={u.userId} className="group relative">
            <span aria-label={`${u.displayName} (${roleLabel(u.role)})`}
              className={`flex h-8 w-8 items-center justify-center rounded-full text-xs font-semibold ring-2 ring-zinc-900 ${u.role === 'gm' ? 'bg-amber-600 text-white' : 'bg-emerald-700 text-white'}`}>
              {initials(u.displayName)}
            </span>
            <Tip side="bottom">{u.displayName} · {roleLabel(u.role)}</Tip>
          </li>
        ))}
      </ul>
    </div>
  )
}
