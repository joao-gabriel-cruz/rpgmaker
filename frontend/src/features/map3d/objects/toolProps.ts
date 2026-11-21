import type { ToolState } from '../tools'
import type { ObjectProps } from './catalogData'

/** Props a new object gets from the tool settings (variant, awning color, wall height). */
export function toolProps(state: Pick<ToolState, 'prefab' | 'variant' | 'color' | 'wallHeight'>): ObjectProps | undefined {
  if (state.prefab === 'tree') return { variant: state.variant }
  if (state.prefab === 'market_stall' && state.color) return { color: state.color }
  if (state.prefab === 'ruined_wall' && state.wallHeight) return { height: state.wallHeight }
  return undefined
}

export const AWNING_OPTIONS = [
  { id: 'red', label: 'Vermelho', short: 'Verm.' }, { id: 'blue', label: 'Azul', short: 'Azul' }, { id: 'green', label: 'Verde', short: 'Verde' },
  { id: 'yellow', label: 'Amarelo', short: 'Amar.' }, { id: 'orange', label: 'Laranja', short: 'Lar.' }, { id: 'purple', label: 'Roxo', short: 'Roxo' },
]
export const WALL_HEIGHT_OPTIONS = [{ id: '1', label: 'Baixo (1)', short: 'Baixo (1)' }, { id: '2', label: 'Alto (2)', short: 'Alto (2)' }]
