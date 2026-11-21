import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { RollLog } from './RollLog'
import { RollTray } from './RollTray'
import type { RollEntry } from './rolls'

const trayProps = vi.fn()
vi.mock('../dice3d/DiceTray', () => ({
  DiceTray: (p: { dice: unknown[]; selected?: number[]; onToggle?: (i: number) => void; onSettled?: () => void }) => {
    trayProps(p)
    return (
      <div data-testid="tray">
        <button onClick={p.onSettled}>settle</button>
        <button onClick={() => p.onToggle?.(0)}>toggle0</button>
        <button onClick={() => p.onToggle?.(2)}>toggle2</button>
      </div>
    )
  },
}))
vi.mock('../auth/auth', () => ({ useMe: () => ({ data: { id: 'u1' } }) }))
const getRolls = vi.fn()
vi.mock('../../lib/api', () => ({ api: { get: (p: string) => getRolls(p) } }))

const entry: RollEntry = {
  rollId: 'r1', userId: 'u1', userName: 'Ana', characterId: 'c1', characterName: 'Rita', kind: 'assimilacao', visibility: 'public',
  dice: [{ kind: 'd6', face: 5, difficulty: false }, { kind: 'd6', face: 2, difficulty: false }, { kind: 'd10', face: 6, difficulty: false }],
  keptCount: 1, suggested: [2], request: { instinct: 'reaction', aptitude: 'athletics' }, seed: 99,
}

function setup(ui: React.ReactNode, send = vi.fn().mockResolvedValue(undefined)) {
  const handlers: Record<string, (p: unknown) => void> = {}
  const api = { status: 'open', send, subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => {} } } as unknown as CampaignSocketApi
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(<QueryClientProvider client={qc}><RealtimeContext.Provider value={api}>{ui}</RealtimeContext.Provider></QueryClientProvider>)
  const emit = (t: string, p: unknown) => act(() => handlers[t](p))
  return { send, emit }
}

beforeEach(() => { trayProps.mockClear(); getRolls.mockResolvedValue([]) })

it('hands the tray exactly the faces of the event', () => {
  const { emit } = setup(<RollTray />)
  emit('roll.rolled', entry)
  const p = trayProps.mock.calls.at(-1)![0]
  expect(p.dice).toEqual([{ kind: 'd6', face: 5, variant: 'normal' }, { kind: 'd6', face: 2, variant: 'normal' }, { kind: 'd10', face: 6, variant: 'normal' }])
  expect(p).toMatchObject({ seed: 99, rollId: 'r1' })
})

it('lets the roller pick the dice, preselects the suggestion and confirms', async () => {
  const { emit, send } = setup(<RollTray />)
  emit('roll.rolled', entry)
  await userEvent.click(screen.getByText('settle'))
  expect(screen.getByText('Escolha 1 dado (1/1)')).toBeInTheDocument()
  await userEvent.click(screen.getByText('toggle2'))
  await userEvent.click(screen.getByText('toggle0'))
  expect(screen.getByRole('button', { name: 'Confirmar' })).toBeEnabled()
  await userEvent.click(screen.getByText('toggle0'))
  await userEvent.click(screen.getByText('toggle2'))
  await userEvent.click(screen.getByRole('button', { name: 'Confirmar' }))
  expect(send).toHaveBeenCalledWith('roll.keep', { rollId: 'r1', kept: [2] })
})

it('disables confirm until the right number of dice is selected', async () => {
  const { emit } = setup(<RollTray />)
  emit('roll.rolled', entry)
  await userEvent.click(screen.getByText('settle'))
  await userEvent.click(screen.getByText('toggle2'))
  expect(screen.getByRole('button', { name: 'Confirmar' })).toBeDisabled()
})

it('shows others "is choosing" and then the result with the penalty', () => {
  const { emit } = setup(<RollTray />)
  emit('roll.rolled', { ...entry, userId: 'u2', userName: 'Bea' })
  expect(screen.getByText('Bea está escolhendo…')).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Confirmar' })).toBeNull()
  emit('roll.finalized', { rollId: 'r1', result: { symbols: { success: 1, adaptation: 0, pressure: 2 }, kept: [2], penalty: 1, assimilated: false, healthLevel: 'Laceração' } })
  expect(screen.getByText('−1 Sucesso (Laceração)')).toBeInTheDocument()
  expect(screen.getByRole('status', { name: 'Resultado da rolagem' })).toHaveTextContent('2Pressões')
})

