import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, expect, it, vi } from 'vitest'
import { api } from '../../lib/api'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { AppliedToast } from './AppliedToast'
import { PendingEffects } from './PendingEffects'
import { RulesEditor } from './RulesEditor'
import type { RuleSet, Suggestion } from './effects'

vi.mock('../../lib/api', () => ({ api: { get: vi.fn(), put: vi.fn(), delete: vi.fn() } }))

const sug: Suggestion = {
  id: 's1', campaignId: 'c1', characterId: 'ch', characterName: 'Ana', sceneId: 'sc', x: 1, z: 2,
  source: 'Lava', trigger: 'on_enter', effect: 'damage', amount: 2, message: 'A lava queima!', createdAt: '',
}

function wrap(ui: React.ReactNode, send = vi.fn().mockResolvedValue(undefined)) {
  const handlers: Record<string, (p: unknown) => void> = {}
  const socket = {
    status: 'open', send,
    subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => {} },
  } as unknown as CampaignSocketApi
  render(<RealtimeContext.Provider value={socket}>{ui}</RealtimeContext.Provider>)
  return { handlers, send }
}

beforeEach(() => vi.mocked(api.get).mockResolvedValue({ suggestions: [], rules: [] }))

it('lists a live suggestion and counts it', async () => {
  const { handlers } = wrap(<PendingEffects campaignId="c1" />)
  expect(screen.getByTestId('pending-count')).toHaveTextContent('0')
  act(() => handlers['effect.suggested']({ suggestion: sug }))
  expect(screen.getByText(/Lava · Ao entrar/)).toBeInTheDocument()
  expect(screen.getByText(/\+2 de dano/)).toBeInTheDocument()
  expect(screen.getByTestId('pending-count')).toHaveTextContent('1')
})

it('loads the pending list on mount', async () => {
  vi.mocked(api.get).mockResolvedValue({ suggestions: [sug] })
  wrap(<PendingEffects campaignId="c1" />)
  expect(await screen.findByText(/Lava · Ao entrar/)).toBeInTheDocument()
})

it('applies and dismisses through commands and clears on effect.resolved', async () => {
  const { handlers, send } = wrap(<PendingEffects campaignId="c1" />)
  act(() => handlers['effect.suggested']({ suggestion: sug }))
  await userEvent.click(screen.getByRole('button', { name: 'Aplicar efeito em Ana' }))
  expect(send).toHaveBeenCalledWith('effect.confirm', { id: 's1' })
  await userEvent.click(screen.getByRole('button', { name: 'Descartar efeito de Ana' }))
  expect(send).toHaveBeenCalledWith('effect.dismiss', { id: 's1' })
  act(() => handlers['effect.resolved']({ id: 's1', applied: true }))
  expect(screen.getByTestId('pending-count')).toHaveTextContent('0')
})

it('toasts the player when an effect is applied', () => {
  const { handlers } = wrap(<AppliedToast />)
  act(() => handlers['effect.applied']({ characterName: 'Ana', source: 'Lava', effect: 'damage', amount: 2, message: '' }))
  expect(screen.getByRole('status')).toHaveTextContent('Lava: Ana sofreu +2 de dano')
})

it('edits and saves the rules of a terrain', async () => {
  const lava: RuleSet = { kind: 'terrain', key: 'lava', label: 'Lava', isDefault: true, blocksMovement: false, blocksVision: false,
    rules: [{ trigger: 'on_enter', effect: 'damage', amount: 2, message: '' }] }
  vi.mocked(api.get).mockResolvedValue({ rules: [lava] })
  vi.mocked(api.put).mockResolvedValue({ ...lava, isDefault: false })
  wrap(<RulesEditor campaignId="c1" />)
  const amount = await screen.findByLabelText('Quantidade Lava 1')
  await userEvent.clear(amount)
  await userEvent.type(amount, '5')
  await userEvent.click(screen.getByRole('button', { name: 'Salvar Lava' }))
  await waitFor(() => expect(api.put).toHaveBeenCalledWith('/campaigns/c1/effects/rules/terrain/lava',
    { rules: [{ trigger: 'on_enter', effect: 'damage', amount: 5, message: '' }] }))
  expect(await screen.findByRole('button', { name: 'Restaurar padrão Lava' })).toBeInTheDocument()
})
