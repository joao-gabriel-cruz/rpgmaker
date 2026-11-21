import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { AssignToObjective } from './AssignToObjective'
import { ConflictProvider } from './ConflictContext'
import { ConflictPanel } from './ConflictPanel'
import { progressPct, type Conflict } from './conflict'

vi.mock('../auth/auth', () => ({ useMe: () => ({ data: { id: 'u1' } }) }))
const getConflict = vi.fn()
vi.mock('../../lib/api', () => ({ api: { get: (p: string) => (p.endsWith('/conflict') ? getConflict() : Promise.resolve([])) } }))

const conflict: Conflict = {
  id: 'k1', sceneId: null, title: 'Cerco', conditions: ['Escuro'],
  objectives: [{ id: 'o1', title: 'Fugir', kind: 'main', scope: 'collective', target: 5, progress: 2 }],
  threats: [{ id: 't1', name: 'Horda', npcCharacterId: 'n1', pool: { d6: 2, d10: 0, d12: 1 } }],
  activations: [{ id: 'a1', symbol: 'pressure', threshold: 3, text: 'A horda avança', source: 'players', count: 1 }],
}

function setup(isGm: boolean, ui: React.ReactNode, send = vi.fn().mockResolvedValue(undefined)) {
  const handlers: Record<string, (p: unknown) => void> = {}
  const api = { status: 'open', send, subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => {} } } as unknown as CampaignSocketApi
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={qc}><RealtimeContext.Provider value={api}>
      <ConflictProvider campaignId="c1" isGm={isGm}>{ui}</ConflictProvider>
    </RealtimeContext.Provider></QueryClientProvider>,
  )
  return { send, emit: (t: string, p: unknown) => act(() => handlers[t](p)) }
}

beforeEach(() => getConflict.mockResolvedValue({ conflict: conflict }))

it('computes the progress percentage capped at 100', () => {
  expect(progressPct({ progress: 2, target: 5 })).toBe(40)
  expect(progressPct({ progress: 9, target: 5 })).toBe(100)
})

it('shows objectives with a progress bar and hides GM data from players', async () => {
  const players = { ...conflict, activations: undefined, threats: [{ id: 't1', name: 'Horda' }] }
  getConflict.mockResolvedValue({ conflict: players })
  setup(false, <ConflictPanel />)
  const bar = await screen.findByRole('progressbar', { name: /Fugir/ })
  expect(bar).toHaveAttribute('aria-valuenow', '2')
  expect(screen.getByText('2/5')).toBeInTheDocument()
  expect(screen.getByText('Horda')).toBeInTheDocument()
  expect(screen.queryByText(/A horda avança/)).toBeNull()
  expect(screen.queryByRole('button', { name: /Rolar/ })).toBeNull()
})

it('updates progress live on conflict.updated', async () => {
  const { emit } = setup(false, <ConflictPanel />)
  await screen.findByText('2/5')
  emit('conflict.updated', { conflict: { ...conflict, objectives: [{ ...conflict.objectives[0], progress: 4 }] } })
  expect(screen.getByText('4/5')).toBeInTheDocument()
})

it('lets the Mestre roll a threat and shows triggered activations', async () => {
  const { send, emit } = setup(true, <ConflictPanel />)
  await userEvent.click(await screen.findByRole('button', { name: 'Rolar Horda' }))
  expect(send).toHaveBeenCalledWith('conflict.threat.roll', { threatId: 't1' })
  emit('conflict.activation.triggered', { activationId: 'a1', symbol: 'pressure', threshold: 3, text: 'A horda avança', times: 1 })
  expect(within(screen.getByRole('status')).getByText(/A horda avança/)).toBeInTheDocument()
})

it('sends the edited partitura with conflict.save', async () => {
  getConflict.mockResolvedValue({ conflict: null })
  const { send } = setup(true, <ConflictPanel />)
  await userEvent.click(await screen.findByRole('button', { name: 'Criar partitura' }))
  await userEvent.type(screen.getByLabelText('Título da partitura'), 'Cerco')
  await userEvent.click(screen.getAllByRole('button', { name: 'Adicionar' })[1])
  await userEvent.type(screen.getByLabelText('Título do objetivo'), 'Fugir')
  await userEvent.click(screen.getByRole('button', { name: 'Salvar partitura' }))
  const payload = send.mock.calls[0][1]
  expect(send.mock.calls[0][0]).toBe('conflict.save')
  expect(payload).toMatchObject({ title: 'Cerco', objectives: [{ title: 'Fugir', target: 5, kind: 'main' }] })
})

it('assigns a roll to an objective and reports server errors', async () => {
  const send = vi.fn().mockRejectedValueOnce(new Error('Esta rolagem já foi atribuída a um objetivo.')).mockResolvedValue(undefined)
  setup(false, <AssignToObjective rollId="r1" ownerId="u1" />, send)
  await userEvent.click(await screen.findByRole('button', { name: 'Atribuir a objetivo…' }))
  await userEvent.click(screen.getByRole('button', { name: 'Atribuir' }))
  expect(send).toHaveBeenCalledWith('conflict.assign', { rollId: 'r1', objectiveId: 'o1' })
  expect(await screen.findByRole('alert')).toHaveTextContent('já foi atribuída')
  await userEvent.click(screen.getByRole('button', { name: 'Atribuir' }))
  expect(await screen.findByText('Atribuída ao objetivo.')).toBeInTheDocument()
})

it('hides the assign button for rolls of other players', async () => {
  setup(false, <AssignToObjective rollId="r1" ownerId="other" />)
  await waitFor(() => expect(getConflict).toHaveBeenCalled())
  expect(screen.queryByRole('button', { name: 'Atribuir a objetivo…' })).toBeNull()
})
