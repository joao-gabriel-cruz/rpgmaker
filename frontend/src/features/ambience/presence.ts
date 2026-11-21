export const BUBBLE_MS = 6000
export const REACTION_MS = 2400
export const LOG_LIMIT = 100
export const MAX_SAY = 140
export const EMOJIS = ['👍', '😂', '😱', '❤️', '💀', '❓'] as const

/** A `chat.said` / `chat.reacted` event payload. */
export interface SaidEvent {
  characterId: string
  name: string
  text?: string
  emoji?: string
  at: number
}

export interface Bubble {
  id: number
  characterId: string
  name: string
  text: string
}

export interface Reaction {
  id: number
  characterId: string
  emoji: string
}

export interface LogEntry {
  id: number
  name: string
  text: string
  at: number
}

/** One bubble per character: a new line replaces the previous one. */
export const putBubble = (list: Bubble[], b: Bubble): Bubble[] => [...list.filter((o) => o.characterId !== b.characterId), b]

export const dropBubble = (list: Bubble[], id: number): Bubble[] => list.filter((b) => b.id !== id)

export const addLog = (list: LogEntry[], e: LogEntry): LogEntry[] => [...list, e].slice(-LOG_LIMIT)
