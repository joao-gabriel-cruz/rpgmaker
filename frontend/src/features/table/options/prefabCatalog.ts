import {
  BrickWall, Box, Car, Castle, Construction, DoorClosed, Fence, Flame, Lamp, Mountain, Package, Sprout, TreePine, Trees,
  type LucideIcon,
} from 'lucide-react'
import { PREFAB_META, type Category, type PrefabMeta } from '../../map3d/objects/catalogData'

/** Icon of each catalog prefab, used as the palette card thumbnail (no 3D render needed). */
export const PREFAB_ICONS: Record<string, LucideIcon> = {
  wall: BrickWall, barricade: Construction, fence: Fence, door: DoorClosed, watchtower: Castle, post: Lamp,
  tree: TreePine, bush: Sprout, rock: Box, volcano: Mountain, car: Car, crate: Package, campfire: Flame,
}

export const prefabIcon = (id: string): LucideIcon => PREFAB_ICONS[id] ?? Trees

export const CATEGORY_TINT: Record<Category, string> = {
  estrutura: 'text-sky-300',
  natureza: 'text-emerald-300',
  objeto: 'text-amber-300',
}

const plain = (s: string) => s.normalize('NFD').replace(/\p{M}/gu, '').toLowerCase().trim()

/** Catalog entries matching the search text (accent-insensitive) and the category, in catalog order. */
export function filterPrefabs(query: string, category: Category | 'all', list: PrefabMeta[] = PREFAB_META): PrefabMeta[] {
  const q = plain(query)
  return list.filter((p) => (category === 'all' || p.category === category) && plain(p.name).includes(q))
}
