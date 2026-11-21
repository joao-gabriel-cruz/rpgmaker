/** Local pub/sub for table effects (`fx:<animation>`, `fx:hit_react`, `ambience:*`...); nothing here touches the network. */
export interface FxPayload {
  x: number
  z: number
  target?: { kind: 'none' | 'cell' | 'token' | 'object'; x?: number; z?: number; tokenId?: string; objectId?: string }
  [extra: string]: unknown
}

export type FxHandler = (payload: FxPayload) => void

const handlers = new Map<string, Set<FxHandler>>()

export const fxBus = {
  emit(name: string, payload?: unknown): void {
    handlers.get(name)?.forEach((h) => h(payload as FxPayload))
  },
  /** Subscribes to one event name; returns the unsubscribe function. */
  on(name: string, handler: FxHandler): () => void {
    const set = handlers.get(name) ?? new Set<FxHandler>()
    set.add(handler)
    handlers.set(name, set)
    return () => {
      set.delete(handler)
    }
  },
}
