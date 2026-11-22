import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { RefugeSection } from './RefugeSection'
import { LEVEL_LABELS, RESOURCE_GROUPS, type Refuge, type ResourceKey } from './refuge'

const resources = Object.fromEntries(RESOURCE_GROUPS.flatMap((g) => g.keys).map((k) => [k, 0])) as Record<ResourceKey, number>
const refuge: Refuge = {
  campaignId: 'c1', name: 'Refúgio', description: 'Fábrica velha', notes: '', updatedAt: '2026-01-01T00:00:00Z',
  defense: 3, morale: 0, belligerence: 6, mobility: 5, week: 4,
  population: { level: 7, max: 10 }, reservesMax: 10, hasWaterSource: false,
  resources: { ...resources, water: 8, food: 5 },
  region: { size: 2, danger: 4, habitation: 0, naturalResources: 1, contamination: 0, travel: 0,
    abundance: { water: 'none', plants: 'medium', animals: 'none', wood: 'abundant', minerals: 'none', biomass: 'none' } },
  buildings: [], proposals: [], history: [], crises: [], residents: [], expeditions: [], crisisSettings: { moralePenalty: true }, conditions: [],
  derived: { defenseLabel: 'Muro de madeira', moraleLabel: 'Amotinada', belligerenceLabel: 'Arrasadora', inhabitants: 70, workers: 6, reservesTotal: 13, reservesFree: 0, reservesOverflow: 3, overcrowded: false },
  items: [{ name: 'Faca', quality: 1, scarcity: 2, pressure: 0, traits: [] }],
}
const pc = { id: 'p1', name: 'Rita', kind: 'pc', ownerId: 'me', ownerName: 'Ana' }

let handlers: Record<string, (p: unknown) => void>
const send = vi.fn(() => Promise.resolve())
const socket: CampaignSocketApi = {
  status: 'open', send,
  subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => { delete handlers[t] } },
} as unknown as CampaignSocketApi

function mockFetch() {
  return vi.spyOn(globalThis, 'fetch').mockImplementation((input, init) => {
    const url = String(input)
    const body = url.endsWith('/refuge') ? (init?.method === 'PUT' ? { ...refuge, ...JSON.parse(String(init.body)) } : refuge)
      : url.endsWith('/characters') ? [pc] : { id: 'me', email: 'a@x', displayName: 'Ana', isAdmin: false }
    return Promise.resolve(new Response(JSON.stringify(body), { status: 200 }))
  })
}

const putBodies = (f: ReturnType<typeof mockFetch>) =>
  f.mock.calls.filter(([, i]) => i?.method === 'PUT').map(([, i]) => JSON.parse(String(i?.body)))

const renderIt = (isGm: boolean) => render(
  <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
    <RealtimeContext.Provider value={socket}><RefugeSection campaignId="c1" isGm={isGm} /></RealtimeContext.Provider>
  </QueryClientProvider>,
)

beforeEach(() => { handlers = {}; send.mockClear() })
afterEach(() => vi.restoreAllMocks())

it('shows the current label of each track and the full level scales', async () => {
  mockFetch()
  renderIt(false)
  expect(await screen.findByLabelText('Defesa atual')).toHaveTextContent('Muro de madeira')
  expect(screen.getByLabelText('Moral atual')).toHaveTextContent('Amotinada')
  expect(screen.getByLabelText('Beligerância atual')).toHaveTextContent('Arrasadora')
  const steps = within(screen.getByTestId('track-Defesa')).getAllByTitle(/.+/)
  expect(steps.map((s) => s.getAttribute('title'))).toEqual(LEVEL_LABELS.defense)
  expect(steps.filter((s) => s.dataset.filled === 'true')).toHaveLength(4)
})

