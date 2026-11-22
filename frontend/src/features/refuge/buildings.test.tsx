import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { CATALOG, allocationOf, costAt, materialCost, defOf, totalOf, type Building, type WeekReport } from './buildings'
import { RefugeSection } from './RefugeSection'
import { RESOURCE_GROUPS, type Refuge, type ResourceKey } from './refuge'

const resources = Object.fromEntries(RESOURCE_GROUPS.flatMap((g) => g.keys).map((k) => [k, 0])) as Record<ResourceKey, number>
const mill: Building = { id: 'b1', kind: 'mill', level: 1, built: 1, state: 'done', progress: 10, paid: false, workers: 2 }
const fort: Building = { id: 'b2', kind: 'fortification', level: 2, built: 1, state: 'building', progress: 5, paid: true, workers: 1 }
const report = (week: number, extra: Partial<WeekReport> = {}): WeekReport => ({
  week, entered: { food: 3 }, spent: { water: 6 }, missing: { wood: 2 }, lost: { food: 1 }, adjustments: {},
  works: [{ buildingId: 'b2', name: 'Fortificação', level: 2, points: 1, progress: 6, cost: 20, done: false }],
  completed: ['Moinho'], notes: [], triggers: [{ kind: 'shortage', resource: 'wood', detail: 'Faltou Madeira (precisava 6, tinha 4)' }], ...extra,
})
const refuge: Refuge = {
  campaignId: 'c1', name: 'Refúgio', description: '', notes: '', updatedAt: '2026-01-01T00:00:00Z',
  defense: 1, morale: 2, belligerence: 0, mobility: 0, week: 4, population: { level: 7, max: 10 }, reservesMax: 50, hasWaterSource: false,
  resources: { ...resources, water: 8 },
  region: { size: 2, danger: 1, habitation: 0, naturalResources: 1, contamination: 0, travel: 0,
    abundance: { water: 'none', plants: 'medium', animals: 'none', wood: 'abundant', minerals: 'none', biomass: 'none' } },
  buildings: [mill, fort], proposals: [{ id: 'p1', kind: 'sawmill', userId: 'u2', by: 'Bia' }], history: [report(2), report(3)],
  crises: [], expeditions: [], crisisSettings: { moralePenalty: true }, conditions: [], residents: [], items: [],
  derived: { defenseLabel: '', moraleLabel: '', belligerenceLabel: '', inhabitants: 70, workers: 6, reservesTotal: 8, reservesFree: 42, reservesOverflow: 0, overcrowded: false },
}

let handlers: Record<string, (p: unknown) => void>
const send = vi.fn(() => Promise.resolve())
const socket = { status: 'open', send, subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => { delete handlers[t] } } } as unknown as CampaignSocketApi

function mockFetch() {
  return vi.spyOn(globalThis, 'fetch').mockImplementation((input, init) => {
    const url = String(input)
    let body: unknown = refuge
    if (url.includes('/week/')) body = report(4, { completed: ['Serraria'] })
    else if (url.endsWith('/characters')) body = []
    else if (url.endsWith('/auth/me') || url.endsWith('/me')) body = { id: 'me', email: 'a@x', displayName: 'Ana', isAdmin: false }
    void init
    return Promise.resolve(new Response(JSON.stringify(body), { status: 200 }))
  })
}

const renderIt = (isGm: boolean) => render(
  <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
    <RealtimeContext.Provider value={socket}><RefugeSection campaignId="c1" isGm={isGm} /></RealtimeContext.Provider>
  </QueryClientProvider>,
)

beforeEach(() => { handlers = {}; send.mockClear() })
afterEach(() => vi.restoreAllMocks())

it('catalog has the 20 buildings with costs, leveled cost and material', () => {
  expect(CATALOG).toHaveLength(20)
  expect(new Set(CATALOG.map((d) => d.kind)).size).toBe(20)
  const fortification = defOf('fortification')!
  expect(costAt(fortification, 3)).toBe(30)
  expect(costAt(defOf('quarry')!, 3)).toBe(20)
  expect(materialCost(defOf('quarry')!, 1)).toBe(2)
})

it('allocation helpers sum the workers', () => {
  expect(allocationOf([mill, fort])).toEqual({ b1: 2, b2: 1 })
  expect(totalOf({ a: 2, b: 3 })).toBe(5)
})

it('lists buildings with state, level, progress bar and workers', async () => {
  mockFetch()
  renderIt(false)
  const card = await screen.findByRole('listitem', { name: 'Fortificação' })
  expect(within(card).getByText('Em obra')).toBeInTheDocument()
  expect(within(card).getByText('nível 1 → 2')).toBeInTheDocument()
  expect(within(card).getByRole('progressbar')).toHaveAttribute('aria-valuenow', '5')
  expect(within(card).getByLabelText('Trabalhadores em Fortificação')).toHaveTextContent('1')
  expect(within(screen.getByRole('listitem', { name: 'Moinho' })).getByText('Concluída')).toBeInTheDocument()
})

