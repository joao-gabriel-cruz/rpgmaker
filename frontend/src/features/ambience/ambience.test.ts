import { expect, it } from 'vitest'
import { createGrid } from '../map3d/grid'
import type { MapObject } from '../map3d/objects/model'
import { applyCharacter, type Token } from '../tokens/tokens'
import { atmosphereFor, isAmbience, LIGHT_MIN_DARKNESS } from './ambience'
import { autoConditions, conditionsOf } from './conditions'
import { lightPoints, MAX_LIGHTS } from './lightPoints'
import { hasLightSource } from './lightSource'

const item = (name: string, icon?: string) => ({ name, icon })

it('applies lighting parameters per time of day: day keeps the original look, night is dim, blue and dark', () => {
  const day = atmosphereFor('day', 'clear')
  expect(day).toMatchObject({ sky: '#18181b', ambientIntensity: 0.8, sunIntensity: 1.6, darkness: 0 })
  const order = (['day', 'dawn', 'dusk', 'night'] as const).map((t) => atmosphereFor(t, 'clear'))
  const intensities = order.map((p) => p.ambientIntensity)
  expect(intensities).toEqual([...intensities].sort((a, b) => b - a))
  expect(order.map((p) => p.darkness)).toEqual([0, 0.25, 0.55, 1])
  expect(order[3].darkness).toBeGreaterThanOrEqual(LIGHT_MIN_DARKNESS)
  expect(atmosphereFor('night', 'clear').ambientIntensity).toBeGreaterThan(0.3)
})

it('weather pulls fog closer and dims the light; clear weather keeps fog out of reach', () => {
  const clear = atmosphereFor('day', 'clear')
  for (const w of ['rain', 'storm', 'fog', 'snow', 'ash'] as const) expect(atmosphereFor('day', w).fogFar).toBeLessThan(clear.fogFar)
  expect(atmosphereFor('day', 'fog').fogFar).toBeLessThan(atmosphereFor('day', 'rain').fogFar)
  expect(atmosphereFor('day', 'storm').sunIntensity).toBeLessThan(clear.sunIntensity)
})

it('validates ambience payloads', () => {
  expect(isAmbience({ sceneId: 's', timeOfDay: 'night', weather: 'rain' })).toBe(true)
  expect(isAmbience({ id: 's', name: 'Cena' })).toBe(false)
  expect(isAmbience(null)).toBe(false)
})

it('hasLightSource: lantern/torch by name (whole word) or flashlight/flame icon, body or backpack', () => {
  expect(hasLightSource({ inventory: { body: [item('Lanterna a óleo')], backpack: [] } })).toBe(true)
  expect(hasLightSource({ inventory: { body: [], backpack: [item('tocha improvisada')] } })).toBe(true)
  expect(hasLightSource({ inventory: { body: [item('Isqueiro', 'flame')], backpack: [] } })).toBe(true)
  expect(hasLightSource({ inventory: { body: [item('Lente', 'Flashlight')], backpack: [] } })).toBe(true)
  expect(hasLightSource({ inventory: { body: [item('Facão', 'sword')], backpack: [item('Tochas de papel?')] } })).toBe(true)
  expect(hasLightSource({ inventory: { body: [item('Facão', 'sword')], backpack: [item('Lanternagem')] } })).toBe(false)
  expect(hasLightSource({ inventory: { body: [], backpack: [] } })).toBe(false)
  expect(hasLightSource(null)).toBe(false)
  expect(hasLightSource({})).toBe(false)
})

it('derives automatic conditions from the sheet and merges manual ones in catalog order', () => {
  const pc = (damage: number, det = 5) => ({ health: { damage }, instincts: { potency: 1, resolution: 1 }, tugOfWar: { determinationPoints: det } })
  expect(autoConditions(pc(0))).toEqual([])
  expect(autoConditions(pc(11))).toEqual([]) // 3 boxes per level: level 3
  expect(autoConditions(pc(12))).toEqual(['debilitated'])
  expect(autoConditions(pc(15))).toEqual(['debilitated', 'unconscious'])
  expect(autoConditions(pc(0, 0))).toEqual(['susceptible'])
  expect(conditionsOf({ ...pc(12, 0), conditions: ['hidden', 'debilitated'] })).toEqual(['debilitated', 'susceptible', 'hidden'])
  expect(autoConditions({ health: { damage: 8, boxesPerLevel: 2 } })).toEqual(['debilitated'])
  expect(autoConditions({ health: { damage: 99 } })).toEqual([])
  expect(autoConditions(null)).toEqual([])
})

const tok = (id: string, x: number, z: number, status?: Token['status']): Token =>
  ({ id, sceneId: 's', characterId: `c${id}`, x, z, name: id, portraitUrl: null, ownerId: 'u', status })
const fire = (id: string, x: number, z: number): MapObject => ({ id, prefab: 'campfire', x, z, rotation: 0, props: {} } as MapObject)

it('light points: torch bearers and campfires only, capped', () => {
  const grid = createGrid(16, 16)
  const lit = { inventory: { body: [item('Lanterna')], backpack: [] } }
  const pts = lightPoints(grid, [tok('a', 1, 1, lit), tok('b', 2, 2, {}), tok('c', 3, 3)], [fire('f', 5, 5), { ...fire('w', 6, 6), prefab: 'wall' }])
  expect(pts.map((p) => [p.key, p.fire])).toEqual([['of', true], ['ta', false]])
  const many = Array.from({ length: 12 }, (_, i) => tok(`t${i}`, i, 0, lit))
  expect(lightPoints(grid, many, [])).toHaveLength(MAX_LIGHTS)
})

it('tokens mirror conditions and inventory from character.updated', () => {
  const out = applyCharacter([tok('a', 1, 1)], { id: 'ca', sheet: { conditions: ['hidden'], inventory: { body: [item('Tocha')], backpack: [] } } })
  expect(conditionsOf(out[0].status)).toEqual(['hidden'])
  expect(hasLightSource(out[0].status)).toBe(true)
})
