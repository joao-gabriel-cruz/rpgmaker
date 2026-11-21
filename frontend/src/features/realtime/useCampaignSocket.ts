import { useCallback, useEffect, useSyncExternalStore } from 'react'
import { CampaignSocket, type ConnectionStatus, type Handler, type SocketOptions } from './CampaignSocket'

const sockets = new Map<string, { socket: CampaignSocket; refs: number }>()

function wsUrl(campaignId: string) {
  const scheme = location.protocol === 'https:' ? 'wss' : 'ws'
  return `${scheme}://${location.host}/api/campaigns/${campaignId}/ws`
}

/** Test seam: lets tests inject a fake WebSocket factory. */
export const socketFactory: { create: SocketOptions['createSocket'] } = { create: undefined }

// Created lazily (no connection yet) so child effects can subscribe before the owner opens it.
function entryFor(campaignId: string) {
  let entry = sockets.get(campaignId)
  if (!entry) {
    entry = { socket: new CampaignSocket({ url: wsUrl(campaignId), createSocket: socketFactory.create }), refs: 0 }
    sockets.set(campaignId, entry)
  }
  return entry
}

/** Drops cached sockets (tests only). */
export function resetSockets() {
  sockets.forEach((e) => e.socket.close())
  sockets.clear()
}

export interface CampaignSocketApi {
  status: ConnectionStatus
  /** Sends a command; resolves on ack, rejects with the server's PT-BR error. */
  send: (type: string, payload?: unknown) => Promise<void>
  /** Subscribes to an event type; returns the unsubscribe function. */
  subscribe: (type: string, handler: Handler) => () => void
}

/** One shared (ref-counted) socket per campaign; every caller for the same id reuses it. */
export function useCampaignSocket(campaignId: string): CampaignSocketApi {
  const { socket } = entryFor(campaignId)
  useEffect(() => {
    const entry = entryFor(campaignId)
    entry.refs += 1
    entry.socket.open()
    return () => {
      if (--entry.refs === 0) entry.socket.close()
    }
  }, [campaignId])

  const status = useSyncExternalStore(
    (cb) => socket.onStatusChange(cb),
    () => socket.status,
  )
  const send = useCallback((type: string, payload?: unknown) => socket.send(type, payload), [socket])
  const subscribe = useCallback((type: string, handler: Handler) => socket.subscribe(type, handler), [socket])
  return { status, send, subscribe }
}
