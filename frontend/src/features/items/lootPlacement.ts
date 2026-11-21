import { useSyncExternalStore } from 'react'

export interface Placement {
  itemId: string
  name: string
  difficulty: number
}

let current: Placement | null = null
const listeners = new Set<() => void>()
const emit = () => listeners.forEach((l) => l())

/** The Mestre picked an item to hide: the next map click places it. Shared by the Itens tab and the loot layer. */
export const lootPlacement = {
  start(p: Placement) {
    current = p
    emit()
  },
  cancel() {
    current = null
    emit()
  },
  get: () => current,
  subscribe(l: () => void) {
    listeners.add(l)
    return () => listeners.delete(l)
  },
}

export const usePlacement = () => useSyncExternalStore(lootPlacement.subscribe, lootPlacement.get)
