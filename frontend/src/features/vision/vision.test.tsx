import { act, render, renderHook, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, expect, it, vi } from 'vitest'
import { stubApi } from '../../test/renderApp'
import { createGrid } from '../map3d/grid'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import type { Token } from '../tokens/tokens'
import { cellsWhere, decodeBits, hasBit, rectCells } from './bits'
import { FogOptions } from '../table/options/FogOptions'
import { useVision } from './useVision'
import { applyEvent, fogCover, fogInstances, visibleTokens, type VisionState } from './vision'

afterEach(() => vi.unstubAllGlobals())

const grid = createGrid(16, 16)

// Bits for the given cells, in the server's wire format.
function encode(cells: [number, number][], w = 16, d = 16): string {
  const b = new Uint8Array(Math.ceil((w * d) / 8))
  for (const [x, z] of cells) b[(z * w + x) >> 3] |= 1 << ((z * w + x) & 7)
  return btoa(String.fromCharCode(...b))
}

const tok = (o: Partial<Token> = {}): Token => ({
  id: 't1', sceneId: 's1', characterId: 'c1', x: 1, z: 1, name: 'Ana', portraitUrl: null, ownerId: 'u1', ...o,
})

it('decodes the server bitset and lists covered cells', () => {
  const bits = decodeBits(encode([[0, 0], [3, 2]]), 16, 16)
  expect(hasBit(bits, 16, 3, 2)).toBe(true)
  expect(hasBit(bits, 16, 4, 2)).toBe(false)
  expect(cellsWhere(bits, 16, 16, false)).toHaveLength(254)
  expect(decodeBits('%%%', 16, 16).every((b) => b === 0)).toBe(true)
  expect(rectCells({ x: 3, z: 3 }, { x: 1, z: 2 })).toHaveLength(6)
})

it('covers the cells outside the player sight and nothing when the fog is off', () => {
  const dyn: VisionState = { mode: 'dynamic', visionRadius: 12, visible: decodeBits(encode([[5, 5]]), 16, 16), revealed: null }
  const cover = fogCover(dyn, grid, false)
  expect(cover?.cells).toHaveLength(255)
  expect(cover?.cells.some((c) => c.x === 5 && c.z === 5)).toBe(false)
  expect(fogCover({ ...dyn, mode: 'off' }, grid, false)).toBeNull()
})

it('draws the cover as columns for players and a thin film for the Mestre', () => {
  const [p] = fogInstances(grid, [{ x: 2, z: 3 }], false)
  expect(p.h).toBeGreaterThan(grid.cells[3 * 16 + 2].height)
  const [g] = fogInstances(grid, [{ x: 2, z: 3 }], true)
  expect(g.h).toBeLessThan(0.2)
  const gmState: VisionState = { mode: 'manual', visionRadius: 12, visible: null, revealed: decodeBits(encode([[0, 0]]), 16, 16) }
  expect(fogCover(gmState, grid, true)).toMatchObject({ gm: true })
  expect(fogCover(gmState, grid, true)?.cells).toHaveLength(255)
})

it('hides tokens outside the sight except the own ones', () => {
  const s: VisionState = { mode: 'dynamic', visionRadius: 12, visible: decodeBits(encode([[1, 1]]), 16, 16), revealed: null }
  const list = [tok(), tok({ id: 't2', x: 9, z: 9 }), tok({ id: 't3', x: 9, z: 9, ownerId: 'me' })]
  expect(visibleTokens(list, s, grid, 'me', false).map((t) => t.id)).toEqual(['t1', 't3'])
  expect(visibleTokens(list, s, grid, 'me', true)).toHaveLength(3)
  expect(visibleTokens(list, { ...s, mode: 'off' }, grid, 'me', false)).toHaveLength(3)
})

