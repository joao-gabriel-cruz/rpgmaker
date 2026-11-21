export type Trigger = 'on_enter' | 'on_turn_start' | 'item_use'
export type EffectKind = 'damage' | 'difficulty' | 'notice' | 'heal'

export interface Rule {
  trigger: Trigger
  effect: EffectKind
  amount: number
  message: string
}

export interface RuleSet {
  kind: 'terrain' | 'prefab'
  key: string
  label: string
  rules: Rule[]
  isDefault: boolean
  blocksMovement: boolean
  blocksVision: boolean
}

export interface Suggestion {
  id: string
  campaignId: string
  characterId: string
  characterName: string
  sceneId: string
  x: number
  z: number
  source: string
  trigger: Trigger
  effect: EffectKind
  amount: number
  message: string
  createdAt: string
}

export const TRIGGER_LABEL: Record<Trigger, string> = { on_enter: 'Ao entrar', on_turn_start: 'Ao começar o turno', item_use: 'Item usado' }
export const EFFECT_LABEL: Record<EffectKind, string> = { damage: 'Dano', difficulty: 'Dificuldade', notice: 'Aviso', heal: 'Cura' }

export function describe(s: Pick<Suggestion, 'effect' | 'amount'>): string {
  if (s.effect === 'damage') return `+${s.amount} de dano`
  if (s.effect === 'heal') return `curar ${s.amount}`
  if (s.effect === 'difficulty') return `+${s.amount}d12 de dificuldade na próxima rolagem`
  return 'Aviso'
}
