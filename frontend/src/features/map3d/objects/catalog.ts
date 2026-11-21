import { PREFAB_META, type PrefabMeta } from './catalogData'
import { BigRock, Bush, Tree } from './nature'
import { Volcano } from './Volcano'
import { Bridge } from './bridgeModel'
import { Barrel, Bench, MarketStall, Sandbags, Sign, Tent, Well } from './camp'
import { Campfire, Car, Crate } from './props'
import { Bunting, LampPost, RuinedWall } from './street'
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
  tent: Tent,
  barrel: Barrel,
  bench: Bench,
  market_stall: MarketStall,
  lamp_post: LampPost,
  sandbags: Sandbags,
  ruined_wall: RuinedWall,
  well: Well,
  sign: Sign,
  bridge: Bridge,
  bunting: Bunting,
}

export const CATALOG: Prefab[] = PREFAB_META.map((m) => ({ ...m, Component: COMPONENTS[m.id] }))

export const PREFABS: Record<string, Prefab> = Object.fromEntries(CATALOG.map((p) => [p.id, p]))