it('applies vision events for players and fog events for the Mestre', () => {
  const prev: VisionState = { mode: 'off', visionRadius: 7, visible: null, revealed: null }
  const p = applyEvent(prev, { mode: 'dynamic', visible: encode([[2, 2]]) }, grid, false)
  expect(hasBit(p.visible!, 16, 2, 2)).toBe(true)
  const g = applyEvent(prev, { mode: 'manual', visionRadius: 9, revealed: encode([[4, 4]]) }, grid, true)
  expect(g.visionRadius).toBe(9)
  expect(hasBit(g.revealed!, 16, 4, 4)).toBe(true)
})

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

function setupGm() {
  stubApi(() => ({ status: 200, body: { sceneId: 's1', mode: 'manual', visionRadius: 12, revealed: '' } }))
  const sock = fakeSocket()
  const wrapper = ({ children }: { children: ReactNode }) => (
    <RealtimeContext.Provider value={sock.api}>{children}</RealtimeContext.Provider>
  )
  const hook = renderHook(() => useVision('c', 's1', grid, true), { wrapper })
  return { ...sock, hook }
}

it('sends fog.reveal and fog.hide from the Mestre tool and follows fog.updated', async () => {
  const { hook, send, emit } = setupGm()
  await waitFor(() => expect(hook.result.current.state.mode).toBe('manual'))
  act(() => hook.result.current.paint(rectCells({ x: 1, z: 1 }, { x: 2, z: 1 }), true))
  expect(send).toHaveBeenCalledWith('fog.reveal', { sceneId: 's1', cells: [{ x: 1, z: 1 }, { x: 2, z: 1 }] })
  act(() => hook.result.current.paint([{ x: 1, z: 1 }], false))
  expect(send).toHaveBeenLastCalledWith('fog.hide', { sceneId: 's1', cells: [{ x: 1, z: 1 }] })
  act(() => hook.result.current.all(true))
  expect((send.mock.calls.at(-1)![1] as { cells: unknown[] }).cells).toHaveLength(256)
  emit('fog.updated', { sceneId: 's1', mode: 'dynamic', visionRadius: 8, revealed: encode([[3, 3]]) })
  expect(hook.result.current.state).toMatchObject({ mode: 'dynamic', visionRadius: 8 })
})

it('lets the Mestre pick the mode, the radius and the brush', async () => {
  const { hook, send } = setupGm()
  await waitFor(() => expect(hook.result.current.state.mode).toBe('manual'))
  const user = userEvent.setup()
  const Panel = () => <FogOptions vision={hook.result.current} />
  const { rerender } = render(<Panel />)
  await user.click(screen.getByRole('button', { name: 'Visão dinâmica' }))
  expect(send).toHaveBeenCalledWith('fog.configure', { sceneId: 's1', mode: 'dynamic', visionRadius: 12 })
  rerender(<Panel />)
  expect(hook.result.current.brush).toBe('reveal')
  await user.click(screen.getByRole('button', { name: 'Esconder área' }))
  rerender(<Panel />)
  expect(hook.result.current.brush).toBe('hide')
  await user.click(screen.getByRole('button', { name: 'Esconder tudo' }))
  expect((send.mock.calls.at(-1)![1] as { cells: unknown[] }).cells).toHaveLength(256)
})

it('players get their sight from the vision.updated event', async () => {
  stubApi(() => ({ status: 200, body: { sceneId: 's1', mode: 'dynamic', visible: encode([[0, 0]]) } }))
  const sock = fakeSocket()
  const wrapper = ({ children }: { children: ReactNode }) => (
    <RealtimeContext.Provider value={sock.api}>{children}</RealtimeContext.Provider>
  )
  const { result } = renderHook(() => useVision('c', 's1', grid, false), { wrapper })
  await waitFor(() => expect(result.current.state.mode).toBe('dynamic'))
  sock.emit('vision.updated', { sceneId: 's1', mode: 'dynamic', visible: encode([[7, 7]]) })
  expect(hasBit(result.current.state.visible!, 16, 7, 7)).toBe(true)
  sock.emit('vision.updated', { sceneId: 'other', mode: 'off' })
  expect(result.current.state.mode).toBe('dynamic')
})