const finalized = { rollId: 'r1', result: { symbols: { success: 1, adaptation: 0, pressure: 0 }, kept: [2], penalty: 0, assimilated: false } }

it('closes by itself 4 seconds after the result for someone who did not roll', async () => {
  vi.useFakeTimers({ shouldAdvanceTime: true })
  const { emit } = setup(<RollTray />)
  emit('roll.rolled', { ...entry, userId: 'u2', userName: 'Bea' })
  act(() => void vi.advanceTimersByTime(10000))
  expect(screen.getByLabelText('Bandeja de dados')).toBeInTheDocument()
  emit('roll.finalized', finalized)
  act(() => void vi.advanceTimersByTime(3900))
  expect(screen.getByLabelText('Bandeja de dados')).toBeInTheDocument()
  act(() => void vi.advanceTimersByTime(200))
  expect(screen.queryByLabelText('Bandeja de dados')).toBeNull()
  vi.useRealTimers()
})

it('keeps the result for the roller until Esc, and queues further rolls', async () => {
  vi.useFakeTimers({ shouldAdvanceTime: true })
  const { emit } = setup(<RollTray />)
  emit('roll.rolled', entry)
  emit('roll.rolled', { ...entry, rollId: 'r2', userId: 'u2', userName: 'Bea' })
  expect(screen.getByText('+1 na fila')).toBeInTheDocument()
  await userEvent.keyboard('{Escape}')
  expect(screen.getByText('Ana · Rita')).toBeInTheDocument()
  emit('roll.finalized', finalized)
  act(() => void vi.advanceTimersByTime(10000))
  expect(screen.getByText('Ana · Rita')).toBeInTheDocument()
  await userEvent.keyboard('{Escape}')
  expect(screen.getByRole('heading', { name: /Bea/ })).toBeInTheDocument()
  expect(screen.queryByText('+1 na fila')).toBeNull()
  vi.useRealTimers()
})

it('labels the pool and the result in the log', async () => {
  const { emit } = setup(<RollLog campaignId="c1" isGm={false} />)
  emit('roll.rolled', entry)
  emit('roll.finalized', { rollId: 'r1', result: { symbols: { success: 2, adaptation: 1, pressure: 0 }, kept: [2], penalty: 0, assimilated: false } })
  expect(screen.getByText('Reação 2 + Atletismo 1')).toBeInTheDocument()
  expect(screen.getByText(/Sucessos: 2/)).toBeInTheDocument()
  expect(screen.getByText(/Rita/)).toBeInTheDocument()
})

it('shows secret rolls only to the Mestre', async () => {
  const secret = { ...entry, rollId: 'r2', visibility: 'gm' as const }
  getRolls.mockResolvedValue([secret])
  const player = setup(<RollLog campaignId="c1" isGm={false} />)
  player.emit('roll.rolled', { ...secret, rollId: 'r3' })
  expect(screen.queryByText('Secreta')).toBeNull()
  expect(await screen.findByText('Nenhuma rolagem ainda.')).toBeInTheDocument()
})

it('marks secret rolls for the Mestre', async () => {
  getRolls.mockResolvedValue([{ ...entry, rollId: 'r2', visibility: 'gm' }])
  setup(<RollLog campaignId="c1" isGm />)
  expect(await screen.findByText('Secreta')).toBeInTheDocument()
})

it('has no text box for rolls: the /r command left the interface', () => {
  setup(<RollLog campaignId="c1" isGm={false} />)
  expect(screen.queryByLabelText('Comando de rolagem')).toBeNull()
  expect(screen.queryByPlaceholderText('/r 2d6+3')).toBeNull()
  expect(screen.queryByRole('textbox')).toBeNull()
})

