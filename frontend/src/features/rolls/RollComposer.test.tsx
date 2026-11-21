import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, expect, it, vi } from 'vitest'
import type { CharacterView } from '../characters/characters'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { newSheet } from '../sheet-assimilacao/types'
import type { RollState } from './modifiers'
import { poolPreview } from './pool'
import { NO_MODIFIERS } from './modifiers'
import { RollComposer, type RollComposerProps } from './RollComposer'

const getState = vi.fn()
vi.mock('../../lib/api', () => ({ api: { get: (p: string) => getState(p) } }))

const free: RollState = { occupationUsed: false, markingEventUsed: false, sessionActive: true, difficultyD12: 0 }
const sheet = newSheet()
sheet.instincts.reaction = 2
sheet.instincts.potency = 3
sheet.practices.athletics = 1
sheet.practices.weapons = 2
const character = { id: 'c1', campaignId: 'k1', ownerId: 'u1', ownerName: 'Ana', kind: 'pc', name: 'Rita', sheet, issues: [], updatedAt: '' } as CharacterView

beforeEach(() => getState.mockResolvedValue(free))

function setup(over: Partial<RollComposerProps> = {}, send = vi.fn().mockResolvedValue(undefined)) {
  const onClose = vi.fn()
  const api = { status: 'open', send, subscribe: () => () => {} } as unknown as CampaignSocketApi
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={qc}>
      <RealtimeContext.Provider value={api}>
        <RollComposer campaignId="k1" character={character} isGm={false} onClose={onClose} {...over} />
      </RealtimeContext.Provider>
    </QueryClientProvider>,
  )
  return { send, onClose }
}

const pick = (name: RegExp) => userEvent.click(screen.getByRole('button', { name }))
const base = { characterId: 'c1', instinct: 'reaction', aptitude: 'athletics', effort: false, useOccupation: false, useMarkingEvent: false, secret: false }

it('computes the pool preview like the server does', () => {
  const m = NO_MODIFIERS
  expect(poolPreview(sheet, { instinct: 'reaction', aptitude: 'weapons' }, m, false).text).toBe('2 d6 + 2 d10 · manter 1')
  expect(poolPreview(sheet, { instinct: 'reaction' }, { ...m, effort: true, useOccupation: true }, false).kept).toBe(2)
  const assimilated = poolPreview(sheet, { instinct: 'reaction', aptitude: 'athletics' }, { ...m, assimilated: true, secondInstinct: 'potency' }, false)
  expect(assimilated.text).toBe('2 d12 + 3 d12 + 1 d10 · manter 2')
  expect(poolPreview(sheet, { instinct: 'reaction' }, { ...m, difficultyD12: 2 }, true).text).toBe('2 d6 (+2 d12 dificuldade) · manter 1')
})

it('shows each value with its dice and updates the preview live', async () => {
  setup()
  expect(screen.getByTestId('pool-preview')).toHaveTextContent('Escolha um Instinto')
  expect(screen.getByRole('button', { name: /^Reação 2, 2 d6/ })).toBeInTheDocument()
  await pick(/^Reação/)
  await pick(/^Armas/)
  expect(screen.getByTestId('pool-preview')).toHaveTextContent('2 d6 + 2 d10 · manter 1')
  await userEvent.click(screen.getByLabelText('Empenho', { exact: false }))
  expect(screen.getByTestId('pool-preview')).toHaveTextContent('2 d6 + 2 d10 · manter 2')
})

it('sends roll.request with the usual payload and closes', async () => {
  const onRoll = vi.fn()
  const { send, onClose } = setup({ onRoll })
  await pick(/^Reação/)
  await pick(/^Atletismo/)
  await userEvent.click(screen.getByRole('button', { name: 'Rolar' }))
  expect(send).toHaveBeenCalledWith('roll.request', base)
  await vi.waitFor(() => expect(onClose).toHaveBeenCalled())
  expect(onRoll).toHaveBeenCalledWith(base)
})

it('starts from the preset and lets an alternative replace it', async () => {
  const { send } = setup({ preset: { instinct: 'reaction', aptitude: 'athletics' }, title: 'Atacar',
    alternatives: [{ instinct: 'potency', aptitude: 'weapons', label: 'Na força bruta' }] })
  expect(screen.getByRole('dialog', { name: 'Atacar' })).toBeInTheDocument()
  expect(screen.getByRole('button', { name: /^Reação/ })).toHaveAttribute('aria-pressed', 'true')
  await pick(/Na força bruta/)
  expect(screen.getByRole('button', { name: /^Potência/ })).toHaveAttribute('aria-pressed', 'true')
  await userEvent.click(screen.getByRole('button', { name: 'Rolar' }))
  expect(send).toHaveBeenCalledWith('roll.request', { ...base, instinct: 'potency', aptitude: 'weapons' })
})

it('sends modifiers and requires the second instinct when assimilated', async () => {
  const { send } = setup()
  await pick(/^Reação/)
  await userEvent.click(screen.getByLabelText('Empenho', { exact: false }))
  await userEvent.click(screen.getByLabelText('Assimilada'))
  expect(screen.getByRole('button', { name: 'Rolar' })).toBeDisabled()
  await userEvent.selectOptions(screen.getByLabelText('Segundo instinto'), 'potency')
  await userEvent.click(screen.getByRole('button', { name: 'Rolar' }))
  expect(send).toHaveBeenCalledWith('roll.request', { characterId: 'c1', instinct: 'reaction', effort: true, useOccupation: false, useMarkingEvent: false, secondInstinct: 'potency', secret: false })
})

it('blocks a spent Evento Marcante and tells why', async () => {
  getState.mockResolvedValue({ ...free, markingEventUsed: true })
  setup()
  expect(await screen.findByText('(usado nesta sessão)')).toBeInTheDocument()
  expect(screen.getByLabelText('Evento Marcante', { exact: false })).toBeDisabled()
})

it('lets the Mestre set the difficulty and roll in secret', async () => {
  const { send } = setup({ isGm: true })
  await pick(/^Reação/)
  await userEvent.type(screen.getByLabelText('Dificuldade'), '{Backspace}2')
  await userEvent.click(screen.getByLabelText('Rolagem secreta'))
  await userEvent.click(screen.getByRole('button', { name: 'Rolar' }))
  expect(send).toHaveBeenCalledWith('roll.request', expect.objectContaining({ instinct: 'reaction', difficultyD12: 2, secret: true }))
})

it('is keyboard driven: arrows move focus, Enter rolls, Esc closes', async () => {
  const { send, onClose } = setup({ preset: { instinct: 'reaction' } })
  expect(screen.getByRole('button', { name: /^Reação/ })).toHaveFocus()
  await userEvent.keyboard('{ArrowRight}')
  expect(screen.getByRole('button', { name: /^Resolução/ })).toHaveFocus()
  await userEvent.keyboard('{Enter}')
  expect(send).toHaveBeenCalledWith('roll.request', expect.objectContaining({ instinct: 'reaction' }))
  await userEvent.keyboard('{Escape}')
  expect(onClose).toHaveBeenCalled()
})
