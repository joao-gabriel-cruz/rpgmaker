export interface Objective {
  id: string
  title: string
  kind: 'main' | 'secondary'
  scope: 'collective' | 'individual'
  characterId?: string
  target: number
  progress: number
}

export interface ThreatPool { d6: number; d10: number; d12: number }

/** `pool` and `notes` only come to the Mestre. */
export interface Threat {
  id: string
  name: string
  pool?: ThreatPool
  npcCharacterId?: string
  notes?: string
}

export type ActivationSymbol = 'success' | 'adaptation' | 'pressure'

/** Only sent to the Mestre. `source` is "players" or "threat:<id>". */
export interface Activation {
  id: string
  symbol: ActivationSymbol
  threshold: number
  text: string
  source: string
  count?: number
}

export interface Conflict {
  id: string
  sceneId: string | null
  title: string
  conditions: string[]
  objectives: Objective[]
  threats: Threat[]
  activations?: Activation[]
}

export interface Triggered {
  activationId: string
  symbol: ActivationSymbol
  threshold: number
  text: string
  times: number
}

export const SYMBOL_LABELS: Record<ActivationSymbol, string> = { success: 'Sucesso', adaptation: 'Adaptação', pressure: 'Pressão' }

export const progressPct = (o: Pick<Objective, 'progress' | 'target'>) =>
  Math.min(100, Math.round((o.progress / Math.max(1, o.target)) * 100))

export const isComplete = (o: Pick<Objective, 'progress' | 'target'>) => o.progress >= o.target

export const newId = () => Math.random().toString(36).slice(2, 10)

export const emptyConflict = (): Conflict => ({
  id: '', sceneId: null, title: '', conditions: [], objectives: [], threats: [], activations: [],
})
