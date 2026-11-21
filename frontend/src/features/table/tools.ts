import {
  BrickWall, CloudFog, MousePointer2, Mountain, Pencil, Radio, Ruler, TreePine, UsersRound,
  type LucideIcon,
} from 'lucide-react'
import type { DrawMode } from '../drawings/drawings'

export type TableToolId = 'move' | 'ping' | 'ruler' | 'draw' | 'terrain' | 'objects' | 'wall' | 'fog' | 'tokens'

export interface TableToolDef {
  id: TableToolId
  label: string
  hint: string
  key: string
  gmOnly: boolean
  icon: LucideIcon
}

export const TABLE_TOOLS: TableToolDef[] = [
  { id: 'move', label: 'Mover', hint: 'Arraste tokens e gire a câmera', key: 'v', gmOnly: false, icon: MousePointer2 },
  { id: 'ping', label: 'Ping', hint: 'Clique para chamar a atenção de todos', key: 'p', gmOnly: false, icon: Radio },
  { id: 'ruler', label: 'Régua', hint: 'Arraste para medir distâncias', key: 'r', gmOnly: false, icon: Ruler },
  { id: 'draw', label: 'Desenhar', hint: 'Desenhe à mão livre no mapa', key: 'd', gmOnly: false, icon: Pencil },
  { id: 'terrain', label: 'Terreno', hint: 'Pinte e eleve o terreno', key: 't', gmOnly: true, icon: Mountain },
  { id: 'objects', label: 'Objetos', hint: 'Posicione objetos no mapa', key: 'o', gmOnly: true, icon: TreePine },
  { id: 'wall', label: 'Muro', hint: 'Desenhe muros em linha', key: 'm', gmOnly: true, icon: BrickWall },
  { id: 'fog', label: 'Névoa', hint: 'Revele ou esconda áreas', key: 'f', gmOnly: true, icon: CloudFog },
  { id: 'tokens', label: 'Tokens', hint: 'Coloque personagens na cena', key: 'k', gmOnly: true, icon: UsersRound },
]

export const toolsFor = (isGm: boolean) => TABLE_TOOLS.filter((t) => isGm || !t.gmOnly)

export const toolByKey = (key: string, isGm: boolean) => toolsFor(isGm).find((t) => t.key === key)

export const toolDef = (id: TableToolId) => TABLE_TOOLS.find((t) => t.id === id) as TableToolDef

/** Tools whose clicks reach the terrain editor (paint / objects / walls). */
export const isEditorTool = (t: TableToolId) => t === 'terrain' || t === 'objects' || t === 'wall'

export const drawModeOf = (t: TableToolId): DrawMode => (t === 'ruler' ? 'ruler' : t === 'draw' ? 'draw' : 'none')

/** Tools that show something in the floating options panel. */
export const hasOptions = (t: TableToolId) => t === 'draw' || t === 'fog' || t === 'tokens' || isEditorTool(t)
