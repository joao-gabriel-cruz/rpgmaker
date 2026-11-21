import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { RoundTracker } from './RoundTracker'

vi.mock('../../lib/api', () => ({
  api: { get: vi.fn().mockResolvedValue({ active: true, round: 2, phase: 'infected' }) },
}))

function setup(isGm: boolean, send = vi.fn().mockResolvedValue(undefined)) {
  const handlers: Record<string, (p: unknown) => void> = {}
  const api = {
    status: 'open', send,
    subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => {} },
  } as unknown as CampaignSocketApi
  render(<RealtimeContext.Provider value={api}><RoundTracker campaignId="c1" isGm={isGm} /></RealtimeContext.Provider>)
  return { handlers, send }
}

it('shows the fetched state to a player without controls', async () => {
  setup(false)
  expect(await screen.findByText('Rodada 2')).toBeInTheDocument()
  expect(screen.getByText('Infectados')).toHaveAttribute('aria-current', 'step')
  expect(screen.queryByRole('button')).toBeNull()
})

it('updates live on round.updated', async () => {
  const { handlers } = setup(false)
  await screen.findByText('Rodada 2')
  act(() => handlers['round.updated']({ state: { active: true, round: 3, phase: 'end' } }))
  expect(screen.getByText('Rodada 3')).toBeInTheDocument()
  expect(screen.getByText('Fim')).toHaveAttribute('aria-current', 'step')
})

it('lets the Mestre drive the round', async () => {
  const { send } = setup(true)
  await userEvent.click(await screen.findByRole('button', { name: 'Avançar' }))
  expect(send).toHaveBeenCalledWith('round.next')
  await userEvent.click(screen.getByRole('button', { name: 'Voltar' }))
  expect(send).toHaveBeenCalledWith('round.prev')
  await waitFor(() => expect(screen.getByRole('button', { name: 'Encerrar' })).toBeInTheDocument())
})
