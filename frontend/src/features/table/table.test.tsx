import { act, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'
import { FakeWebSocket } from '../../test/fakeWebSocket'
import { renderApp, stubApi } from '../../test/renderApp'
import { openTab } from '../../test/table'

vi.mock('../map3d/TerrainView', () => ({ TerrainView: () => <div data-testid="view" /> }))
vi.mock('../map3d/TerrainEditor', () => ({ TerrainEditor: () => <div data-testid="editor" /> }))

const me = { id: 'u1', email: 'a@x.com', displayName: 'Ana', isAdmin: false, createdAt: '' }
const camp = (role: string) => ({ id: 'c1', name: 'Mesa', systemId: 'assimilacao', systemName: 'Assimilação RPG', role })

afterEach(() => vi.unstubAllGlobals())

function open(role: 'gm' | 'player') {
  stubApi((_m, path) => {
    if (path === '/auth/me') return { status: 200, body: me }
    if (path === '/campaigns/c1') return { status: 200, body: camp(role) }
    return { status: 200, body: [] }
  })
  renderApp('/campaigns/c1')
  return screen.findByRole('navigation', { name: 'Ferramentas' })
}

it('shows the Mestre tools only to the Mestre', async () => {
  const rail = within(await open('gm'))
  for (const n of ['Mover', 'Ping', 'Régua', 'Desenhar', 'Terreno', 'Objetos', 'Muro', 'Névoa', 'Tokens', 'Painel de rascunho']) {
    expect(rail.getByRole('button', { name: n })).toBeInTheDocument()
  }
})

it('gives players only move, ping, ruler and draw', async () => {
  const rail = within(await open('player'))
  expect(rail.getAllByRole('button').map((b) => b.getAttribute('aria-label'))).toEqual(['Mover', 'Ping', 'Régua', 'Desenhar'])
  expect(screen.queryByRole('button', { name: 'Gerenciar campanha' })).toBeNull()
  expect(screen.queryByRole('tab', { name: /Efeitos/ })).toBeNull()
})

it('marks the active tool, follows keyboard shortcuts and Esc, and shows the help on ?', async () => {
  const rail = within(await open('gm'))
  expect(rail.getByRole('button', { name: 'Mover' })).toHaveAttribute('aria-pressed', 'true')
  await userEvent.keyboard('t')
  expect(rail.getByRole('button', { name: 'Terreno' })).toHaveAttribute('aria-pressed', 'true')
  expect(screen.getByRole('region', { name: 'Opções da ferramenta' })).toBeVisible()
  await userEvent.keyboard('{Escape}')
  expect(rail.getByRole('button', { name: 'Mover' })).toHaveAttribute('aria-pressed', 'true')
  await userEvent.keyboard('?')
  expect(await screen.findByRole('dialog', { name: 'Atalhos de teclado' })).toBeInTheDocument()
})

it('collapses and expands the sidebar and switches tabs', async () => {
  await open('gm')
  const side = screen.getByRole('complementary', { name: 'Barra lateral' })
  expect(within(side).getByRole('tab', { name: 'Grupo' })).toHaveAttribute('aria-selected', 'true')
  await openTab('Refúgio')
  expect(within(side).getByRole('tab', { name: 'Refúgio' })).toHaveAttribute('aria-selected', 'true')
  await userEvent.click(within(side).getByRole('button', { name: 'Recolher barra lateral' }))
  expect(within(side).getByRole('button', { name: 'Expandir barra lateral' })).toHaveAttribute('aria-expanded', 'false')
  expect(screen.getByRole('tabpanel', { hidden: true, name: 'Refúgio' })).not.toBeVisible()
  await userEvent.keyboard(']')
  expect(within(side).getByRole('button', { name: 'Recolher barra lateral' })).toBeInTheDocument()
  await userEvent.keyboard('[[')
  expect(within(side).getByRole('button', { name: 'Expandir barra lateral' })).toBeInTheDocument()
})

it('badges pending effects (Mestre) and new rolls while the Rolagens tab is closed', async () => {
  await open('gm')
  const ws = FakeWebSocket.last
  act(() => ws.open())
  const sug = { id: 's1', campaignId: 'c1', characterId: 'ch', characterName: 'Ana', sceneId: 'sc', x: 1, z: 2, source: 'Lava', trigger: 'on_enter', effect: 'damage', amount: 2, message: '', createdAt: '' }
  act(() => ws.receive({ type: 'effect.suggested', payload: { suggestion: sug } }))
  expect(await screen.findByRole('tab', { name: 'Efeitos (1)' })).toBeInTheDocument()
  expect(screen.getByTestId('badge-effects')).toHaveTextContent('1')
  const roll = { rollId: 'r1', userId: 'u2', userName: 'Bob', kind: 'generic', visibility: 'public', dice: [{ kind: 'd6', face: 3, difficulty: false }], total: 3, request: { expression: '1d6' } }
  act(() => ws.receive({ type: 'roll.rolled', payload: roll }))
  expect(await screen.findByRole('tab', { name: 'Rolagens (novo)' })).toBeInTheDocument()
  await openTab('Rolagens')
  expect(screen.getByRole('tab', { name: 'Rolagens' })).toBeInTheDocument()
})
