import type { Sheet, InstinctKey, KnowledgeKey, PracticeKey } from '../sheet-assimilacao/types'

/** Simplified NPC sheet; shares the PC keys so the server can roll for NPCs. */
export interface NpcSheetData {
  identity: { name: string; portraitUrl: string | null }
  instincts: Record<InstinctKey, number>
  knowledge: Record<KnowledgeKey, number>
  practices: Record<PracticeKey, number>
  health: { damage: number; boxesPerLevel: number }
  threats: string[]
  notes: string
}

// The sync layer is typed on the PC sheet; NPC sheets only differ in shape, not in transport.
export const asNpc = (s: Sheet): NpcSheetData => s as unknown as NpcSheetData
export const asSheet = (s: NpcSheetData): Sheet => s as unknown as Sheet
