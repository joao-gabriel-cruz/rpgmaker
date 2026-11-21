import { useQuery } from '@tanstack/react-query'
import {
  Axe, Bandage, Compass, Crosshair, Drum, Eye, Flame, Footprints, Hammer, Hand, HeartPulse, Key, MessageCircle, Moon, Music,
  Search, Shield, Skull, Smile, Sparkles, Star, Swords, Target, Wrench, Zap, type LucideIcon,
} from 'lucide-react'
import { api } from '../../lib/api'
import type { AptitudeKey, InstinctKey, Sheet } from '../sheet-assimilacao/types'

export type ActionCategory = 'combat' | 'exploration' | 'social' | 'survival'
export type ActionTargetKind = 'none' | 'cell' | 'token' | 'object'
export type ActionOutcome = 'none' | 'damage' | 'heal' | 'loot' | 'rest' | 'morale' | 'notice'

export const ANIMATIONS = ['melee_strike', 'ranged_shot', 'look_around', 'search', 'sneak', 'repair', 'first_aid', 'rest', 'dance', 'persuade', 'intimidate'] as const
export type Animation = (typeof ANIMATIONS)[number]

export interface ActionRoll {
  instinct: InstinctKey
  aptitude?: AptitudeKey
  label: string
}

/** A catalog entry as served by `GET /campaigns/{id}/actions`. */
export interface GameAction {
  id: string
  name: string
  icon: string
  category: ActionCategory
  target: ActionTargetKind
  rolls: ActionRoll[]
  animation: Animation
  outcome: ActionOutcome
  range?: number
  enabled: boolean
  builtin: boolean
}

/** Where an action is aimed, as sent in `action.perform` and echoed by `action.started`. */
export interface ActionTarget {
  kind: ActionTargetKind
  x?: number
  z?: number
  tokenId?: string
  objectId?: string
}

export const CATEGORY_LABELS: Record<ActionCategory, string> = {
  combat: 'Combate', exploration: 'Exploração', social: 'Social', survival: 'Sobrevivência',
}
export const CATEGORY_ORDER: ActionCategory[] = ['combat', 'exploration', 'survival', 'social']
export const TARGET_LABELS: Record<ActionTargetKind, string> = { none: 'Sem alvo', cell: 'Célula', token: 'Personagem', object: 'Objeto ou célula' }
export const OUTCOME_LABELS: Record<ActionOutcome, string> = {
  none: 'Nenhuma', damage: 'Sugerir dano', heal: 'Sugerir cura', loot: 'Revelar itens escondidos', rest: 'Teste de Descanso',
  morale: 'Sugerir Moral do Refúgio', notice: 'Aviso',
}
export const ANIMATION_LABELS: Record<Animation, string> = {
  melee_strike: 'Golpe corpo a corpo', ranged_shot: 'Disparo', look_around: 'Olhar em volta', search: 'Vasculhar', sneak: 'Esgueirar',
  repair: 'Consertar', first_aid: 'Primeiros socorros', rest: 'Descansar', dance: 'Dançar', persuade: 'Convencer', intimidate: 'Intimidar',
}

export const ACTION_ICONS: Record<string, LucideIcon> = {
  swords: Swords, crosshair: Crosshair, eye: Eye, search: Search, footprints: Footprints, hammer: Hammer, 'heart-pulse': HeartPulse,
  moon: Moon, music: Music, 'message-circle': MessageCircle, skull: Skull, flame: Flame, shield: Shield, wrench: Wrench,
  sparkles: Sparkles, hand: Hand, smile: Smile, zap: Zap, star: Star, key: Key, compass: Compass, drum: Drum, target: Target,
  axe: Axe, bandage: Bandage,
}
export const iconOf = (name: string): LucideIcon => ACTION_ICONS[name] ?? Zap

export const actionsKey = (campaignId: string) => ['campaigns', campaignId, 'actions']

/** The campaign's action catalog (the Mestre also gets the disabled ones). */
export function useActions(campaignId: string) {
  const query = useQuery({
    queryKey: actionsKey(campaignId),
    queryFn: () => api.get<{ actions: GameAction[] }>(`/campaigns/${campaignId}/actions`),
  })
  const list = query.data?.actions
  return { actions: Array.isArray(list) ? list : [], loading: query.isPending }
}

interface SheetItem {
  kind?: string
  attackAnimation?: string
  effects?: { keptBonus?: { actionCategory: string; amount: number } }
}

const bodyItems = (sheet: Sheet): SheetItem[] => (sheet.inventory?.body ?? []) as unknown as SheetItem[]

/** Extra kept dice from equipped (body) items for the action's category; mirrors the server (max, capped at 2). */
export function keptBonusFor(sheet: Sheet, category: ActionCategory): number {
  const best = bodyItems(sheet).reduce((n, it) => {
    const b = it.effects?.keptBonus
    return b && b.actionCategory === category ? Math.max(n, b.amount) : n
  }, 0)
  return Math.min(best, 2)
}

/** The item behind the bonus, for the composer hint. */
export function bonusSource(sheet: Sheet, category: ActionCategory): string {
  const it = bodyItems(sheet).find((i) => i.effects?.keptBonus?.actionCategory === category) as (SheetItem & { name?: string }) | undefined
  return it?.name ?? ''
}
