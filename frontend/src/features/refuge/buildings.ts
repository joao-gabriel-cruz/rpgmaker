import type { ResourceKey } from './refuge'

export type BuildingState = 'planned' | 'building' | 'done'

export interface Building {
  id: string
  kind: string
  level: number
  built: number
  state: BuildingState
  progress: number
  paid: boolean
  workers: number
}

export interface Proposal { id: string; kind: string; userId: string; by: string }

export interface CrisisTrigger { kind: 'overpopulation' | 'breach' | 'shortage'; detail: string; resource?: string }

export interface WorkLine {
  buildingId: string; name: string; level: number; points: number; progress: number; cost: number; done: boolean; blocked?: string
}

export type Flow = Partial<Record<ResourceKey, number>>

export interface WeekReport {
  week: number
  entered: Flow
  spent: Flow
  missing: Flow
  lost: Flow
  adjustments: Flow
  works: WorkLine[]
  completed: string[]
  triggers: CrisisTrigger[]
  notes: string[]
}

export interface BuildingDef { kind: string; name: string; effect: string; cost: number; leveled?: boolean }

/** Mirrors backend catalog_buildings.go; [casa] marks house rules. */
export const CATALOG: BuildingDef[] = [
  { kind: 'apothecary', name: 'Boticário', cost: 10, effect: 'Converte Plantas em Remédios (1 por trabalhador)' },
  { kind: 'tailor', name: 'Casa de Costura', cost: 10, effect: '[casa] Converte Plantas em Vestuário; +1 Vestuário a cada Vestuário produzido' },
  { kind: 'community', name: 'Centro Comunitário', cost: 10, leveled: true, effect: '+1 Moral por nível' },
  { kind: 'biomassGathering', name: 'Coleta de Biomassa', cost: 20, effect: 'Produz Biomassa conforme a abundância da região' },
  { kind: 'dormitory', name: 'Dormitório', cost: 20, effect: '[casa] +5 no teto de População' },
  { kind: 'ammoFactory', name: 'Fábrica de Munição', cost: 20, effect: 'Converte Minerais em Munição' },
  { kind: 'worksFactory', name: 'Fábrica para Obras', cost: 20, effect: 'Converte Minerais em Material de Construção' },
  { kind: 'waterSource', name: 'Fonte de Água', cost: 10, effect: 'A água não é mais consumida do estoque; com trabalhadores, produz Água' },
  { kind: 'fortification', name: 'Fortificação', cost: 10, leveled: true, effect: '+1 Defesa por nível; exige 1 População alocada como guarda' },
  { kind: 'mill', name: 'Moinho', cost: 10, effect: '[casa] Converte Plantas em Alimento' },
  { kind: 'equipmentWorkshop', name: 'Oficina de Equipamentos', cost: 10, leveled: true, effect: '[casa] Permite fabricar itens com Escassez até o nível (o Mestre usa o compêndio)' },
  { kind: 'mechanicWorkshop', name: 'Oficina Mecânica', cost: 10, effect: 'Restaura veículos (efeito narrativo)' },
  { kind: 'quarry', name: 'Pedreira', cost: 20, effect: 'Produz Minerais conforme a abundância da região' },
  { kind: 'port', name: 'Porto', cost: 20, effect: '[casa] +1 Mobilidade; manutenção a cada 26 semanas: 5 Madeira' },
  { kind: 'canteen', name: 'Refeitório', cost: 10, effect: 'Dobra o Alimento produzido' },
  { kind: 'refineryAnimal', name: 'Refinaria Animal', cost: 10, effect: 'Converte Animais em Combustível (até 10 por semana)' },
  { kind: 'refineryBiomass', name: 'Refinaria de Biomassa', cost: 10, effect: 'Converte Biomassa em Combustível (até 10 por semana)' },
  { kind: 'refineryVegetal', name: 'Refinaria Vegetal', cost: 10, effect: 'Converte Plantas em Combustível (até 10 por semana)' },
  { kind: 'expandedReserves', name: 'Reservas Expandidas', cost: 10, effect: '[casa] +10 no teto de Reservas' },
  { kind: 'sawmill', name: 'Serraria', cost: 10, effect: 'Converte Madeira em Material de Construção' },
]

export const STATE_LABEL: Record<BuildingState, string> = { planned: 'Planejada', building: 'Em obra', done: 'Concluída' }

export const defOf = (kind: string) => CATALOG.find((d) => d.kind === kind)

export const costAt = (d: BuildingDef, level: number) => (d.leveled ? d.cost * level : d.cost)

/** Material de Construção paid when work on a level starts [casa]. */
export const materialCost = (d: BuildingDef, level: number) => Math.floor(costAt(d, level) / 10)

export const allocationOf = (bs: Building[]): Record<string, number> => Object.fromEntries(bs.map((b) => [b.id, b.workers]))

export const totalOf = (alloc: Record<string, number>) => Object.values(alloc).reduce((a, b) => a + b, 0)

/** Kinds that already exist or wait for approval cannot be created again. */
export const takenKinds = (bs: Building[], ps: Proposal[]) => new Set([...bs.map((b) => b.kind), ...ps.map((p) => p.kind)])
