import { describe, expect, it } from 'vitest'
import { createGrid } from '../grid'
import { rulesFor } from './catalogData'
import {
  addObject,
  applyObjectPatch,
  moveObject,
  objectsAt,
  occupiedCells,
  removeObject,
  rotateObject,
  sanitizeObjects,
  setObjectProps,
  type MapObject,
} from './model'
import { drawWall } from './walls'

const grid = createGrid(8, 8)
const none: MapObject[] = []

describe('object model', () => {
  it('adds, and the patch reproduces the result', () => {
    const r = addObject(none, grid, { prefab: 'crate', x: 2, z: 3, id: 'a' })
    expect(r.objects).toHaveLength(1)
    expect(applyObjectPatch(none, JSON.parse(JSON.stringify(r.patch)))).toEqual(r.objects)
  })

  it('rotates footprint of a 2x1 car', () => {
    const car = { prefab: 'car', x: 1, z: 1, rotation: 0 as const }
    expect(occupiedCells(car)).toEqual([{ x: 1, z: 1 }, { x: 2, z: 1 }])
    expect(occupiedCells({ ...car, rotation: 90 })).toEqual([{ x: 1, z: 1 }, { x: 1, z: 2 }])
  })

  it('finds objects by any covered cell', () => {
    const { objects } = addObject(none, grid, { prefab: 'volcano', x: 1, z: 1, id: 'v' })
    expect(objectsAt(objects, 3, 3)).toHaveLength(1)
    expect(objectsAt(objects, 4, 3)).toHaveLength(0)
  })

  it('rejects out-of-grid footprints and unknown prefabs', () => {
    expect(addObject(none, grid, { prefab: 'volcano', x: 6, z: 0 }).patch.add).toEqual([])
    expect(addObject(none, grid, { prefab: 'car', x: -1, z: 0 }).patch.add).toEqual([])
    expect(addObject(none, grid, { prefab: 'nope', x: 0, z: 0 }).patch.add).toEqual([])
  })

})

describe('object model transforms', () => {
  it('rotates, moves and removes with serializable patches', () => {
    let { objects } = addObject(none, grid, { prefab: 'car', x: 6, z: 0, id: 'c' })
    const rot = rotateObject(objects, grid, 'c')
    expect(rot.objects[0].rotation).toBe(90)
    expect(rotateObject(rot.objects, grid, 'c', -90).objects[0].rotation).toBe(0)
    objects = rot.objects
    expect(moveObject(objects, grid, 'c', 7, 7).patch.update).toEqual([])
    const mv = moveObject(objects, grid, 'c', 2, 2)
    expect(mv.objects[0]).toMatchObject({ x: 2, z: 2 })
    expect(applyObjectPatch(objects, mv.patch)).toEqual(mv.objects)
    expect(removeObject(mv.objects, 'c').objects).toEqual([])
    expect(removeObject(mv.objects, 'zzz').patch.remove).toEqual([])
  })

})

describe('object model rules', () => {
  it('rotation that leaves the grid is refused', () => {
    const { objects } = addObject(none, grid, { prefab: 'car', x: 0, z: 7, id: 'c' })
    expect(rotateObject(objects, grid, 'c').patch.update).toEqual([])
  })

  it('door props change rules', () => {
    const { objects } = addObject(none, grid, { prefab: 'door', x: 0, z: 0, id: 'd' })
    expect(rulesFor('door', objects[0].props).blocksVision).toBe(true)
    const open = setObjectProps(objects, 'd', { open: true }).objects[0]
    expect(rulesFor('door', open.props)).toEqual({ blocksMovement: false, blocksVision: false })
  })

  it('sanitizes untrusted data', () => {
    const good = addObject(none, grid, { prefab: 'bush', x: 1, z: 1, id: 'b' }).objects[0]
    expect(sanitizeObjects([good, { id: 1 }, { ...good, x: 99 }, null], grid)).toEqual([good])
  })
})

describe('drawWall', () => {
  it('draws a straight line', () => {
    const r = drawWall(none, grid, { x: 1, z: 2 }, { x: 4, z: 2 })
    expect(r.patch.add.map((o) => [o.x, o.z])).toEqual([[1, 2], [2, 2], [3, 2], [4, 2]])
  })

  it('draws an L when both axes differ', () => {
    const r = drawWall(none, grid, { x: 1, z: 1 }, { x: 3, z: 3 })
    expect(r.patch.add.map((o) => [o.x, o.z])).toEqual([[1, 1], [2, 1], [3, 1], [3, 2], [3, 3]])
  })

  it('does not duplicate existing walls and supports reversed direction', () => {
    const first = drawWall(none, grid, { x: 1, z: 2 }, { x: 4, z: 2 })
    const second = drawWall(first.objects, grid, { x: 6, z: 2 }, { x: 3, z: 2 })
    expect(second.patch.add.map((o) => o.x)).toEqual([6, 5])
    expect(second.objects).toHaveLength(6)
    expect(applyObjectPatch(first.objects, second.patch)).toEqual(second.objects)
  })

  it('skips cells outside the grid', () => {
    expect(drawWall(none, grid, { x: 6, z: 0 }, { x: 9, z: 0 }).patch.add).toHaveLength(2)
  })
})
