import { PREFAB_META, type PrefabMeta } from './catalogData'
import { BigRock, Bush, Tree } from './nature'
import { Volcano } from './Volcano'
import { Campfire, Car, Crate } from './props'
import { Barricade, Door, Fence, Post, Wall, Watchtower } from './structures'
import type { PrefabComponent } from './types'

export interface Prefab extends PrefabMeta {
  Component: PrefabComponent
}

const COMPONENTS: Record<string, PrefabComponent> = {
  wall: Wall,
  barricade: Barricade,
  fence: Fence,
  door: Door,
  watchtower: Watchtower,
  post: Post,
  tree: Tree,
  bush: Bush,
  rock: BigRock,
  volcano: Volcano,
  car: Car,
  crate: Crate,
  campfire: Campfire,
}

export const CATALOG: Prefab[] = PREFAB_META.map((m) => ({ ...m, Component: COMPONENTS[m.id] }))

export const PREFABS: Record<string, Prefab> = Object.fromEntries(CATALOG.map((p) => [p.id, p]))
