import { act, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { FakeWebSocket } from '../../test/fakeWebSocket'
import { renderApp, stubApi } from '../../test/renderApp'
import { resetSockets } from '../realtime/useCampaignSocket'
import { newSheet } from '../sheet-assimilacao/types'

vi.mock('../map3d/TerrainView', () => ({ TerrainView: () => <div data-testid="view" /> }))
vi.mock('../map3d/TerrainEditor', () => ({ TerrainEditor: () => <div data-testid="editor" /> }))

const me = { id: 'u1', email: 'a@x.com', displayName: 'Ana', isAdmin: false, createdAt: '' }
const camp = (role: string) => ({ id: 'c1', name: 'Mesa', systemId: 'assimilacao', systemName: 'Assimilação RPG', role, settings: {} })

function sheetOf(name: string, damage = 0, determinationPoints = 5) {
  const s = newSheet()
  s.identity.name = name
  s.instincts.reaction = 2
  s.practices.athletics = 1
  s.health.damage = damage
  s.tugOfWar = { determinationLevel: 6, determinationPoints, assimilationLevel: 4, assimilationPoints: 1 }
  return s
}
const char = (id: string, name: string, ownerId: string | null, over: Record<string, unknown> = {}) => ({
  id, campaignId: 'c1', ownerId, ownerName: ownerId === 'u1' ? 'Ana' : 'Bob', kind: 'pc', name, sheet: sheetOf(name),
  updatedAt: '2026-01-01T00:00:01Z', derived: { boxesPerLevel: 4, currentLevel: 0, levelName: 'Saudável' }, issues: [], ...over,
})

beforeEach(() => { FakeWebSocket.instances = [] })
afterEach(() => { vi.unstubAllGlobals(); resetSockets() })

function open(role: 'gm' | 'player', list: ReturnType<typeof char>[]) {
  stubApi((method, path) => {
    if (path === '/auth/me') return { status: 200, body: me }
    if (path === '/campaigns/c1') return { status: 200, body: camp(role) }
    if (method === 'GET' && path === '/campaigns/c1/characters') return { status: 200, body: list }
    const one = list.find((c) => path === `/campaigns/c1/characters/${c.id}`)
    if (one) return { status: 200, body: one }
    if (path.startsWith('/campaigns/c1/rolls/state')) return { status: 200, body: { occupationUsed: false, markingEventUsed: false, sessionActive: true, difficultyD12: 0 } }
    return { status: 200, body: [] }
  })
  renderApp('/campaigns/c1')
  return screen.findByRole('region', { name: 'Hotbar do personagem' })
}

const sentTypes = () => FakeWebSocket.last.sent.map((m) => JSON.parse(m) as { type: string; payload: Record<string, unknown> })

it('shows the player their character with health level, penalty and the tug of war, and follows live updates', async () => {
  const bar = within(await open('player', [char('ch1', 'Rita', 'u1', { sheet: sheetOf('Rita', 7, 0) }), char('ch2', 'Outro', 'u2')]))
  expect(await bar.findByText('Rita')).toBeInTheDocument()
  expect(bar.queryByLabelText('Personagem ativo')).toBeNull()
  expect(bar.getByText('Laceração · −1 Sucesso')).toBeInTheDocument()
  expect(bar.getByText('Determinação 0/6')).toBeInTheDocument()
  expect(bar.getByText('Assimilação 1/4')).toBeInTheDocument()
  expect(bar.getByText('Suscetível')).toBeInTheDocument()
  act(() => FakeWebSocket.last.open())
  act(() => FakeWebSocket.last.receive({ type: 'character.updated', payload: char('ch1', 'Rita', 'u1', { sheet: sheetOf('Rita', 13, 3) }) }))
  expect(await bar.findByText('Debilitação · −2 Sucessos')).toBeInTheDocument()
  expect(bar.getByText('Debilitado')).toBeInTheDocument()
  expect(bar.queryByText('Suscetível')).toBeNull()
})

it('lets a player with several characters switch between them', async () => {
  const bar = within(await open('player', [char('ch1', 'Rita', 'u1'), char('ch2', 'Dora', 'u1')]))
  const select = await bar.findByLabelText('Personagem ativo')
  expect(select).toHaveValue('ch1')
  await userEvent.selectOptions(select, 'ch2')
  expect(bar.getByLabelText('Personagem ativo')).toHaveValue('ch2')
})

it('gives the Mestre a selector, with Rolar only for PCs', async () => {
  const npc = char('n1', 'Saqueador', null, { kind: 'npc' })
  const bar = within(await open('gm', [char('ch1', 'Rita', 'u1'), npc]))
  const select = await bar.findByLabelText('Personagem ativo')
  expect(select).toHaveValue('')
  expect(bar.queryByRole('button', { name: /Rolar/ })).toBeNull()
  await userEvent.selectOptions(select, 'ch1')
  expect(bar.getByRole('button', { name: /Rolar/ })).toBeInTheDocument()
  await userEvent.selectOptions(select, 'n1')
  expect(bar.getByRole('button', { name: 'Ficha' })).toBeInTheDocument()
  expect(bar.queryByRole('button', { name: /Rolar/ })).toBeNull()
})

it('rolls from the hotbar with Space, remembers the roll and repeats it with Enter', async () => {
  const bar = within(await open('player', [char('ch1', 'Rita', 'u1')]))
  await bar.findByText('Rita')
  act(() => FakeWebSocket.last.open())
  await userEvent.click(document.body)
  await userEvent.keyboard(' ')
  const composer = await screen.findByRole('dialog', { name: 'Compositor de rolagem' })
  await userEvent.click(within(composer).getByRole('button', { name: /^Reação/ }))
  await userEvent.click(within(composer).getByRole('button', { name: /^Atletismo/ }))
  expect(within(composer).getByTestId('pool-preview')).toHaveTextContent('2 d6 + 1 d10 · manter 1')
  await userEvent.click(within(composer).getByRole('button', { name: 'Rolar' }))
  await waitFor(() => expect(sentTypes().filter((m) => m.type === 'roll.request')).toHaveLength(1))
  const first = sentTypes().find((m) => m.type === 'roll.request')!
  expect(first.payload).toMatchObject({ characterId: 'ch1', instinct: 'reaction', aptitude: 'athletics', secret: false })
  act(() => FakeWebSocket.last.receive({ type: 'ack', payload: { reqId: (JSON.parse(FakeWebSocket.last.sent.at(-1)!) as { reqId: string }).reqId } }))
  await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Compositor de rolagem' })).toBeNull())
  const again = await bar.findByRole('button', { name: /Rolar de novo/ })
  await userEvent.click(document.body)
  await userEvent.keyboard('{Enter}')
  await waitFor(() => expect(sentTypes().filter((m) => m.type === 'roll.request')).toHaveLength(2))
  expect(again).toBeInTheDocument()
})

it('closes the composer with Escape', async () => {
  const bar = within(await open('player', [char('ch1', 'Rita', 'u1')]))
  await bar.findByText('Rita')
  await userEvent.keyboard(' ')
  expect(await screen.findByRole('dialog', { name: 'Compositor de rolagem' })).toBeInTheDocument()
  await userEvent.keyboard('{Escape}')
  expect(screen.queryByRole('dialog', { name: 'Compositor de rolagem' })).toBeNull()
})

it('opens the sheet in the drawer from the hotbar and from the Grupo card, keeping the table', async () => {
  const bar = within(await open('player', [char('ch1', 'Rita', 'u1')]))
  await userEvent.click(await bar.findByRole('button', { name: 'Ficha' }))
  const drawer = await screen.findByRole('complementary', { name: 'Ficha do personagem' })
  expect(await within(drawer).findByLabelText('Nome')).toHaveValue('Rita')
  expect(within(drawer).getByRole('link', { name: /Abrir em página/ })).toHaveAttribute('href', '/campaigns/c1/characters/ch1')
  expect(screen.getByRole('navigation', { name: 'Ferramentas' })).toBeInTheDocument()
  await userEvent.click(within(drawer).getByRole('button', { name: 'Fechar ficha' }))
  expect(screen.queryByRole('complementary', { name: 'Ficha do personagem' })).toBeNull()
  await userEvent.click(within(screen.getByRole('tabpanel', { name: /Grupo/ })).getByRole('button', { name: /Rita/ }))
  expect(await screen.findByRole('complementary', { name: 'Ficha do personagem' })).toBeInTheDocument()
  await userEvent.keyboard('{Escape}')
  expect(screen.queryByRole('complementary', { name: 'Ficha do personagem' })).toBeNull()
})

it('selecting a character in the Grupo makes it the Mestre active one', async () => {
  const bar = within(await open('gm', [char('ch1', 'Rita', 'u1')]))
  await bar.findByLabelText('Personagem ativo')
  await userEvent.click(within(screen.getByRole('tabpanel', { name: /Grupo/, hidden: true })).getByRole('button', { name: /Rita/, hidden: true }))
  expect(bar.getByLabelText('Personagem ativo')).toHaveValue('ch1')
})
