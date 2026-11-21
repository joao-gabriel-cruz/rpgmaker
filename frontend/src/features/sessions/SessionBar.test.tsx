import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { SessionBar } from './SessionBar'

vi.mock('../../lib/api', () => ({ api: { get: () => Promise.resolve({ active: false, startedAt: null }) } }))

function setup(isGm: boolean) {
  const send = vi.fn().mockResolvedValue(undefined)
  const handlers: Record<string, (p: unknown) => void> = {}
  const api = { status: 'open', send, subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => {} } } as unknown as CampaignSocketApi
  render(<RealtimeContext.Provider value={api}><SessionBar campaignId="k" isGm={isGm} /></RealtimeContext.Provider>)
  return { send, emit: (p: unknown) => act(() => handlers['session.updated'](p)) }
}

it('lets the Mestre start and end, and shows the indicator', async () => {
  const { send, emit } = setup(true)
  expect(await screen.findByText('Sem sessão ativa')).toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Iniciar sessão' }))
  expect(send).toHaveBeenCalledWith('session.start')
  emit({ active: true, startedAt: '2026-01-01T00:00:00Z' })
  expect(screen.getByText('Sessão ativa')).toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Encerrar sessão' }))
  expect(send).toHaveBeenCalledWith('session.end')
})

it('shows only the indicator to players', async () => {
  const { emit } = setup(false)
  emit({ active: true, startedAt: '2026-01-01T00:00:00Z' })
  expect(screen.getByText('Sessão ativa')).toBeInTheDocument()
  expect(screen.queryByRole('button')).toBeNull()
})
