import { act, screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'
import { FakeWebSocket } from '../../test/fakeWebSocket'
import { renderApp, stubApi } from '../../test/renderApp'
import { createGrid, serializeGrid } from '../map3d/grid'

vi.mock('../map3d/TerrainView', () => ({
  TerrainView: ({ grid }: { grid: { cells: { height: number }[] } }) => <div data-testid="view">{Math.max(...grid.cells.map((c) => c.height))}</div>,
}))
vi.mock('../map3d/TerrainEditor', () => ({ TerrainEditor: () => <div data-testid="editor" /> }))

const user = { id: 'u2', email: 'b@x.com', displayName: 'Bob', isAdmin: false, createdAt: '' }
const camp = { id: 'c1', name: 'Teste', systemId: 'assimilacao', systemName: 'Assimilação RPG', role: 'player' }

afterEach(() => vi.unstubAllGlobals())

function mockServer(active: { current: string }) {
  stubApi((_m, path) => {
    if (path === '/auth/me') return { status: 200, body: user }
    if (path === '/campaigns/c1') return { status: 200, body: camp }
    if (path === '/campaigns/c1/members') return { status: 200, body: [] }
    if (path === '/campaigns/c1/scenes') {
      return { status: 200, body: ['a', 'b'].map((id) => ({ id, name: `Cena ${id}`, width: 16, depth: 16, isActive: id === active.current })) }
    }
    const id = path.split('/').pop()
    return { status: 200, body: { id, name: id, grid: serializeGrid(createGrid(16, 16)), objects: null } }
  })
}

it('shows the active scene to players, follows live patches and scene changes', async () => {
  const active = { current: 'a' }
  mockServer(active)
  renderApp('/campaigns/c1')
  expect(await screen.findByText('Cena atual: Cena a')).toBeInTheDocument()
  const view = await screen.findByTestId('view')
  expect(view).toHaveTextContent('1')
  const ws = FakeWebSocket.last
  act(() => ws.open())
  act(() => ws.receive({ type: 'map.patched', payload: { sceneId: 'a', patch: { cells: [{ x: 0, z: 0, terrain: 'stone', height: 5 }] } } }))
  expect(await screen.findByText('5')).toBeInTheDocument()
  active.current = 'b'
  act(() => ws.receive({ type: 'scenes.changed', payload: {} }))
  expect(await screen.findByText('Cena atual: Cena b')).toBeInTheDocument()
  expect(await screen.findByTestId('view')).toHaveTextContent('1')
  expect(screen.queryByTestId('editor')).toBeNull()
})

it('gives the gm the scene selector and the editor', async () => {
  mockServer({ current: 'a' })
  camp.role = 'gm'
  renderApp('/campaigns/c1')
  expect(await screen.findByRole('button', { name: 'Criar cena' })).toBeInTheDocument()
  expect(await screen.findByTestId('editor')).toBeInTheDocument()
  camp.role = 'player'
})
