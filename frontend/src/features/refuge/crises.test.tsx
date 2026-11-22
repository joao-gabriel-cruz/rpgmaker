import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import type { Crisis, Expedition } from './crises'
import { DiseaseAlert } from './DiseaseAlert'
import { RefugeSection } from './RefugeSection'
import { RESOURCE_GROUPS, type Refuge, type ResourceKey } from './refuge'

const resources = Object.fromEntries(RESOURCE_GROUPS.flatMap((g) => g.keys).map((k) => [k, 0])) as Record<ResourceKey, number>
const crisis = (over: Partial<Crisis>): Crisis => ({
  id: 'k1', name: 'Falta de Água', kind: 'shortage', grade: 2, status: 'active', description: 'Faltou Água', week: 3, createdBy: 'system', suggestions: [], ...over,
})
const base: Refuge = {
  campaignId: 'c1', name: 'Refúgio', description: '', notes: '', updatedAt: '', defense: 3, morale: 2, belligerence: 1, mobility: 1, week: 3,
  population: { level: 7, max: 10 }, reservesMax: 10, hasWaterSource: false, resources,
  region: { size: 2, danger: 4, habitation: 0, naturalResources: 1, contamination: 0, travel: 0,
    abundance: { water: 'abundant', plants: 'none', animals: 'none', wood: 'none', minerals: 'none', biomass: 'none' } },
  buildings: [], proposals: [], history: [], crises: [], residents: [], expeditions: [], crisisSettings: { moralePenalty: true }, conditions: [],
  derived: { defenseLabel: '', moraleLabel: '', belligerenceLabel: '', inhabitants: 70, workers: 6, reservesTotal: 0, reservesFree: 10, reservesOverflow: 0, overcrowded: false },
  items: [],
}
const pc = { id: 'p1', name: 'Rita', kind: 'pc', ownerId: 'me', ownerName: 'Ana', sheet: {} }
const npc = { id: 'n1', name: 'Dona Lia', kind: 'npc', ownerId: null, ownerName: '', sheet: {} }

const send = vi.fn(() => Promise.resolve())
let handlers: Record<string, (p: unknown) => void>
const socket = { status: 'open', send, subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => undefined } } as unknown as CampaignSocketApi

function mockFetch(refuge: Refuge) {
  vi.spyOn(globalThis, 'fetch').mockImplementation((input) => {
    const url = String(input)
    const body = url.endsWith('/refuge') ? refuge : url.endsWith('/characters') ? [pc, npc] : { id: 'me', email: 'a@x', displayName: 'Ana', isAdmin: false }
    return Promise.resolve(new Response(JSON.stringify(body), { status: 200 }))
  })
}
const wrap = (ui: React.ReactNode) => (
  <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
    <RealtimeContext.Provider value={socket}>{ui}</RealtimeContext.Provider>
  </QueryClientProvider>
)
const open = (refuge: Refuge, isGm: boolean) => { mockFetch(refuge); return render(wrap(<RefugeSection campaignId="c1" isGm={isGm} />)) }

beforeEach(() => { handlers = {}; send.mockClear() })
afterEach(() => vi.restoreAllMocks())

it('shows crisis cards with status and grade, hiding suggested ones from players', async () => {
  const refuge = { ...base, crises: [crisis({}), crisis({ id: 'k2', name: 'Superlotação', status: 'suggested' }), crisis({ id: 'k3', name: 'Praga', status: 'contained', grade: 5 })] }
  const { unmount } = open(refuge, false)
  const cards = await screen.findAllByTestId('crisis-card')
  expect(cards.map((c) => c.dataset.status)).toEqual(['active', 'contained'])
  expect(screen.getByLabelText('Gravidade de Praga')).toHaveTextContent('Gravidade 5')
  expect(screen.queryByText('Rolar gravidade')).toBeNull()
  unmount()
  open(refuge, true)
  expect(await screen.findAllByTestId('crisis-card')).toHaveLength(3)
  fireEvent.click(screen.getByText('Confirmar crise'))
  expect(send).toHaveBeenCalledWith('refuge.crisis.confirm', { id: 'k2' })
})

