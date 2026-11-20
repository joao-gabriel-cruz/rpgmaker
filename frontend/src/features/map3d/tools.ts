import type { TerrainTypeId } from './terrainTypes'

export type ToolId = 'paint' | 'raise' | 'lower' | 'height' | 'fill' | 'erase'

export interface ToolState {
  tool: ToolId
  terrain: TerrainTypeId
  brushSize: 1 | 3 | 5
  height: number
}

export const TOOLS: { id: ToolId; label: string; key: string }[] = [
  { id: 'paint', label: 'Pincel', key: 'b' },
  { id: 'raise', label: 'Subir', key: 'u' },
  { id: 'lower', label: 'Descer', key: 'd' },
  { id: 'height', label: 'Definir altura', key: 'h' },
  { id: 'fill', label: 'Preencher', key: 'f' },
  { id: 'erase', label: 'Apagar', key: 'e' },
]

export const BRUSH_SIZES = [1, 3, 5] as const

export const DEFAULT_TOOL_STATE: ToolState = {
  tool: 'paint',
  terrain: 'grass',
  brushSize: 1,
  height: 2,
}
