type Handler = (payload: unknown) => void

const handlers = new Map<string, Set<Handler>>()

/** Local event bus for visual/sound effects (`fx:<animation>`, `ambience:*`); nothing here touches the network. */
export const fxBus = {
  emit(name: string, payload?: unknown): void {
    handlers.get(name)?.forEach((h) => h(payload))
  },
  on(name: string, handler: Handler): () => void {
    const set = handlers.get(name) ?? new Set<Handler>()
    set.add(handler)
    handlers.set(name, set)
    return () => {
      set.delete(handler)
    }
  },
}
