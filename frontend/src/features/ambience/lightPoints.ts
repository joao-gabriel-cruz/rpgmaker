import { getCell, surfaceHeight, type MapGrid } from '../map3d/grid'
import type { MapObject } from '../map3d/objects/model'
import type { Token } from '../tokens/tokens'
import { hasLightSource } from './lightSource'

export interface LightPoint {
  key: string
  x: number
  y: number
  z: number
  /** Campfires burn bigger and redder than a carried torch. */
  fire: boolean
}

/** Most lights drawn at once: each one costs every lit fragment, so the cap keeps low-end GPUs happy. */
export const MAX_LIGHTS = 8

/** Prefabs that glow at night like a campfire, with the light's height above the ground. */
const LIGHT_HEIGHT: Record<string, number> = { campfire: 0.8, lamp_post: 2.9 }

/** World positions of torch bearers (inventory light source) and campfires, capped at MAX_LIGHTS. */
export function lightPoints(grid: MapGrid, tokens: readonly Token[], objects: readonly MapObject[]): LightPoint[] {
  const at = (x: number, z: number) => surfaceHeight(getCell(grid, x, z)?.height ?? 1)
  const fires = objects.filter((o) => o.prefab in LIGHT_HEIGHT)
    .map((o) => ({ key: `o${o.id}`, x: o.x + 0.5, y: at(o.x, o.z) + LIGHT_HEIGHT[o.prefab], z: o.z + 0.5, fire: true }))
  const bearers = tokens.filter((t) => hasLightSource(t.status))
    .map((t) => ({ key: `t${t.id}`, x: t.x + 0.5, y: at(t.x, t.z) + 1.3, z: t.z + 0.5, fire: false }))
  return [...fires, ...bearers].slice(0, MAX_LIGHTS)
}
