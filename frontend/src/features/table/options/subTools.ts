import {
  ArrowDownToLine, ArrowUpDown, ArrowUpFromLine, Eraser, MousePointerClick, PaintBucket, Paintbrush, Pointer, Trash2,
} from 'lucide-react'
import type { ObjectToolId, TerrainToolId } from '../../map3d/tools'
import type { SegOption } from './ui'

/** Number keys pick the sub-tool: the letters belong to the rail and the camera. */
export const TERRAIN_SUBTOOLS: SegOption<TerrainToolId>[] = [
  { id: 'paint', label: 'Pintar', icon: Paintbrush, keys: '1' },
  { id: 'raise', label: 'Subir', icon: ArrowUpFromLine, keys: '2' },
  { id: 'lower', label: 'Descer', icon: ArrowDownToLine, keys: '3' },
  { id: 'height', label: 'Definir altura', icon: ArrowUpDown, keys: '4' },
  { id: 'fill', label: 'Preencher', icon: PaintBucket, keys: '5' },
  { id: 'erase', label: 'Apagar', icon: Eraser, keys: '6' },
]

export const OBJECT_MODES: SegOption<ObjectToolId>[] = [
  { id: 'object', label: 'Colocar', text: 'Colocar', icon: MousePointerClick, keys: '1' },
  { id: 'select', label: 'Selecionar', text: 'Selecionar', icon: Pointer, keys: '2' },
  { id: 'removeObject', label: 'Remover', text: 'Remover', icon: Trash2, keys: '3' },
]

