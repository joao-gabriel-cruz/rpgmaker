import { BatteryLow, Bed, Biohazard, Droplets, EyeOff, Ghost, ShieldOff, type LucideIcon } from 'lucide-react'
import { hasLightSource } from './lightSource'

export type ConditionId = 'wounded' | 'debilitated' | 'susceptible' | 'scared' | 'hidden' | 'infected' | 'unconscious'

export interface ConditionDef {
  id: ConditionId
  label: string
  icon: LucideIcon
  /** Tailwind text color class for the icon. */
  tone: string
  /** Players may toggle it on their own character (the Mestre can always). */
  ownerCan: boolean
}

export const CONDITIONS: ConditionDef[] = [
  { id: 'wounded', label: 'Ferido', icon: Droplets, tone: 'text-red-400', ownerCan: false },
  { id: 'debilitated', label: 'Debilitado', icon: BatteryLow, tone: 'text-yellow-400', ownerCan: false },
  { id: 'susceptible', label: 'Suscetível', icon: ShieldOff, tone: 'text-fuchsia-400', ownerCan: false },
  { id: 'scared', label: 'Assustado', icon: Ghost, tone: 'text-sky-300', ownerCan: false },
  { id: 'hidden', label: 'Escondido', icon: EyeOff, tone: 'text-zinc-300', ownerCan: true },
  { id: 'infected', label: 'Infectado', icon: Biohazard, tone: 'text-lime-400', ownerCan: false },
  { id: 'unconscious', label: 'Inconsciente', icon: Bed, tone: 'text-indigo-300', ownerCan: false },
]

export const conditionDef = (id: string) => CONDITIONS.find((c) => c.id === id)

/** The part of a sheet (or of a token's `status`) that conditions and light are read from. */
export interface SheetStatus {
  conditions?: string[]
  health?: { damage?: number; boxesPerLevel?: number }
  tugOfWar?: { determinationPoints?: number }
  instincts?: { potency?: number; resolution?: number }
  inventory?: { body?: { name?: string; icon?: string }[]; backpack?: { name?: string; icon?: string }[] }
}

// Six health levels: Debilitação is level 4 and Incapacitação level 5 (same rule as the sheet).
function healthLevel(s: SheetStatus): number | null {
  const damage = s.health?.damage
  if (damage === undefined) return null
  const boxes = s.health?.boxesPerLevel ?? (s.instincts ? 1 + (s.instincts.potency ?? 0) + (s.instincts.resolution ?? 0) : null)
  return boxes ? Math.min(5, Math.floor(Math.max(0, damage) / Math.max(1, boxes))) : null
}

/** Conditions derived from the sheet: Debilitado (level 4), Inconsciente/Incapacitado (level 5), Suscetível (Determinação 0). */
export function autoConditions(s: SheetStatus | null | undefined): ConditionId[] {
  if (!s) return []
  const level = healthLevel(s)
  const out: ConditionId[] = []
  if (level !== null && level >= 4) out.push('debilitated')
  if (level !== null && level >= 5) out.push('unconscious')
  if (s.tugOfWar && s.tugOfWar.determinationPoints === 0) out.push('susceptible')
  return out
}

/** Manual plus automatic conditions, deduplicated, in catalog order. */
export function conditionsOf(s: SheetStatus | null | undefined): ConditionId[] {
  const set = new Set<string>([...(s?.conditions ?? []), ...autoConditions(s)])
  return CONDITIONS.filter((c) => set.has(c.id)).map((c) => c.id)
}

export { hasLightSource }
