import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { FakeWebSocket } from '../../test/fakeWebSocket'
import { renderApp, stubApi } from '../../test/renderApp'
import { resetSockets } from '../realtime/useCampaignSocket'
import { newSheet } from '../sheet-assimilacao/types'
import type { NpcSheetData } from './npc'

const gmUser = { id: 'g1', email: 'g@x.com', displayName: 'Mestre', isAdmin: false, createdAt: '' }
const camp = (role: string) => ({ id: 'c1', name: 'Teste', systemId: 'assimilacao', systemName: 'Assimilação RPG', role, settings: {} })
const npcSheet = (): NpcSheetData => {
  const s = newSheet()
  return {
    identity: { name: 'Saqueador', portraitUrl: null }, instincts: s.instincts, knowledge: s.knowledge,
    practices: s.practices, health: { damage: 0, boxesPerLevel: 3 }, threats: [], notes: '',
  }
}
const npc = (over: Record<string, unknown> = {}) => ({
  id: 'n1', campaignId: 'c1', ownerId: null, ownerName: '', kind: 'npc', name: 'Saqueador', sheet: npcSheet(),
  updatedAt: '2026-01-01T00:00:01Z', derived: { boxesPerLevel: 3, currentLevel: 0, levelName: 'Saudável' }, issues: [], ...over,
})
const pc = { ...npc({ id: 'p1', kind: 'pc', name: 'Rita', ownerId: 'u1', ownerName: 'Ana' }) }

beforeEach(() => { FakeWebSocket.instances = [] })
afterEach(() => { vi.unstubAllGlobals(); resetSockets() })

function stub(role: string, list: unknown[]) {
  const posts: string[] = []
  const puts: string[] = []
  stubApi((method, path, body) => {
    if (path === '/auth/me') return { status: 200, body: gmUser }
    if (path === '/campaigns/c1') return { status: 200, body: camp(role) }
    if (method === 'GET' && path === '/campaigns/c1/characters') return { status: 200, body: list }
    if (method === 'POST' && path === '/campaigns/c1/characters') { posts.push(body ?? ''); return { status: 201, body: npc() } }
    if (method === 'GET' && path === '/campaigns/c1/characters/n1') return { status: 200, body: npc() }
    if (method === 'PUT' && path === '/campaigns/c1/characters/n1') { puts.push(body ?? ''); return { status: 200, body: npc({ updatedAt: '2026-01-01T00:00:09Z' }) } }
    return { status: 200, body: [] }
  })
  return { posts, puts }
}

it('shows the NPC list and creation only to the gm', async () => {
  stub('player', [pc])
  renderApp('/campaigns/c1')
  await screen.findByRole('link', { name: /Rita/ })
  expect(screen.queryByRole('heading', { name: 'NPCs' })).toBeNull()
  expect(screen.queryByRole('button', { name: 'Criar NPC' })).toBeNull()
})

it('lists NPCs apart from PCs for the gm and creates one', async () => {
  const { posts } = stub('gm', [pc, npc()])
  renderApp('/campaigns/c1')
  expect(await screen.findByRole('heading', { name: 'NPCs' })).toBeInTheDocument()
  expect(await screen.findByRole('link', { name: /Saqueador/ })).toBeInTheDocument()
  expect(screen.getByRole('link', { name: /Rita/ })).toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Criar NPC' }))
  await waitFor(() => expect(posts).toEqual(['{"kind":"npc"}']))
  expect(await screen.findByLabelText('Nome')).toHaveValue('Saqueador')
})

it('edits an NPC sheet with autosave', async () => {
  const { puts } = stub('gm', [npc()])
  renderApp('/campaigns/c1/characters/n1')
  await userEvent.type(await screen.findByLabelText('Nome'), '!')
  await userEvent.click(screen.getByRole('button', { name: 'Aumentar Caixas por nível' }))
  await userEvent.click(screen.getByRole('button', { name: 'Adicionar ameaça' }))
  const lastSheet = () => JSON.parse(puts[puts.length - 1] ?? '{"sheet":{"threats":[]}}').sheet as NpcSheetData
  await waitFor(() => expect(lastSheet().threats).toEqual(['']))
  await waitFor(() => expect(screen.getByText('Salvo')).toBeInTheDocument())
  const last = lastSheet()
  expect(last.identity.name).toBe('Saqueador!')
  expect(last.health.boxesPerLevel).toBe(4)
  expect(last.threats).toEqual([''])
})