it('gm raises Moral and the whole editable state is PUT', async () => {
  const f = mockFetch()
  renderIt(true)
  fireEvent.click(await screen.findByLabelText('Aumentar Moral'))
  await waitFor(() => expect(putBodies(f)).toHaveLength(1))
  const body = putBodies(f)[0]
  expect(body.morale).toBe(1)
  expect(body.defense).toBe(3)
  expect(body.resources.water).toBe(8)
  expect(body.population).toEqual({ level: 7, max: 10 })
})

it('gm edits a resource, the water source and a region abundance', async () => {
  const f = mockFetch()
  renderIt(true)
  fireEvent.click(await screen.findByLabelText('Diminuir Água'))
  fireEvent.click(screen.getByLabelText('Fonte de Água'))
  fireEvent.change(screen.getByLabelText('Abundância de Animais'), { target: { value: 'scarce' } })
  fireEvent.click(screen.getByLabelText('Aumentar Perigo'))
  await waitFor(() => expect(putBodies(f)).toHaveLength(4))
  const [a, b, c, d] = putBodies(f)
  expect(a.resources.water).toBe(7)
  expect(b.hasWaterSource).toBe(true)
  expect(c.region.abundance.animals).toBe('scarce')
  expect(d.region.danger).toBe(5)
})

it('groups the 12 resources in three groups', async () => {
  mockFetch()
  renderIt(false)
  await screen.findByLabelText('Água')
  for (const g of RESOURCE_GROUPS) expect(screen.getByText(g.title)).toBeInTheDocument()
  expect(RESOURCE_GROUPS.flatMap((g) => g.keys)).toHaveLength(12)
  expect(screen.getByLabelText('Material de Construção')).toHaveTextContent('0')
})

it('shows population derived values and the reserves overflow alert', async () => {
  mockFetch()
  renderIt(false)
  expect(await screen.findByLabelText('Habitantes')).toHaveTextContent('70')
  expect(screen.getByLabelText('Trabalhadores')).toHaveTextContent('6')
  expect(screen.getByLabelText('Total de Reservas')).toHaveTextContent('13')
  expect(screen.getByRole('alert')).toHaveTextContent('Excedente de 3')
})

it('players are read-only and live updates replace the values', async () => {
  mockFetch()
  renderIt(false)
  expect(await screen.findByLabelText('Água')).toHaveTextContent('8')
  expect(screen.queryByLabelText('Diminuir Água')).toBeNull()
  expect(screen.queryByLabelText('Aumentar Moral')).toBeNull()
  expect(screen.getByLabelText('Fonte de Água')).toBeDisabled()
  expect(screen.getByLabelText('Abundância de Água')).toBeDisabled()
  act(() => handlers['refuge.updated']({ refuge: { ...refuge, resources: { ...refuge.resources, water: 6 } } }))
  await waitFor(() => expect(screen.getByLabelText('Água')).toHaveTextContent('6'))
})

it('region section collapses', async () => {
  mockFetch()
  renderIt(false)
  expect(await screen.findByLabelText('Abundância de Madeira')).toHaveValue('abundant')
  fireEvent.click(screen.getByText('Região do Refúgio'))
  expect(screen.queryByLabelText('Abundância de Madeira')).toBeNull()
})

it('take sends withdraw with the chosen character and destination', async () => {
  mockFetch()
  renderIt(false)
  await screen.findByText('Faca')
  await screen.findByRole('option', { name: 'Rita' })
  fireEvent.change(screen.getByLabelText('Destino'), { target: { value: 'body' } })
  fireEvent.click(screen.getByText('Pegar'))
  expect(send).toHaveBeenCalledWith('refuge.item.withdraw', { characterId: 'p1', to: 'body', refugeIndex: 0 })
})

it('warns when the population level is above the cap', async () => {
  const over = { ...refuge, population: { level: 12, max: 10 }, derived: { ...refuge.derived, overcrowded: true } }
  vi.spyOn(globalThis, 'fetch').mockImplementation(() => Promise.resolve(new Response(JSON.stringify(over), { status: 200 })))
  renderIt(false)
  expect(await screen.findByText(/Superlotação: a População \(12\) passa do teto \(10\)/)).toBeInTheDocument()
})
