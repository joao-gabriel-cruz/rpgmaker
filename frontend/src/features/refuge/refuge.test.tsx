import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { RefugeSection } from './RefugeSection'
import type { Refuge } from './refuge'

const refuge: Refuge = {
  campaignId: 'c1', name: 'Refúgio', description: '', defense: 3, morale: 4, mobility: 5, belligerence: 6,
  water: 8, food: 2, wood: 1, notes: '', updatedAt: '2026-01-01T00:00:00Z',
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

const renderIt = (isGm: boolean) => render(
  <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
    <RealtimeContext.Provider value={socket}><RefugeSection campaignId="c1" isGm={isGm} /></RealtimeContext.Provider>
  </QueryClientProvider>,
)

beforeEach(() => { handlers = {}; send.mockClear() })
afterEach(() => vi.restoreAllMocks())

it('gm lowers water and sends the whole refuge', async () => {
  const f = mockFetch()
  renderIt(true)
  fireEvent.click(await screen.findByLabelText('Diminuir Água'))
  await waitFor(() => expect(f.mock.calls.some(([, i]) => i?.method === 'PUT' && JSON.parse(String(i.body)).water === 7)).toBe(true))
})

it('players see values but have no steppers; live update changes the value', async () => {
  mockFetch()
  renderIt(false)
  expect(await screen.findByLabelText('Água')).toHaveTextContent('8')
  expect(screen.queryByLabelText('Diminuir Água')).toBeNull()
  act(() => handlers['refuge.updated']({ refuge: { ...refuge, water: 6 } }))
  await waitFor(() => expect(screen.getByLabelText('Água')).toHaveTextContent('6'))
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
