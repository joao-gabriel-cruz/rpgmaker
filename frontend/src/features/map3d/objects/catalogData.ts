export type Category = 'estrutura' | 'natureza' | 'objeto'
export type ObjectProps = Record<string, string | number | boolean>

export interface Rules {
  blocksMovement: boolean
  blocksVision: boolean
}

export interface PrefabMeta {
  id: string
  name: string
  category: Category
  /** Footprint in cells at rotation 0 (x extent, z extent). */
  width: number
  depth: number
  /** Height in blocks. */
  height: number
  rules: Rules
  defaultProps?: ObjectProps
}

export const CATEGORY_LABELS: Record<Category, string> = {
  estrutura: 'Estruturas',
  natureza: 'Natureza',
  objeto: 'Objetos',
}

const BOTH: Rules = { blocksMovement: true, blocksVision: true }
const MOVE: Rules = { blocksMovement: true, blocksVision: false }
const NONE: Rules = { blocksMovement: false, blocksVision: false }

const m = (
  id: string,
  name: string,
  category: Category,
  [width, depth, height]: [number, number, number],
  rules: Rules,
  defaultProps?: ObjectProps,
): PrefabMeta => ({ id, name, category, width, depth, height, rules, defaultProps })

export const PREFAB_META: PrefabMeta[] = [
  m('wall', 'Muro', 'estrutura', [1, 1, 3], BOTH),
  m('barricade', 'Barricada', 'estrutura', [1, 1, 1], MOVE),
  m('fence', 'Cerca', 'estrutura', [1, 1, 1], MOVE),
  m('door', 'Porta', 'estrutura', [1, 1, 3], BOTH, { open: false }),
  m('watchtower', 'Torre de vigia', 'estrutura', [1, 1, 5], MOVE),
  m('post', 'Poste', 'estrutura', [1, 1, 2], NONE),
  m('tree', 'Árvore', 'natureza', [1, 1, 3], BOTH, { variant: 'pine' }),
  m('bush', 'Arbusto', 'natureza', [1, 1, 1], NONE),
  m('rock', 'Rocha grande', 'natureza', [1, 1, 1], BOTH),
  m('volcano', 'Vulcão', 'natureza', [3, 3, 3], BOTH),
  m('car', 'Carro abandonado', 'objeto', [2, 1, 1], BOTH),
  m('crate', 'Caixote', 'objeto', [1, 1, 1], BOTH),
  m('campfire', 'Fogueira', 'objeto', [1, 1, 1], NONE),
  m('tent', 'Barraca', 'objeto', [2, 2, 2], MOVE),
  m('barrel', 'Barril', 'objeto', [1, 1, 1], NONE),
  m('bench', 'Banco', 'objeto', [2, 1, 1], MOVE),
  m('market_stall', 'Barraca de feira', 'objeto', [2, 2, 2], NONE, { color: 'red' }),
  m('lamp_post', 'Poste de luz', 'estrutura', [1, 1, 3], NONE),
  m('sandbags', 'Sacos de areia', 'objeto', [1, 1, 1], MOVE),
  m('ruined_wall', 'Muro quebrado', 'estrutura', [1, 1, 2], BOTH, { height: 1 }),
  m('well', 'Poço', 'objeto', [1, 1, 1], MOVE),
  m('sign', 'Placa', 'objeto', [1, 1, 1], NONE),
  m('bridge', 'Ponte', 'estrutura', [1, 1, 1], NONE),
  m('bunting', 'Bandeirolas', 'estrutura', [2, 1, 2], NONE),
]

export const PREFAB_META_BY_ID: Record<string, PrefabMeta> = Object.fromEntries(
  PREFAB_META.map((p) => [p.id, p]),
)

// An open door lets everything through; every other prefab uses its default rules.
export function rulesFor(prefab: string, props: ObjectProps): Rules {
  const meta = PREFAB_META_BY_ID[prefab]
  if (!meta) return NONE
  if (prefab === 'door' && props.open === true) return NONE
  return meta.rules
}