it('gm rolls severity, discards one die and applies the migration suggestion', async () => {
  const roll = { rollId: 'r1', faces: [3, 4, 1, 2], discarded: null, symbols: null }
  const sug = { id: 's1', kind: 'migration', text: 'Metade da População migra', status: 'pending' as const }
  open({ ...base, crises: [crisis({ roll, suggestions: [sug] })] }, true)
  fireEvent.click(await screen.findByText('Rolar gravidade'))
  expect(send).toHaveBeenCalledWith('refuge.crisis.roll', { crisisId: 'k1' })
  fireEvent.click(screen.getByLabelText('Descartar dado 3 (face 1)'))
  expect(send).toHaveBeenCalledWith('refuge.crisis.discard', { crisisId: 'k1', index: 2 })
  fireEvent.click(screen.getByLabelText('Aplicar sugestão de Falta de Água'))
  expect(send).toHaveBeenCalledWith('refuge.crisis.apply', { crisisId: 'k1', suggestionId: 's1' })
})

it('players see the dice but cannot discard', async () => {
  open({ ...base, crises: [crisis({ roll: { rollId: 'r1', faces: [3, 4, 1, 2], discarded: null, symbols: null } })] }, false)
  expect(await screen.findByLabelText('Descartar dado 1 (face 3)')).toBeDisabled()
  expect(screen.getByText('O Mestre vai descartar 1 dado.')).toBeInTheDocument()
})

it('gm sends an expedition and spends the Adaptation on an extra resource', async () => {
  const done: Expedition = { id: 'e1', status: 'done', characterIds: ['p1'], leaderName: 'Rita', resource: 'water', regionNote: '', proposedBy: 'gm', week: 3,
    factor: 3, successes: 2, adaptations: 1, pressures: 0, found: 6, discarded: 0, extras: 1, complications: 0, log: ['2 Sucesso(s) x 3 = 6 de Água nas Reservas'] }
  open({ ...base, expeditions: [done] }, true)
  expect(await screen.findByText(/2 Sucesso\(s\) x 3 = 6 de Água/)).toBeInTheDocument()
  fireEvent.click(screen.getByText('Dar extra'))
  expect(send).toHaveBeenCalledWith('refuge.expedition.extra', { id: 'e1', resource: 'food' })
  fireEvent.click(screen.getByText('Enviar expedição'))
  const dialog = screen.getByRole('dialog', { name: 'Enviar expedição' })
  fireEvent.click(within(dialog).getByLabelText('Rita'))
  fireEvent.change(within(dialog).getByLabelText('Recurso procurado'), { target: { value: 'wood' } })
  fireEvent.click(within(dialog).getByText('Enviar'))
  expect(send).toHaveBeenCalledWith('refuge.expedition.start', { characterIds: ['p1'], resource: 'wood', regionNote: '' })
})

it('the leader rolls a ready expedition and a player proposal waits for the Mestre', async () => {
  const ready: Expedition = { id: 'e2', status: 'ready', characterIds: ['p1'], leaderName: 'Rita', resource: 'water', regionNote: '', proposedBy: 'me', week: 3,
    factor: 0, successes: 0, adaptations: 0, pressures: 0, found: 0, discarded: 0, extras: 0, complications: 0, log: [] }
  open({ ...base, expeditions: [ready, { ...ready, id: 'e3', status: 'proposed' }] }, false)
  fireEvent.click(await screen.findByText('Rolar Percepção + Sobrevivência'))
  expect(send).toHaveBeenCalledWith('refuge.expedition.roll', { id: 'e2', instinct: 'perception' })
  expect(screen.getByText('Aguardando aprovação')).toBeInTheDocument()
  expect(screen.queryByText('Aprovar')).toBeNull()
})

it('lists residents with role and the Mestre adds an NPC', async () => {
  open({ ...base, residents: [{ id: 'r1', characterId: 'x', name: 'Seu Zé', role: 'medic', note: 'Cuida dos doentes' }] }, true)
  expect(await screen.findByText('Seu Zé')).toBeInTheDocument()
  expect(screen.getByText(/Médico · Cuida dos doentes/)).toBeInTheDocument()
  await screen.findByRole('option', { name: 'Dona Lia' })
  fireEvent.change(screen.getByLabelText('Função do morador'), { target: { value: 'guard' } })
  fireEvent.click(screen.getByText('Adicionar morador'))
  expect(send).toHaveBeenCalledWith('refuge.resident.add', { characterId: 'n1', role: 'guard', note: '' })
})

it('shows the Doença alert while the condition is active, live', async () => {
  mockFetch(base)
  render(wrap(<DiseaseAlert campaignId="c1" />))
  await waitFor(() => expect(handlers['refuge.updated']).toBeDefined())
  expect(screen.queryByRole('status')).toBeNull()
  handlers['refuge.updated']({ refuge: { ...base, conditions: ['disease'] } })
  expect(await screen.findByRole('status', { name: 'Doença no Refúgio' })).toHaveTextContent('−1 dado')
})