it('gm picks a card from the catalog dialog and sends add', async () => {
  mockFetch()
  renderIt(true)
  fireEvent.click(await screen.findByText('Nova construção'))
  const card = within(screen.getByRole('dialog')).getByRole('article', { name: 'Pedreira' })
  expect(card).toHaveTextContent('20 Pontos de Obra')
  fireEvent.click(within(card).getByText('Construir'))
  expect(send).toHaveBeenCalledWith('refuge.building.add', { kind: 'quarry' })
})

it('players propose and cannot pick kinds that already exist or were proposed', async () => {
  mockFetch()
  renderIt(false)
  fireEvent.click(await screen.findByText('Nova construção'))
  const dlg = within(screen.getByRole('dialog'))
  expect(within(dlg.getByRole('article', { name: 'Moinho' })).getByRole('button')).toBeDisabled()
  expect(within(dlg.getByRole('article', { name: 'Serraria' })).getByRole('button')).toBeDisabled()
  fireEvent.click(within(dlg.getByRole('article', { name: 'Porto' })).getByText('Propor'))
  expect(send).toHaveBeenCalledWith('refuge.building.propose', { kind: 'port' })
})

it('gm approves or rejects proposals; players only read them', async () => {
  mockFetch()
  const { unmount } = renderIt(false)
  expect(await screen.findByText(/Bia propôs/)).toBeInTheDocument()
  expect(screen.queryByText('Aprovar')).toBeNull()
  unmount()
  renderIt(true)
  fireEvent.click(await screen.findByLabelText('Aprovar Serraria'))
  fireEvent.click(screen.getByLabelText('Recusar Serraria'))
  expect(send).toHaveBeenCalledWith('refuge.building.approve', { id: 'p1' })
  expect(send).toHaveBeenCalledWith('refuge.building.reject', { id: 'p1' })
})

it('allocation stays within the available workers and saves the whole map', async () => {
  mockFetch()
  renderIt(true)
  expect(await screen.findByLabelText('Trabalhadores alocados')).toHaveTextContent('3')
  const plus = screen.getByLabelText('Aumentar Alocar em Moinho')
  fireEvent.click(plus); fireEvent.click(plus); fireEvent.click(plus)
  expect(screen.getByLabelText('Alocar em Moinho')).toHaveTextContent('5')
  expect(plus).toBeDisabled()
  fireEvent.click(screen.getByText('Salvar alocação'))
  expect(send).toHaveBeenCalledWith('refuge.work.allocate', { allocation: { b1: 5, b2: 1 } })
})

it('players see the allocation read-only', async () => {
  mockFetch()
  renderIt(false)
  await screen.findByLabelText('Trabalhadores alocados')
  expect(screen.queryByLabelText('Aumentar Alocar em Moinho')).toBeNull()
  expect(screen.queryByText('Avançar semana')).toBeNull()
})

it('history shows the readable reports, newest first', async () => {
  mockFetch()
  renderIt(false)
  const hist = await screen.findByLabelText('Histórico de semanas')
  const summaries = within(hist).getAllByText(/^Semana \d/)
  expect(summaries.map((s) => s.textContent)).toEqual(['Semana 3 · 1 alerta(s)', 'Semana 2 · 1 alerta(s)'])
  const r = within(hist).getAllByTestId('report-3')[0]
  expect(r).toHaveTextContent('Entrou: 3 Alimento')
  expect(r).toHaveTextContent('Faltou: 2 Madeira')
  expect(r).toHaveTextContent('Faltou Madeira (precisava 6, tinha 4)')
})

it('players get a toast when the week advances', async () => {
  mockFetch()
  renderIt(false)
  await screen.findByText('Nova construção')
  act(() => handlers['refuge.week.advanced']({ report: report(4) }))
  expect(await screen.findByRole('status')).toHaveTextContent('Semana 4 concluída · obras: Moinho')
})

it('gm previews, adjusts (preview refetched) and confirms', async () => {
  const f = mockFetch()
  renderIt(true)
  fireEvent.click(await screen.findByText('Avançar semana'))
  const dlg = await screen.findByRole('dialog', { name: 'Avançar semana' })
  await within(dlg).findByText('Obras concluídas:')
  expect(dlg).toHaveTextContent('Serraria')
  fireEvent.click(within(dlg).getByText('Ajustar números (somados antes do consumo)'))
  fireEvent.click(within(dlg).getByLabelText('Aumentar Ajuste de Água'))
  await waitFor(() => expect(f.mock.calls.filter(([u]) => String(u).endsWith('/week/preview'))).toHaveLength(2))
  const last = f.mock.calls.filter(([u]) => String(u).endsWith('/week/preview')).at(-1)!
  expect(JSON.parse(String(last[1]?.body))).toEqual({ adjust: { water: 1 } })
  expect(f.mock.calls.some(([u]) => String(u).endsWith('/week/confirm'))).toBe(false)
  fireEvent.click(within(dlg).getByText('Confirmar semana'))
  await waitFor(() => expect(f.mock.calls.some(([u]) => String(u).endsWith('/week/confirm'))).toBe(true))
  await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Avançar semana' })).toBeNull())
})
