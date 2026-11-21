import type { Rotation } from './objects/model'
import type { TerrainTypeId } from './terrainTypes'

export type TerrainToolId = 'paint' | 'raise' | 'lower' | 'height' | 'fill' | 'erase'
export type ObjectToolId = 'object' | 'wall' | 'select' | 'removeObject'
export type ToolId = TerrainToolId | ObjectToolId

export interface ToolState {
  tool: ToolId
  terrain: TerrainTypeId
  brushSize: 1 | 3 | 5
  height: number
  prefab: string
  variant: 'pine' | 'round'
  rotation: Rotation
}

interface ToolDef<T extends ToolId> {
  id: T
  label: string
  key: string
}

export const TERRAIN_TOOLS: ToolDef<TerrainToolId>[] = [
  { id: 'paint', label: 'Pincel', key: 'b' },
  { id: 'raise', label: 'Subir', key: 'u' },
  { id: 'lower', label: 'Descer', key: 'd' },
  { id: 'height', label: 'Definir altura', key: 'h' },
  { id: 'fill', label: 'Preencher', key: 'f' },
  { id: 'erase', label: 'Apagar', key: 'e' },
]

export const OBJECT_TOOLS: ToolDef<ObjectToolId>[] = [
  { id: 'object', label: 'Objetos', key: 'o' },
  { id: 'wall', label: 'Muro', key: 'm' },
  { id: 'select', label: 'Selecionar objeto', key: 'v' },
  { id: 'removeObject', label: 'Remover objeto', key: 'x' },
]

export const TOOLS: ToolDef<ToolId>[] = [...TERRAIN_TOOLS, ...OBJECT_TOOLS]

export const isObjectTool = (t: ToolId): t is ObjectToolId => OBJECT_TOOLS.some((o) => o.id === t)

export const BRUSH_SIZES = [1, 3, 5] as const

export const DEFAULT_TOOL_STATE: ToolState = {
  tool: 'paint',
  terrain: 'grass',
  brushSize: 1,
  height: 2,
  prefab: 'tree',
  variant: 'pine',
  rotation: 0,
}