it('puts generic rolls through the 3D tray with the total once the dice settle', async () => {
  const { emit } = setup(<RollTray />)
  emit('roll.rolled', { rollId: 'g1', userId: 'u2', userName: 'Bob', characterId: null, kind: 'generic', visibility: 'public',
    dice: [{ kind: 'd20', face: 17, difficulty: false }], expression: 'd20', total: 17, seed: 4 })
  expect(trayProps.mock.calls.at(-1)![0].dice).toEqual([{ kind: 'd20', face: 17, variant: 'normal' }])
  expect(screen.queryByTestId('generic-total')).toBeNull()
  await userEvent.click(screen.getByText('settle'))
  expect(screen.getByTestId('generic-total')).toHaveTextContent('d20: 17')
  expect(screen.queryByRole('button', { name: 'Confirmar' })).toBeNull()
})

it('skips the tray for generic rolls with dice the renderer lacks', () => {
  const { emit } = setup(<RollTray />)
  emit('roll.rolled', { rollId: 'g2', userId: 'u1', userName: 'Ana', characterId: null, kind: 'generic', visibility: 'public',
    dice: [{ kind: 'd100', face: 50, difficulty: false }], expression: 'd100', total: 50, seed: 1 })
  expect(screen.queryByTestId('tray')).toBeNull()
})

it('gives the Mestre a button-driven pool builder that rolls through roll.pool', async () => {
  const { send } = setup(<RollLog campaignId="c1" isGm />)
  const roll = screen.getByRole('button', { name: 'Rolar parada' })
  expect(roll).toBeDisabled()
  await userEvent.click(screen.getByRole('button', { name: 'Adicionar d6' }))
  await userEvent.click(screen.getByRole('button', { name: 'Adicionar d6' }))
  await userEvent.click(screen.getByRole('button', { name: 'Adicionar d12' }))
  expect(screen.getByTestId('pool-builder-text')).toHaveTextContent('2d6 + 1d12')
  await userEvent.click(roll)
  expect(send).toHaveBeenCalledWith('roll.pool', { d6: 2, d10: 0, d12: 1 })
  expect(await screen.findByText('Nenhum dado ainda')).toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Adicionar d10' }))
  await userEvent.click(screen.getByRole('button', { name: 'Limpar' }))
  expect(screen.getByText('Nenhum dado ainda')).toBeInTheDocument()
})

it('offers the Mestre generic shortcuts (d6, 2d6, d10, d12, d20) and hides the tools from players', async () => {
  const { send } = setup(<RollLog campaignId="c1" isGm />)
  await userEvent.click(screen.getByRole('button', { name: 'Rolar 2d6' }))
  await userEvent.click(screen.getByRole('button', { name: 'Rolar d20' }))
  expect(send).toHaveBeenCalledWith('roll.generic', { expression: '2d6' })
  expect(send).toHaveBeenCalledWith('roll.generic', { expression: 'd20' })
  expect(screen.getAllByRole('button', { name: /^Rolar (d6|2d6|d10|d12|d20)$/ })).toHaveLength(5)
})

it('keeps the pool builder and loose dice away from players', () => {
  setup(<RollLog campaignId="c1" isGm={false} />)
  expect(screen.queryByRole('button', { name: 'Rolar parada' })).toBeNull()
  expect(screen.queryByRole('button', { name: 'Rolar d20' })).toBeNull()
})

it('renders generic rolls with the total', async () => {
  const { emit } = setup(<RollLog campaignId="c1" isGm={false} />)
  emit('roll.rolled', { rollId: 'g1', userId: 'u1', userName: 'Ana', characterId: null, kind: 'generic', visibility: 'public',
    dice: [{ kind: 'd6', face: 2, difficulty: false }, { kind: 'd6', face: 3, difficulty: false }], expression: '2d6+3', total: 8, seed: 1 })
  expect(screen.getByText('2d6+3')).toBeInTheDocument()
  expect(screen.getByText('8')).toBeInTheDocument()
})

it('shows the modifiers used in the log and the tray', async () => {
  const modded = { ...entry, request: { ...entry.request!, effort: true, useOccupation: true, difficultyD12: 2 } }
  getRolls.mockResolvedValue([modded])
  setup(<><RollLog campaignId="k" isGm={false} /><RollTray /></>)
  expect(await screen.findByText('Empenho, Ocupação, +2 d12 dificuldade')).toBeInTheDocument()
})
