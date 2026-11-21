import { act, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'
import { FakeWebSocket } from '../../test/fakeWebSocket'
import { renderApp, stubApi } from '../../test/renderApp'
import { openManage } from '../../test/table'

const ana = { id: 'u1', email: 'a@x.com', displayName: 'Ana', isAdmin: false, createdAt: '' }
const camp = { id: 'c1', name: 'Teste', systemId: 'assimilacao', systemName: 'Assimilação RPG', role: 'gm' }

afterEach(() => vi.unstubAllGlobals())

it('creates a campaign choosing the system and opens it', async () => {
  const posted: string[] = []
  stubApi((method, path, body) => {
    if (path === '/auth/me') return { status: 200, body: ana }
    if (path === '/systems') return { status: 200, body: [{ id: 'assimilacao', name: 'Assimilação RPG', version: '1' }] }
    if (method === 'POST' && path === '/campaigns') {
      posted.push(body ?? '')
      return { status: 201, body: camp }
    }
    if (path === '/campaigns/c1') return { status: 200, body: camp }
    if (path === '/campaigns/c1/members') return { status: 200, body: [{ userId: 'u1', displayName: 'Ana', role: 'gm' }] }
    if (path === '/campaigns') return { status: 200, body: [] }
    return { status: 200, body: {} }
  })
  renderApp('/')
  await userEvent.click(await screen.findByRole('button', { name: 'Nova campanha' }))
  await userEvent.type(await screen.findByLabelText('Nome da campanha'), 'Teste')
  await screen.findByRole('option', { name: 'Assimilação RPG' })
  await userEvent.click(screen.getByRole('button', { name: 'Criar campanha' }))
  expect(await screen.findByRole('heading', { name: 'Teste' })).toBeInTheDocument()
  expect(posted[0]).toContain('"systemId":"assimilacao"')
  expect(screen.getByText(/Você é Mestre/)).toBeInTheDocument()
})

it('lets the gm generate a copyable invite link and remove a player', async () => {
  stubApi((method, path) => {
    if (path === '/auth/me') return { status: 200, body: ana }
    if (path === '/campaigns/c1') return { status: 200, body: camp }
    if (path === '/campaigns/c1/members') {
      return { status: 200, body: [
        { userId: 'u1', displayName: 'Ana', role: 'gm' },
        { userId: 'u2', displayName: 'Bob', role: 'player' },
      ] }
    }
    if (method === 'POST' && path === '/campaigns/c1/invites') {
      return { status: 201, body: { token: 'tok123', campaignName: 'Teste', expiresAt: '' } }
    }
    return { status: 200, body: {} }
  })
  renderApp('/campaigns/c1')
  await openManage()
  expect(await screen.findByRole('button', { name: 'Remover Bob' })).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Remover Ana' })).toBeNull()
  await userEvent.click(screen.getByRole('button', { name: 'Gerar convite' }))
  expect(await screen.findByLabelText('Link do convite')).toHaveValue(`${window.location.origin}/invite/tok123`)
})

it('shows the signup form for a valid invite when logged out', async () => {
  stubApi((_m, path) => {
    if (path === '/auth/me') return { status: 401 }
    if (path === '/invites/tok') return { status: 200, body: { token: 'tok', campaignName: 'Teste', expiresAt: '' } }
    return { status: 200, body: {} }
  })
  renderApp('/invite/tok')
  expect(await screen.findByRole('heading', { name: 'Convite para Teste' })).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Cadastrar e entrar' })).toBeInTheDocument()
})

it('offers to join when already logged in', async () => {
  stubApi((_m, path) => {
    if (path === '/auth/me') return { status: 200, body: ana }
    if (path === '/invites/tok') return { status: 200, body: { token: 'tok', campaignName: 'Teste', expiresAt: '' } }
    return { status: 200, body: {} }
  })
  renderApp('/invite/tok')
  expect(await screen.findByRole('button', { name: 'Entrar na campanha' })).toBeInTheDocument()
})

it('shows the invalid invite message', async () => {
  stubApi((_m, path) => {
    if (path === '/auth/me') return { status: 401 }
    return { status: 404, body: { error: 'Convite inválido ou expirado' } }
  })
  renderApp('/invite/old')
  expect(await screen.findByRole('alert')).toHaveTextContent('Convite inválido ou expirado')
})

it('shows who is online and the connection state on the campaign page', async () => {
  stubApi((_m, path) => {
    if (path === '/auth/me') return { status: 200, body: ana }
    if (path === '/campaigns/c1') return { status: 200, body: camp }
    return { status: 200, body: [] }
  })
  renderApp('/campaigns/c1')
  expect(await screen.findByText('Conectando…')).toBeInTheDocument()
  const ws = FakeWebSocket.last
  act(() => ws.open())
  act(() => ws.receive({ type: 'presence.updated', payload: { users: [{ userId: 'u1', displayName: 'Ana', role: 'gm' }] } }))
  expect(await screen.findByText('Conectado')).toBeInTheDocument()
  expect(within(screen.getByRole('list', { name: 'Online agora' })).getByLabelText('Ana (Mestre)')).toBeInTheDocument()
  act(() => ws.drop())
  expect(await screen.findByText('Reconectando…')).toBeInTheDocument()
})
