import { act, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'
import { FakeWebSocket } from '../../test/fakeWebSocket'
import { renderApp, stubApi } from '../../test/renderApp'
import { createGrid, serializeGrid } from '../map3d/grid'

vi.mock('../map3d/TerrainView', () => ({
  TerrainView: ({ objects }: { objects: unknown[] }) => <div data-testid="view">{objects.length}</div>,
}))
vi.mock('../map3d/TerrainEditor', () => ({
  TerrainEditor: ({ objects }: { objects: unknown[] }) => <div data-testid="editor">{objects.length}</div>,
}))

const user = { id: 'u1', email: 'a@x.com', displayName: 'Ana', isAdmin: false, createdAt: '' }
const camp = { id: 'c1', name: 'T', systemId: 'assimilacao', systemName: 'A', role: 'gm' }
const volcano = { id: 'v', prefab: 'wall', x: 1, z: 1, rotation: 0 }
const grid = serializeGrid(createGrid(16, 16))

afterEach(() => { vi.unstubAllGlobals(); camp.role = 'gm' })

function mockServer(draft: { current: boolean }) {
  stubApi((_m, path) => {
    if (path === '/auth/me') return { status: 200, body: user }
    if (path === '/campaigns/c1') return { status: 200, body: camp }
    if (path === '/campaigns/c1/members') return { status: 200, body: [] }
    if (path === '/campaigns/c1/scenes') return { status: 200, body: [{ id: 'a', name: 'A', width: 16, depth: 16, isActive: true }] }
    if (path === '/campaigns/c1/scenes/a/snapshots') return { status: 200, body: [{ id: 's1', name: 'erupção', createdAt: '' }] }
    return { status: 200, body: { id: 'a', name: 'A', grid, objects: null, draft: draft.current ? { grid, objects: [] } : undefined } }
  })
}

function sent(ws: FakeWebSocket) {
  return ws.sent.map((m) => JSON.parse(m) as { type: string; payload: Record<string, unknown> })
}

it('toggles draft, shows the banner and applies/discards', async () => {
  const draft = { current: false }
  mockServer(draft)
  renderApp('/campaigns/c1')
  const toggle = await screen.findByLabelText('Rascunho')
  const ws = FakeWebSocket.last
  act(() => ws.open())
  await userEvent.click(toggle)
  await waitFor(() => expect(sent(ws).some((m) => m.type === 'map.draft.start' && m.payload.sceneId === 'a')).toBe(true))
  expect(screen.queryByText('Editando rascunho — jogadores não veem')).toBeNull()
  draft.current = true
  act(() => ws.receive({ type: 'map.draft.changed', payload: { sceneId: 'a', active: true } }))
  expect(await screen.findByText('Editando rascunho — jogadores não veem')).toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Aplicar' }))
  await waitFor(() => expect(sent(ws).some((m) => m.type === 'map.draft.apply')).toBe(true))
  act(() => ws.receive({ type: 'map.replaced', payload: { sceneId: 'a', grid, objects: [volcano] } }))
  await waitFor(() => expect(screen.getByTestId('editor')).toHaveTextContent('1'))
  expect(screen.queryByText('Editando rascunho — jogadores não veem')).toBeNull()
})

it('saves, fires and deletes saved states', async () => {
  mockServer({ current: false })
  renderApp('/campaigns/c1')
  await screen.findByText('erupção')
  const ws = FakeWebSocket.last
  act(() => ws.open())
  await userEvent.click(screen.getByRole('button', { name: 'Disparar erupção' }))
  await waitFor(() => expect(sent(ws).some((m) => m.type === 'map.snapshot.apply' && m.payload.snapshotId === 's1')).toBe(true))
})

it('shows players no draft tools and applies map.replaced', async () => {
  camp.role = 'player'
  mockServer({ current: false })
  renderApp('/campaigns/c1')
  expect(await screen.findByTestId('view')).toHaveTextContent('0')
  expect(screen.queryByLabelText('Rascunho')).toBeNull()
  expect(screen.queryByText('Estados salvos')).toBeNull()
  act(() => FakeWebSocket.last.receive({ type: 'map.replaced', payload: { sceneId: 'a', grid, objects: [volcano] } }))
  await waitFor(() => expect(screen.getByTestId('view')).toHaveTextContent('1'))
})
