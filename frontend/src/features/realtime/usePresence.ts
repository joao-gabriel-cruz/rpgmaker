import type { Role } from '../campaigns/campaigns'
import { useEffect, useState } from 'react'
import { useRealtime } from './realtimeContext'

export interface OnlineUser {
  userId: string
  displayName: string
  role: Role
}

/** Users currently online in the campaign (server event presence.updated). */
export function usePresence(): OnlineUser[] {
  const { subscribe } = useRealtime()
  const [users, setUsers] = useState<OnlineUser[]>([])
  useEffect(() => subscribe('presence.updated', (p) => setUsers((p as { users: OnlineUser[] }).users)), [subscribe])
  return users
}
