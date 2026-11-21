import { act, fireEvent, renderHook } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import { applyPatch, createGrid, getCell, type MapPatch } from './grid'
import { inverseObjects, inversePaint, mergePaint, type Edit } from './history'
import { addObject, applyObjectPatch, type MapObject, type ObjectPatch } from './objects/model'
import { useEditHistory } from './useEditHistory'
import { runHistoryKey } from './useHistoryKeys'

const grid = createGrid(8, 8)
const patch = (...cells: [number, number, 'grass' | 'water', number][]): MapPatch => ({
  cells: cells.map(([x, z, terrain, height]) => ({ x, z, terrain, height })),
})
const tree = (id: string, x = 1): MapObject => ({ id, prefab: 'tree', x, z: 1, rotation: 0, props: { variant: 'pine' } })

it('builds the inverse of a paint patch from the cells it overwrites', () => {
  const p = patch([1, 1, 'grass', 3], [2, 2, 'water', 1])
  expect(inversePaint(grid, p).cells).toEqual([
    { x: 1, z: 1, terrain: 'floor', height: 1 },
    { x: 2, z: 2, terrain: 'floor', height: 1 },
  ])
  expect(applyPatch(applyPatch(grid, p), inversePaint(grid, p)).cells).toEqual(grid.cells)
})

it('builds the inverse of add, remove and update object patches', () => {
  const objs = [tree('a'), tree('b', 3)]
  const p: ObjectPatch = { add: [tree('c', 5)], remove: ['a'], update: [{ ...objs[1], rotation: 90 }] }
  const inv = inverseObjects(objs, p)
  expect(inv).toEqual({ add: [objs[0]], remove: ['c'], update: [objs[1]] })
  expect(applyObjectPatch(applyObjectPatch(objs, p), inv).map((o) => o.id).sort()).toEqual(['a', 'b'])
})

it('merges a stroke keeping the oldest before and the newest after of each cell', () => {
  const first: Edit = { kind: 'paint', undo: patch([1, 1, 'grass', 1]), redo: patch([1, 1, 'water', 2]) }
  const merged = mergePaint(first as Edit & { kind: 'paint' }, patch([1, 1, 'water', 2], [2, 1, 'grass', 1]), patch([1, 1, 'grass', 5], [2, 1, 'water', 2]))
  expect(merged).toMatchObject({
    undo: { cells: [{ x: 1, z: 1, terrain: 'grass', height: 1 }, { x: 2, z: 1, terrain: 'grass', height: 1 }] },
    redo: { cells: [{ x: 1, z: 1, terrain: 'grass', height: 5 }, { x: 2, z: 1, terrain: 'water', height: 2 }] },
  })
})

function setup() {
  const paint = vi.fn()
  const editObjects = vi.fn()
  const hook = renderHook(() => useEditHistory({ grid, objects: [], paint, editObjects }))
  return { hook, paint, editObjects }
}

it('undoes and redoes a brush stroke by sending the inverse patch as one step', () => {
  const { hook, paint } = setup()
  act(() => {
    hook.result.current.paint(patch([1, 1, 'grass', 1]))
    hook.result.current.paint(patch([2, 1, 'grass', 1]))
  })
  expect(hook.result.current.canUndo).toBe(true)
  paint.mockClear()
  act(() => hook.result.current.undo())
  expect(paint).toHaveBeenCalledTimes(1)
  expect(paint.mock.calls[0][0].cells).toEqual([
    { x: 1, z: 1, terrain: 'floor', height: 1 },
    { x: 2, z: 1, terrain: 'floor', height: 1 },
  ])
  expect(hook.result.current.canUndo).toBe(false)
  act(() => hook.result.current.redo())
  expect(paint.mock.calls[1][0].cells.map((c: { terrain: string }) => c.terrain)).toEqual(['grass', 'grass'])
})

it('starts a new undo step after the pointer is released', () => {
  const { hook, paint } = setup()
  act(() => hook.result.current.paint(patch([1, 1, 'grass', 1])))
  fireEvent.pointerUp(window)
  act(() => hook.result.current.paint(patch([2, 1, 'grass', 1])))
  paint.mockClear()
  act(() => hook.result.current.undo())
  expect(paint.mock.calls[0][0].cells).toHaveLength(1)
  expect(hook.result.current.canUndo).toBe(true)
})

it('undoes an object placement by removing it and clears redo on a new edit', () => {
  const { hook, editObjects } = setup()
  const add = addObject([], grid, { prefab: 'tree', x: 2, z: 2, id: 'n' }).patch
  act(() => hook.result.current.editObjects(add))
  editObjects.mockClear()
  act(() => hook.result.current.undo())
  expect(editObjects).toHaveBeenCalledWith({ add: [], remove: ['n'], update: [] })
  expect(hook.result.current.canRedo).toBe(true)
  act(() => hook.result.current.paint(patch([0, 0, 'grass', 1])))
  expect(hook.result.current.canRedo).toBe(false)
  expect(getCell(grid, 0, 0)).toBeDefined()
})

it('maps Ctrl/Cmd+Z to undo and Shift+Ctrl/Cmd+Z to redo', () => {
  const a = { undo: vi.fn(), redo: vi.fn() }
  expect(runHistoryKey({ key: 'z', ctrlKey: true, metaKey: false, shiftKey: false }, a)).toBe(true)
  expect(a.undo).toHaveBeenCalled()
  runHistoryKey({ key: 'Z', ctrlKey: false, metaKey: true, shiftKey: true }, a)
  expect(a.redo).toHaveBeenCalledTimes(1)
  expect(runHistoryKey({ key: 'z', ctrlKey: false, metaKey: false, shiftKey: false }, a)).toBe(false)
})
