import { describe, expect, it } from 'vitest'
import { lightPoints } from '../../ambience/lightPoints'
import { prefabIcon } from '../../table/options/prefabCatalog'
import { createGrid, type MapGrid } from '../grid'
import { bridgeDeck, standHeight } from './bridge'
import { CATALOG, PREFABS } from './catalog'
import { PREFAB_META_BY_ID, rulesFor } from './catalogData'
import { addObject, footprint, occupiedCells, type MapObject } from './model'
import { awningColor } from './palette'
import { baseHeight } from './placement'
import { toolProps } from './toolProps'

const IDS = ['tent', 'barrel', 'bench', 'market_stall', 'lamp_post', 'sandbags', 'ruined_wall', 'well', 'sign', 'bridge', 'bunting']
const obj = (prefab: string, x: number, z: number, props: MapObject['props'] = {}): MapObject =>
  ({ id: `${prefab}${x}${z}`, prefab, x, z, rotation: 0, props })

function gap(): MapGrid {
  const g = createGrid(5, 3)
  const cells = g.cells.map((_, i) => (i % 5 === 0 || i % 5 === 4 ? { terrain: 'grass' as const, height: 4 } : { terrain: 'water' as const, height: 0 }))
  return { ...g, cells }
}

describe('new prefabs', () => {
  it('are in the catalog with a model, icon and PT-BR name', () => {
    for (const id of IDS) {
      expect(PREFABS[id].Component).toBeTypeOf('function')
      expect(prefabIcon(id)).toBeTruthy()
      expect(PREFAB_META_BY_ID[id].name).not.toBe(id)
    }
    expect(CATALOG.length).toBeGreaterThanOrEqual(24)
  })

  it('footprints and rules mirror the backend', () => {
    expect(footprint('tent', 0)).toEqual({ w: 2, d: 2 })
    expect(footprint('bench', 90)).toEqual({ w: 1, d: 2 })
    expect(occupiedCells({ prefab: 'market_stall', x: 1, z: 1, rotation: 0 })).toHaveLength(4)
    expect(rulesFor('tent', {}).blocksMovement).toBe(true)
    expect(rulesFor('sandbags', {})).toEqual({ blocksMovement: true, blocksVision: false })
    expect(rulesFor('bridge', {})).toEqual({ blocksMovement: false, blocksVision: false })
    expect(rulesFor('ruined_wall', {}).blocksMovement).toBe(true)
  })

  it('applies default props and tool props', () => {
    const r = addObject([], createGrid(6, 6), { prefab: 'market_stall', x: 0, z: 0, id: 'm' })
    expect(r.objects[0].props).toEqual({ color: 'red' })
    expect(toolProps({ prefab: 'market_stall', variant: 'pine', color: 'blue' })).toEqual({ color: 'blue' })
    expect(toolProps({ prefab: 'ruined_wall', variant: 'pine', wallHeight: 2 })).toEqual({ height: 2 })
    expect(awningColor('green')).toBe('#4f8f5a')
    expect(awningColor('#112233')).toBe('#112233')
    expect(awningColor('nope')).toBe('#b8453a')
  })
})

describe('bridge', () => {
  it('rests at the height of the highest neighbouring terrain, or the deck prop', () => {
    const g = gap()
    expect(bridgeDeck(g, obj('bridge', 1, 1))).toBe(4)
    expect(bridgeDeck(g, obj('bridge', 2, 1, { deck: 2 }))).toBe(2)
    expect(baseHeight(g, obj('bridge', 1, 1))).toBe(4)
  })

  it('a chain of bridge cells over a wide gap shares the bank height', () => {
    const g = gap()
    const chain = [obj('bridge', 1, 1), obj('bridge', 2, 1), obj('bridge', 3, 1)]
    expect(bridgeDeck(g, chain[1], chain)).toBe(4)
    expect(standHeight(g, chain, 2, 1)).toBe(4)
    expect(baseHeight(g, chain[1], chain)).toBe(4)
  })

  it('lifts a token over the gap, only where the bridge is', () => {
    const g = gap()
    const bridge = [obj('bridge', 1, 1)]
    expect(standHeight(g, bridge, 1, 1)).toBe(4)
    expect(standHeight(g, bridge, 1, 0)).toBe(0.15)
    expect(standHeight(g, undefined, 1, 1)).toBe(0.15)
  })
})

it('a lamp post lights the night like a campfire', () => {
  const pts = lightPoints(createGrid(8, 8), [], [obj('lamp_post', 2, 2), obj('sign', 3, 3), obj('campfire', 4, 4)])
  expect(pts.map((p) => [p.key, p.fire])).toEqual([['olamp_post22', true], ['ocampfire44', true]])
  expect(pts[0].y).toBeGreaterThan(pts[1].y)
})
