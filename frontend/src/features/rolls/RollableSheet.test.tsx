import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import type { RollState } from './modifiers'
import { RollableSheet } from './RollableSheet'

vi.mock('../sheet-assimilacao/AssimilacaoSheet', () => ({
  AssimilacaoSheet: (p: { onRollRequest: (r: unknown) => void }) => (
    <button onClick={() => p.onRollRequest({ instinct: 'reaction', aptitude: 'athletics' })}>rolar</button>
  ),
}))
const getState = vi.fn()
vi.mock('../../lib/api', () => ({ api: { get: (p: string) => getState(p) } }))

const free: RollState = { occupationUsed: false, markingEventUsed: false, sessionActive: true, difficultyD12: 0 }

beforeEach(() => getState.mockResolvedValue(free))

function setup(isGm: boolean) {
  const send = vi.fn().mockResolvedValue(undefined)
  const api = { status: 'open', send, subscribe: () => () => {} } as unknown as CampaignSocketApi
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={qc}>
      <RealtimeContext.Provider value={api}>
        <RollableSheet campaignId="k1" characterId="c1" isGm={isGm} sheet={{} as never} onChange={() => {}} />
      </RealtimeContext.Provider>
    </QueryClientProvider>,
  )
  return send
}

const base = { characterId: 'c1', instinct: 'reaction', aptitude: 'athletics', effort: false, useOccupation: false, useMarkingEvent: false, secret: false }

it('sends roll.request for the character', async () => {
  const send = setup(false)
  expect(screen.queryByLabelText('Rolagem secreta')).toBeNull()
  expect(screen.queryByLabelText('Dificuldade')).toBeNull()
  await userEvent.click(screen.getByText('rolar'))
  expect(send).toHaveBeenCalledWith('roll.request', base)
})

it('lets the Mestre roll in secret', async () => {
  const send = setup(true)
  await userEvent.click(screen.getByLabelText('Rolagem secreta'))
  await userEvent.click(screen.getByText('rolar'))
  expect(send).toHaveBeenCalledWith('roll.request', expect.objectContaining({ secret: true }))
})

it('sends the chosen modifiers', async () => {
  const send = setup(false)
  await screen.findByText('Modificadores')
  await userEvent.click(screen.getByLabelText(/Empenho/))
  await userEvent.click(screen.getByLabelText('Ocupação'))
  await userEvent.click(screen.getByLabelText('Assimilada'))
  await userEvent.selectOptions(screen.getByLabelText('Segundo instinto'), 'potency')
  await userEvent.click(screen.getByText('rolar'))
  expect(send).toHaveBeenCalledWith('roll.request', { ...base, effort: true, useOccupation: true, secondInstinct: 'potency' })
})

it('disables modifiers already used in the session', async () => {
  getState.mockResolvedValue({ ...free, occupationUsed: true, markingEventUsed: true })
  setup(false)
  expect(await screen.findAllByText('(usado nesta sessão)')).toHaveLength(2)
  expect(screen.getByLabelText(/Ocupação/)).toBeDisabled()
  expect(screen.getByLabelText(/Evento Marcante/)).toBeDisabled()
  expect(screen.getByLabelText(/Empenho/)).toBeEnabled()
})

it('disables session-bound modifiers without an active session', async () => {
  getState.mockResolvedValue({ ...free, sessionActive: false })
  setup(false)
  expect((await screen.findAllByText('(requer sessão ativa)')).length).toBe(2)
})

it('lets the Mestre set difficulty and roll with it', async () => {
  const send = setup(true)
  const input = await screen.findByLabelText('Dificuldade')
  await userEvent.clear(input)
  await userEvent.type(input, '2')
  await userEvent.click(screen.getByText('Definir para a próxima rolagem'))
  expect(send).toHaveBeenCalledWith('roll.difficulty.set', { characterId: 'c1', d12: 2 })
  await userEvent.click(screen.getByText('rolar'))
  expect(send).toHaveBeenCalledWith('roll.request', expect.objectContaining({ difficultyD12: 2 }))
})
