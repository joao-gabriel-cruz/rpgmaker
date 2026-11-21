import { act, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { FakeWebSocket } from '../../test/fakeWebSocket'
import { renderApp, stubApi } from '../../test/renderApp'
import { resetSockets } from '../realtime/useCampaignSocket'
import { newSheet } from '../sheet-assimilacao/types'
import type { GameAction } from './actions'

vi.mock('../map3d/TerrainView', () => ({ TerrainView: () => <div data-testid="view" /> }))
vi.mock('../map3d/TerrainEditor', () => ({ TerrainEditor: () => <div data-testid="editor" /> }))

const me = { id: 'u1', email: 'a@x.com', displayName: 'Ana', isAdmin: false, createdAt: '' }
const camp = (role: string) => ({ id: 'c1', name: 'Mesa', systemId: 'assimilacao', systemName: 'Assimilação RPG', role, settings: {} })

const act_ = (over: Partial<GameAction>): GameAction => ({
  id: 'x', name: 'X', icon: 'zap', category: 'combat', target: 'none', outcome: 'none', animation: 'dance', range: undefined,
  rolls: [{ instinct: 'potency', aptitude: 'weapons', label: 'Potência + Armas' }], enabled: true, builtin: true, ...over,
})
const CATALOG: GameAction[] = [
  act_({ id: 'melee_attack', name: 'Atacar corpo a corpo', target: 'token', outcome: 'damage', animation: 'melee_strike', range: 1,
    rolls: [{ instinct: 'potency', aptitude: 'weapons', label: 'Potência + Armas' }, { instinct: 'reaction', aptitude: 'weapons', label: 'Reação + Armas' }] }),
  act_({ id: 'perceive', name: 'Perceber', icon: 'eye', category: 'exploration', outcome: 'loot', animation: 'look_around',
    rolls: [{ instinct: 'perception', aptitude: 'survival', label: 'Percepção + Sobrevivência' }] }),
  act_({ id: 'dance', name: 'Dançar', icon: 'music', category: 'social', outcome: 'morale', animation: 'dance',
    rolls: [{ instinct: 'influence', aptitude: 'expression', label: 'Influência + Expressão' }] }),
  act_({ id: 'gone', name: 'Desativada', enabled: false }),
]

function sheetWith(body: unknown[] = []) {
  const s = newSheet()
  s.identity.name = 'Rita'
  Object.assign(s.instincts, { potency: 2, reaction: 2, perception: 2, influence: 2 })
  Object.assign(s.practices, { weapons: 1, survival: 1, expression: 1 })
  s.inventory.body = body as never
  return s
}
const rita = (body: unknown[] = []) => ({
  id: 'ch1', campaignId: 'c1', ownerId: 'u1', ownerName: 'Ana', kind: 'pc', name: 'Rita', sheet: sheetWith(body),
  updatedAt: '', derived: { boxesPerLevel: 4, currentLevel: 0, levelName: 'Saudável' }, issues: [],
})

beforeEach(() => { FakeWebSocket.instances = [] })
afterEach(() => { vi.unstubAllGlobals(); resetSockets() })

function open(body: unknown[] = [], role: 'gm' | 'player' = 'player') {
  const char = rita(body)
  stubApi((method, path) => {
    if (path === '/auth/me') return { status: 200, body: me }
    if (path === '/campaigns/c1') return { status: 200, body: camp(role) }
    if (method === 'GET' && path === '/campaigns/c1/characters') return { status: 200, body: [char] }
    if (path === '/campaigns/c1/characters/ch1') return { status: 200, body: char }
    if (path === '/campaigns/c1/actions') return { status: 200, body: { actions: CATALOG } }
    if (path.startsWith('/campaigns/c1/rolls/state')) return { status: 200, body: { occupationUsed: false, markingEventUsed: false, sessionActive: true, difficultyD12: 0 } }
    return { status: 200, body: [] }
  })
  renderApp('/campaigns/c1')
  return screen.findByRole('region', { name: 'Hotbar do personagem' })
}

const sent = () => FakeWebSocket.last.sent.map((m) => JSON.parse(m) as { type: string; payload: Record<string, unknown>; reqId: string })

it('shows the action bar grouped by category, hides disabled actions and explains each in a tooltip', async () => {
  const bar = within(await open())
  const toolbar = await bar.findByRole('toolbar', { name: 'Ações' })
  expect(within(toolbar).getByRole('group', { name: 'Combate' })).toBeInTheDocument()
  expect(within(toolbar).getByRole('group', { name: 'Exploração' })).toBeInTheDocument()
  expect(within(toolbar).getByRole('group', { name: 'Social' })).toBeInTheDocument()
  expect(within(toolbar).queryByRole('button', { name: 'Desativada' })).toBeNull()
  const melee = within(toolbar).getByRole('button', { name: 'Atacar corpo a corpo' })
  expect(within(melee).getByRole('tooltip', { hidden: true })).toHaveTextContent('Potência + Armas · escolha o alvo no mapa')
})

it('opens the composer with the default roll marked and the others as chips, and sends action.perform', async () => {
  const bar = within(await open())
  act(() => FakeWebSocket.last.open())
  await userEvent.click(await bar.findByRole('button', { name: 'Dançar' }))
  const composer = await screen.findByRole('dialog', { name: 'Rita: Dançar' })
  expect(within(composer).getByRole('button', { name: 'Influência + Expressão' })).toHaveAttribute('aria-pressed', 'true')
  expect(within(composer).queryByRole('button', { name: /^Reação 2/ })).toBeNull()
  await userEvent.click(within(composer).getByLabelText('Empenho', { exact: false }))
  await userEvent.click(within(composer).getByRole('button', { name: 'Rolar' }))
  await waitFor(() => expect(sent().some((m) => m.type === 'action.perform')).toBe(true))
  expect(sent().find((m) => m.type === 'action.perform')!.payload).toMatchObject({
    characterId: 'ch1', actionId: 'dance', target: { kind: 'none' }, rollChoice: 0, modifiers: { effort: true },
  })
  expect(sent().some((m) => m.type === 'roll.request')).toBe(false)
})

it('aims first when the action needs a target, and Esc cancels the aim', async () => {
  const bar = within(await open())
  await userEvent.click(await bar.findByRole('button', { name: 'Atacar corpo a corpo' }))
  const hint = await bar.findByRole('status')
  expect(hint).toHaveTextContent('Atacar corpo a corpo: escolha o alvo no mapa (até 1 células)')
  expect(screen.queryByRole('dialog', { name: /Rita: Atacar/ })).toBeNull()
  await userEvent.keyboard('{Escape}')
  expect(bar.queryByRole('status')).toBeNull()
  await userEvent.click(bar.getByRole('button', { name: 'Atacar corpo a corpo' }))
  await userEvent.click(bar.getByRole('button', { name: 'Cancelar mira' }))
  expect(bar.queryByRole('status')).toBeNull()
})

it('pre-fills the pool preview from the default roll of the action', async () => {
  const bar = within(await open())
  act(() => FakeWebSocket.last.open())
  await userEvent.click(await bar.findByRole('button', { name: 'Perceber' }))
  const composer = await screen.findByRole('dialog', { name: 'Rita: Perceber' })
  expect(within(composer).getByTestId('pool-preview')).toHaveTextContent('2 d6 + 1 d10 · manter 1')
})

it('adds the equipped weapon kept-dice bonus to the preview', async () => {
  const sword = { name: 'Facão', kind: 'melee', attackAnimation: 'slash', effects: { keptBonus: { actionCategory: 'exploration', amount: 1 } } }
  const bar = within(await open([sword]))
  await userEvent.click(await bar.findByRole('button', { name: 'Perceber' }))
  const composer = await screen.findByRole('dialog', { name: 'Rita: Perceber' })
  expect(within(composer).getByTestId('pool-preview')).toHaveTextContent('2 d6 + 1 d10 · manter 2')
  expect(within(composer).getByText('+1 dado mantido (Facão)')).toBeInTheDocument()
})

it('starts the n-th action with its digit key', async () => {
  const bar = within(await open())
  await bar.findByRole('toolbar', { name: 'Ações' })
  await userEvent.click(document.body)
  await userEvent.keyboard('1')
  expect(await bar.findByRole('status')).toHaveTextContent('Atacar corpo a corpo')
})

it('gives the Mestre the Ações tab to disable, edit and create actions', async () => {
  const calls: { method: string; path: string; body?: string }[] = []
  const char = rita()
  stubApi((method, path, body) => {
    calls.push({ method, path, body })
    if (path === '/auth/me') return { status: 200, body: me }
    if (path === '/campaigns/c1') return { status: 200, body: camp('gm') }
    if (path === '/campaigns/c1/actions' && method === 'GET') return { status: 200, body: { actions: CATALOG } }
    if (path.startsWith('/campaigns/c1/actions')) return { status: 200, body: {} }
    if (method === 'GET' && path === '/campaigns/c1/characters') return { status: 200, body: [char] }
    return { status: 200, body: [] }
  })
  renderApp('/campaigns/c1')
  await userEvent.click(await screen.findByRole('button', { name: 'Gerenciar campanha' }))
  await userEvent.click(screen.getByRole('tab', { name: 'Ações' }))
  await userEvent.click(await screen.findByLabelText('Ativar Dançar'))
  await waitFor(() => expect(calls.some((c) => c.method === 'PUT' && c.path === '/campaigns/c1/actions/dance')).toBe(true))
  expect(JSON.parse(calls.find((c) => c.method === 'PUT')!.body!)).toMatchObject({ id: 'dance', enabled: false })
  await userEvent.click(screen.getByRole('button', { name: 'Nova ação' }))
  await userEvent.type(screen.getByLabelText('Nome da ação'), 'Tocar tambor')
  await userEvent.click(screen.getByRole('button', { name: 'Salvar' }))
  await waitFor(() => expect(calls.some((c) => c.method === 'POST' && c.path === '/campaigns/c1/actions')).toBe(true))
  expect(JSON.parse(calls.find((c) => c.method === 'POST')!.body!)).toMatchObject({ name: 'Tocar tambor', category: 'exploration', rolls: [{ instinct: 'perception' }] })
  await userEvent.click(screen.getByRole('button', { name: 'Restaurar Dançar' }))
  await waitFor(() => expect(calls.some((c) => c.method === 'DELETE' && c.path === '/campaigns/c1/actions/dance')).toBe(true))
})
