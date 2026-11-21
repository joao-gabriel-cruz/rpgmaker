import { act, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { FakeWebSocket } from '../../test/fakeWebSocket'
import { renderApp, stubApi } from '../../test/renderApp'
import { openManage } from '../../test/table'
import { resetSockets } from '../realtime/useCampaignSocket'
import { newSheet } from '../sheet-assimilacao/types'

const ana = { id: 'u1', email: 'a@x.com', displayName: 'Ana', isAdmin: false, createdAt: '' }
const camp = (role: string) => ({ id: 'c1', name: 'Teste', systemId: 'assimilacao', systemName: 'Assimilação RPG', role, settings: {} })
const char = (over: Record<string, unknown> = {}) => ({
  id: 'ch1', campaignId: 'c1', ownerId: 'u1', ownerName: 'Ana', kind: 'pc', name: 'Rita',
  sheet: { ...newSheet(), identity: { ...newSheet().identity, name: 'Rita' } },
  updatedAt: '2026-01-01T00:00:01Z', derived: { boxesPerLevel: 3, currentLevel: 0, levelName: 'Saudável' }, issues: [], ...over,
})

beforeEach(() => { FakeWebSocket.instances = [] })
afterEach(() => { vi.unstubAllGlobals(); resetSockets() })

interface Opts { role?: string; character?: ReturnType<typeof char>; put?: (body: string) => { status: number; body: unknown } }

function stub(o: Opts = {}) {
  const puts: string[] = []
  const posts: string[] = []
  stubApi((method, path, body) => {
    if (path === '/auth/me') return { status: 200, body: ana }
    if (path === '/campaigns/c1') return { status: 200, body: camp(o.role ?? 'player') }
    if (method === 'GET' && path === '/campaigns/c1/characters') return { status: 200, body: [o.character ?? char()] }
    if (method === 'POST' && path === '/campaigns/c1/characters') { posts.push(body ?? ''); return { status: 201, body: char() } }
    if (method === 'GET' && path === '/campaigns/c1/characters/ch1') return { status: 200, body: o.character ?? char() }
    if (method === 'PUT' && path === '/campaigns/c1/characters/ch1') {
      puts.push(body ?? '')
      return o.put ? o.put(body ?? '') : { status: 200, body: char({ updatedAt: '2026-01-01T00:00:09Z' }) }
    }
    return { status: 200, body: [] }
  })
  return { puts, posts }
}

it('lists characters with owner and health level, and creates one', async () => {
  const { posts } = stub()
  renderApp('/campaigns/c1')
  const link = await screen.findByRole('link', { name: /Rita/ })
  expect(link).toHaveTextContent('Ana')
  expect(link).toHaveTextContent('Saudável')
  await userEvent.click(screen.getByRole('button', { name: 'Criar personagem' }))
  await waitFor(() => expect(posts).toHaveLength(1))
  expect(await screen.findByLabelText('Nome')).toHaveValue('Rita')
})

it('opens the sheet from the list and autosaves after the debounce', async () => {
  const { puts } = stub()
  renderApp('/campaigns/c1')
  await userEvent.click(await screen.findByRole('link', { name: /Rita/ }))
  const name = await screen.findByLabelText('Nome')
  await userEvent.type(name, '!')
  expect(screen.getByRole('status', { name: 'Estado do salvamento' })).toHaveTextContent('Salvando…')
  expect(puts).toHaveLength(0)
  await waitFor(() => expect(puts).toHaveLength(1))
  expect(JSON.parse(puts[0]).sheet.identity.name).toBe('Rita!')
  await waitFor(() => expect(screen.getByText('Salvo')).toBeInTheDocument())
})

it('is read-only for a player viewing someone else sheet', async () => {
  stub({ character: char({ ownerId: 'u2', ownerName: 'Bob' }) })
  renderApp('/campaigns/c1/characters/ch1')
  expect(await screen.findByText(/Ficha de Bob \(somente leitura\)/)).toBeInTheDocument()
  expect(screen.getByLabelText('Nome')).toBeDisabled()
  expect(screen.queryByText('Salvo')).toBeNull()
})

it('lets the gm edit someone else sheet', async () => {
  stub({ role: 'gm', character: char({ ownerId: 'u2', ownerName: 'Bob' }) })
  renderApp('/campaigns/c1/characters/ch1')
  expect(await screen.findByLabelText('Nome')).toBeEnabled()
})

it('applies character.updated live but waits for the local save when editing', async () => {
  let release: () => void = () => {}
  const gate = new Promise<void>((r) => { release = r })
  stub({ put: () => ({ status: 200, body: char({ updatedAt: '2026-01-01T00:00:05Z' }) }) })
  renderApp('/campaigns/c1/characters/ch1')
  const name = await screen.findByLabelText('Nome')
  const ws = FakeWebSocket.last
  const remote = (n: string, at: string) => ({ type: 'character.updated', payload: char({ updatedAt: at, sheet: { ...newSheet(), identity: { ...newSheet().identity, name: n } } }) })
  act(() => ws.receive(remote('Remoto', '2026-01-01T00:00:02Z')))
  await waitFor(() => expect(name).toHaveValue('Remoto'))
  await userEvent.type(name, 'X')
  act(() => ws.receive(remote('Outro', '2026-01-01T00:00:10Z')))
  expect(name).toHaveValue('RemotoX')
  release()
  await gate
  await waitFor(() => expect(name).toHaveValue('Outro'))
})

it('shows server issues as a warning when saved, and the reason when rejected', async () => {
  const issue = { code: 'instinct_total', severity: 'limit', field: 'instincts', message: 'Instintos acima do limite' }
  stub({ put: () => ({ status: 200, body: char({ updatedAt: '2026-01-01T00:00:09Z', issues: [issue] }) }) })
  renderApp('/campaigns/c1/characters/ch1')
  await userEvent.type(await screen.findByLabelText('Nome'), 'a')
  expect(await screen.findByText(/Ficha salva com avisos/)).toBeInTheDocument()
})

it('does not claim saved when the server blocks the sheet', async () => {
  const issue = { code: 'instinct_total', severity: 'limit', field: 'instincts', message: 'Instintos acima do limite' }
  const { puts } = stub({ put: () => ({ status: 400, body: { error: 'A ficha não pôde ser salva', issues: [issue] } }) })
  renderApp('/campaigns/c1/characters/ch1')
  await userEvent.type(await screen.findByLabelText('Nome'), 'a')
  expect(await screen.findByRole('alert')).toHaveTextContent('Não salvo: A ficha não pôde ser salva')
  expect(screen.getByRole('status', { name: 'Estado do salvamento' })).toHaveTextContent('Erro ao salvar')
  expect(puts).toHaveLength(1)
})

it('lets only the gm change the creation limits mode', async () => {
  const patches: string[] = []
  stubApi((method, path, body) => {
    if (path === '/auth/me') return { status: 200, body: ana }
    if (method === 'PATCH' && path === '/campaigns/c1/settings') { patches.push(body ?? ''); return { status: 200, body: { ...camp('gm'), settings: { creationLimits: 'block' } } } }
    if (path === '/campaigns/c1') return { status: 200, body: camp('gm') }
    return { status: 200, body: [] }
  })
  renderApp('/campaigns/c1')
  await openManage('Configurações')
  await userEvent.selectOptions(await screen.findByLabelText('Limites de criação de personagem'), 'block')
  await waitFor(() => expect(patches).toEqual(['{"creationLimits":"block"}']))
})
