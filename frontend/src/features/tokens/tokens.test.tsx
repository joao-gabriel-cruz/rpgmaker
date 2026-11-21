import { act, render, renderHook, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { PerspectiveCamera, Vector2 } from 'three'
import { afterEach, expect, it, vi } from 'vitest'
import { stubApi } from '../../test/renderApp'
import { createGrid } from '../map3d/grid'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { pickCell } from './pickCell'
import { applyMoved, applyPlaced, canMoveToken, initials, playerColor, spawnCandidates, type Token } from './tokens'
import { TokensPanel } from './TokensPanel'
import { useSceneTokens } from './useSceneTokens'

afterEach(() => vi.unstubAllGlobals())

const tok = (o: Partial<Token> = {}): Token => ({
  id: 't1', sceneId: 's1', characterId: 'c1', x: 1, z: 1, name: 'Ana Maria', portraitUrl: null, ownerId: 'u1', ...o,
})
const grid = createGrid(16, 16)

function fakeSocket(send: CampaignSocketApi['send'] = vi.fn().mockResolvedValue(undefined)) {
  const handlers = new Map<string, (p: unknown) => void>()
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

function setup(send?: CampaignSocketApi['send'], list: Token[] = [tok()]) {
  stubApi(() => ({ status: 200, body: list }))
  const sock = fakeSocket(send)
  const wrapper = ({ children }: { children: ReactNode }) => (
    <RealtimeContext.Provider value={sock.api}>{children}</RealtimeContext.Provider>
  )
  const hook = renderHook(() => useSceneTokens('c', 's1', grid), { wrapper })
  return { ...sock, hook }
}

it('reduces placed, moved and duplicate events', () => {
  expect(applyPlaced([], 's1', tok())).toHaveLength(1)
  expect(applyPlaced([tok()], 's1', tok())).toHaveLength(1)
  expect(applyPlaced([], 'other', tok())).toHaveLength(0)
  expect(applyMoved([tok()], 't1', 5, 6)[0]).toMatchObject({ x: 5, z: 6 })
})

it('lets players move only their tokens and the gm all', () => {
  expect(canMoveToken(tok(), 'u1', false)).toBe(true)
  expect(canMoveToken(tok(), 'u2', false)).toBe(false)
  expect(canMoveToken(tok(), 'u2', true)).toBe(true)
})

it('derives initials and a stable color', () => {
  expect(initials('Ana Maria Souza')).toBe('AS')
  expect(initials('  ')).toBe('?')
  expect(playerColor('u1')).toBe(playerColor('u1'))
  expect(playerColor(null)).toBe('#71717a')
})

it('suggests free cells near the center', () => {
  const c = spawnCandidates(grid, [tok({ x: 8, z: 8 })], 3)
  expect(c).toHaveLength(3)
  expect(c.some((p) => p.x === 8 && p.z === 8)).toBe(false)
})

it('picks the cell under the pointer', () => {
  const cam = new PerspectiveCamera(50, 1, 0.1, 100)
  cam.position.set(8.5, 20, 8.5)
  cam.lookAt(8.5, 0, 8.5)
  cam.updateMatrixWorld()
  expect(pickCell(grid, new Vector2(0, 0), cam)).toEqual({ x: 8, z: 8 })
})

it('sends token.move when dropping and applies it optimistically', async () => {
  const { hook, send } = setup()
  await waitFor(() => expect(hook.result.current.tokens).toHaveLength(1))
  act(() => hook.result.current.move('t1', 4, 5))
  expect(send).toHaveBeenCalledWith('token.move', { tokenId: 't1', x: 4, z: 5, force: false })
  expect(hook.result.current.tokens[0]).toMatchObject({ x: 4, z: 5 })
})

it('rolls the token back and shows the reason when the server refuses', async () => {
  const { hook } = setup(vi.fn().mockRejectedValue(new Error('Há um objeto bloqueando a célula.')))
  await waitFor(() => expect(hook.result.current.tokens).toHaveLength(1))
  act(() => hook.result.current.move('t1', 4, 5))
  await waitFor(() => expect(hook.result.current.error).toBe('Há um objeto bloqueando a célula.'))
  expect(hook.result.current.tokens[0]).toMatchObject({ x: 1, z: 1 })
})

it('follows remote moves, placements and removals', async () => {
  const { hook, emit } = setup()
  await waitFor(() => expect(hook.result.current.tokens).toHaveLength(1))
  emit('token.moved', { tokenId: 't1', x: 9, z: 9, by: 'u2' })
  expect(hook.result.current.tokens[0]).toMatchObject({ x: 9, z: 9 })
  emit('token.placed', { token: tok({ id: 't2', characterId: 'c2' }) })
  expect(hook.result.current.tokens).toHaveLength(2)
  emit('token.removed', { tokenId: 't1', sceneId: 's1' })
  expect(hook.result.current.tokens.map((t) => t.id)).toEqual(['t2'])
})

it('shows a received ping and centers the camera on it, ignoring other scenes', async () => {
  vi.useFakeTimers({ shouldAdvanceTime: true })
  const { hook, emit } = setup()
  emit('map.pinged', { sceneId: 'other', x: 1, z: 1, userName: 'X' })
  expect(hook.result.current.pings).toHaveLength(0)
  emit('map.pinged', { sceneId: 's1', x: 3.5, z: 4, userName: 'Ana' })
  expect(hook.result.current.pings[0]).toMatchObject({ x: 3.5, z: 4, userName: 'Ana' })
  expect(hook.result.current.camera).toMatchObject({ x: 3.5, z: 4 })
  act(() => { vi.advanceTimersByTime(3500) })
  expect(hook.result.current.pings).toHaveLength(0)
  emit('camera.focused', { sceneId: 's1', x: 7, z: 8 })
  expect(hook.result.current.camera).toMatchObject({ x: 7, z: 8 })
  vi.useRealTimers()
})

it('sends map.ping and camera.focusAll', async () => {
  const { hook, send } = setup()
  act(() => hook.result.current.ping(2, 3))
  expect(send).toHaveBeenCalledWith('map.ping', { sceneId: 's1', x: 2, z: 3 })
  act(() => hook.result.current.focusAll())
  expect(send).toHaveBeenCalledWith('camera.focusAll', { sceneId: 's1', x: 8, z: 8 })
})

function panel(send = vi.fn().mockResolvedValue(undefined) as CampaignSocketApi['send']) {
  const chars = [
    { id: 'c1', name: 'Ana', ownerId: 'u1', kind: 'pc' },
    { id: 'c2', name: 'Beto', ownerId: 'u2', kind: 'pc' },
  ]
  stubApi((_m, path) => ({ status: 200, body: path.endsWith('/tokens') ? [] : chars }))
  const sock = fakeSocket(send)
  const Harness = () => {
    const tokens = useSceneTokens('c', 's1', grid)
    return <TokensPanel campaignId="c" tokens={tokens} userId="u1" />
  }
  render(
    <QueryClientProvider client={new QueryClient()}>
      <RealtimeContext.Provider value={sock.api}><Harness /></RealtimeContext.Provider>
    </QueryClientProvider>,
  )
  return send
}

it('offers players a button to place their own token', async () => {
  const send = panel()
  await userEvent.click(await screen.findByRole('button', { name: 'Colocar meu token' }))
  expect(send).toHaveBeenCalledWith('token.place', expect.objectContaining({ sceneId: 's1', characterId: 'c1' }))
})
