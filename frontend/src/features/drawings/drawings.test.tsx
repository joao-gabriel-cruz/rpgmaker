import { act, render, renderHook, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, expect, it, vi } from 'vitest'
import { stubApi } from '../../test/renderApp'
import { createGrid, type MapGrid } from '../map3d/grid'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { applyAdded, applyRemoved, formatMeasure, MAX_POINTS, measure, projectStroke, simplify, type Drawing } from './drawings'
import { DrawingToolbar } from './DrawingToolbar'
import { useDrawings } from './useDrawings'

afterEach(() => vi.unstubAllGlobals())

const drawing = (o: Partial<Drawing> = {}): Drawing => ({
  id: 'd1', sceneId: 's1', userId: 'u1', color: '#f87171', width: 4, points: [{ x: 1, z: 1 }, { x: 2, z: 2 }], ...o,
})

function gridWithStep(): MapGrid {
  const g = createGrid(16, 16)
  return { ...g, cells: g.cells.map((c, i) => (i === 5 * 16 + 5 ? { ...c, height: c.height + 2 } : c)) }
}

function fakeSocket() {
  const handlers = new Map<string, (p: unknown) => void>()
  const send = vi.fn().mockResolvedValue(undefined)
  const api: CampaignSocketApi = {
    status: 'open',
    send,
    subscribe: (type, h) => {
      handlers.set(type, h)
      return () => handlers.delete(type)
    },
  }
  return { api, send, emit: (type: string, p: unknown) => act(() => handlers.get(type)?.(p)) }
}

function setup(list: Drawing[] = [drawing()]) {
  stubApi(() => ({ status: 200, body: list }))
  const sock = fakeSocket()
  const wrapper = ({ children }: { children: ReactNode }) => (
    <RealtimeContext.Provider value={sock.api}>{children}</RealtimeContext.Provider>
  )
  const hook = renderHook(() => useDrawings('c', 's1', 'u1'), { wrapper })
  return { ...sock, hook }
}

it('measures Chebyshev distance and the height difference', () => {
  const m = measure(gridWithStep(), { x: 0, z: 0 }, { x: 5, z: 5 })
  expect(m).toEqual({ cells: 5, dh: 2 })
  expect(formatMeasure(m)).toBe('5 células · +2 de altura')
  expect(measure(gridWithStep(), { x: 0, z: 0 }, { x: 3, z: 1 }).cells).toBe(3)
  expect(formatMeasure({ cells: 1, dh: 0 })).toBe('1 célula')
  expect(formatMeasure({ cells: 2, dh: -1 })).toBe('2 células · −1 de altura')
})

it('simplifies strokes under the point limit and keeps the endpoint', () => {
  const dense = Array.from({ length: 5000 }, (_, i) => ({ x: i * 0.05, z: 0 }))
  const out = simplify(dense)
  expect(out.length).toBeLessThanOrEqual(MAX_POINTS)
  expect(out[out.length - 1]).toEqual(dense[dense.length - 1])
  expect(simplify([{ x: 0, z: 0 }, { x: 0.01, z: 0 }])).toHaveLength(1)
})

it('projects strokes along the terrain height', () => {
  const pts = projectStroke(gridWithStep(), [{ x: 4.5, z: 5.5 }, { x: 5.5, z: 5.5 }])
  expect(pts[0][1]).toBeLessThan(pts[pts.length - 1][1])
})

it('reduces added and removed events', () => {
  expect(applyAdded([], 's1', drawing())).toHaveLength(1)
  expect(applyAdded([drawing()], 's1', drawing())).toHaveLength(1)
  expect(applyAdded([], 'other', drawing())).toHaveLength(0)
  expect(applyRemoved([drawing()], 'd1')).toHaveLength(0)
})

it('sends drawing.add with color, width and simplified points', async () => {
  const { hook, send } = setup([])
  act(() => hook.result.current.setColor('#60a5fa'))
  act(() => hook.result.current.add([{ x: 1, z: 1 }, { x: 3, z: 3 }]))
  expect(send).toHaveBeenCalledWith('drawing.add', {
    sceneId: 's1', color: '#60a5fa', width: 4, points: [{ x: 1, z: 1 }, { x: 3, z: 3 }],
  })
})

it('updates the list from drawing events and removes own drawings', async () => {
  const { hook, emit, send } = setup([drawing(), drawing({ id: 'd2', userId: 'u2' })])
  await waitFor(() => expect(hook.result.current.drawings).toHaveLength(2))
  await emit('drawing.added', { drawing: drawing({ id: 'd3' }) })
  expect(hook.result.current.drawings).toHaveLength(3)
  act(() => hook.result.current.removeMine())
  expect(send).toHaveBeenCalledWith('drawing.remove', { drawingId: 'd1' })
  expect(send).toHaveBeenCalledWith('drawing.remove', { drawingId: 'd3' })
  expect(send).not.toHaveBeenCalledWith('drawing.remove', { drawingId: 'd2' })
  await emit('drawing.removed', { drawingId: 'd1', sceneId: 's1' })
  expect(hook.result.current.drawings.map((d) => d.id)).toEqual(['d2', 'd3'])
  await emit('drawings.cleared', { sceneId: 's1' })
  expect(hook.result.current.drawings).toHaveLength(0)
})

it('shows remote rulers and drops them on release', async () => {
  const { hook, emit } = setup([])
  const ev = { sceneId: 's1', userId: 'u2', userName: 'Beto' }
  await emit('ruler.moved', { ...ev, from: { x: 1, z: 1 }, to: { x: 4, z: 2 } })
  expect(hook.result.current.rulers).toHaveLength(1)
  await emit('ruler.moved', { ...ev, from: null, to: null })
  expect(hook.result.current.rulers).toHaveLength(0)
})

it('toolbar toggles modes for players and offers clear only to the Mestre', async () => {
  const { hook } = setup([drawing()])
  await waitFor(() => expect(hook.result.current.drawings).toHaveLength(1))
  const Bar = ({ gm }: { gm: boolean }) => <DrawingToolbar draw={hook.result.current} isGm={gm} />
  const { rerender } = render(<Bar gm={false} />)
  expect(screen.queryByRole('button', { name: 'Limpar desenhos' })).toBeNull()
  await userEvent.click(screen.getByRole('button', { name: 'Régua' }))
  expect(hook.result.current.mode).toBe('ruler')
  rerender(<Bar gm />)
  expect(screen.getByRole('button', { name: 'Limpar desenhos' })).toBeEnabled()
})
