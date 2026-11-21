import type { ReactNode } from 'react'
import { vi } from 'vitest'
import { RealtimeContext } from '../features/realtime/realtimeContext'
import type { CampaignSocketApi } from '../features/realtime/useCampaignSocket'

type Handler = (payload: unknown) => void

/** In-memory realtime double: `emit` delivers an event to subscribers, `send` is a spy that resolves. */
export function fakeRealtime() {
  const handlers = new Map<string, Set<Handler>>()
  const send = vi.fn<(type: string, payload?: unknown) => Promise<void>>().mockResolvedValue(undefined)
  const api: CampaignSocketApi = {
    status: 'open',
    send,
    subscribe: (type, h) => {
      const set = handlers.get(type) ?? new Set<Handler>()
      set.add(h as Handler)
      handlers.set(type, set)
      return () => { set.delete(h as Handler) }
    },
  }
  return {
    send,
    emit: (type: string, payload: unknown) => handlers.get(type)?.forEach((h) => h(payload)),
    Provider: ({ children }: { children: ReactNode }) => <RealtimeContext.Provider value={api}>{children}</RealtimeContext.Provider>,
  }
}
