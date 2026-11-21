import { roleLabel } from '../campaigns/campaigns'
import { useRealtime } from './realtimeContext'
import { usePresence } from './usePresence'
import type { ConnectionStatus } from './CampaignSocket'

const STATUS_LABEL: Record<ConnectionStatus, string> = {
  connecting: 'Conectando…',
  open: 'Conectado',
  reconnecting: 'Reconectando…',
  closed: 'Desconectado',
}

export function OnlinePanel() {
  const { status } = useRealtime()
  const users = usePresence()
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-3">
        <h2 className="text-xl font-semibold">Online agora</h2>
        <span role="status" className={`text-sm ${status === 'open' ? 'text-emerald-400' : 'text-amber-400'}`}>
          {STATUS_LABEL[status]}
        </span>
      </div>
      <ul className="flex max-w-xl flex-col gap-2">
        {users.map((u) => (
          <li key={u.userId} className="flex items-center gap-3 rounded bg-zinc-900 px-4 py-2">
            <span aria-hidden className="h-2.5 w-2.5 rounded-full bg-emerald-500" />
            <span className="font-medium">{u.displayName}</span>
            <span className="text-sm text-zinc-400">{roleLabel(u.role)}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}
