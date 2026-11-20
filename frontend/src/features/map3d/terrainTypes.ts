export const TERRAIN_IDS = [
  'floor',
  'grass',
  'dirt',
  'sand',
  'water',
  'stone',
  'snow',
  'lava',
  'ruin',
] as const

export type TerrainTypeId = (typeof TERRAIN_IDS)[number]

export interface TerrainType {
  id: TerrainTypeId
  name: string
  color: string
  emissive?: string
  animated?: 'water' | 'lava'
  blocksMovement: boolean
  blocksVision: boolean
}

export const TERRAIN_TYPES: Record<TerrainTypeId, TerrainType> = {
  floor: { id: 'floor', name: 'Chão', color: '#8b8173', blocksMovement: false, blocksVision: false },
  grass: { id: 'grass', name: 'Grama', color: '#5a9e4b', blocksMovement: false, blocksVision: false },
  dirt: { id: 'dirt', name: 'Terra', color: '#8a5a3b', blocksMovement: false, blocksVision: false },
  sand: { id: 'sand', name: 'Areia', color: '#e0c98a', blocksMovement: false, blocksVision: false },
  water: {
    id: 'water',
    name: 'Água',
    color: '#3b82c4',
    emissive: '#0b2a55',
    animated: 'water',
    blocksMovement: false,
    blocksVision: false,
  },
  stone: { id: 'stone', name: 'Pedra', color: '#7d8590', blocksMovement: true, blocksVision: true },
  snow: { id: 'snow', name: 'Neve', color: '#eef4fa', blocksMovement: false, blocksVision: false },
  lava: {
    id: 'lava',
    name: 'Lava',
    color: '#e8531c',
    emissive: '#ff4a00',
    animated: 'lava',
    blocksMovement: false,
    blocksVision: false,
  },
  ruin: { id: 'ruin', name: 'Ruína', color: '#5f6368', blocksMovement: true, blocksVision: true },
}

export const TERRAIN_LIST: TerrainType[] = TERRAIN_IDS.map((id) => TERRAIN_TYPES[id])
