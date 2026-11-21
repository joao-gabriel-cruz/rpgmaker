import { act, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { beforeEach, expect, it, vi } from 'vitest'
import { api } from '../../lib/api'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { filterItems, blankItem, type CompendiumItem } from './items'
import { ItemForm } from './ItemForm'
import { ItemsPanel } from './ItemsPanel'
import { lootPlacement } from './lootPlacement'

vi.mock('../../lib/api', () => ({ api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() } }))

const bandage: CompendiumItem = {
  id: 'i1', name: 'Bandagem', description: 'Pano limpo', icon: 'bandage', color: '#ffffff', kind: 'consumable',
  quality: 1, scarcity: 2, pressure: 0, traits: ['médico'], effects: { heal: 2 },
}
const crowbar: CompendiumItem = { ...bandage, id: 'i2', name: 'Pé de cabra', icon: 'wrench', kind: 'tool', traits: [], effects: { keptBonus: { actionCategory: 'exploration', amount: 1 } } }

function wrap(ui: React.ReactNode, send = vi.fn().mockResolvedValue(undefined)) {
  const handlers: Record<string, (p: unknown) => void> = {}
  const socket = { status: 'open', send, subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => {} } } as unknown as CampaignSocketApi
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(<QueryClientProvider client={qc}><RealtimeContext.Provider value={socket}>{ui}</RealtimeContext.Provider></QueryClientProvider>)
  return { handlers, send }
}

beforeEach(() => {
  vi.mocked(api.get).mockImplementation((path: string) =>
    Promise.resolve(path.endsWith('/items') ? { items: [bandage, crowbar] } : [{ id: 'ch1', name: 'Kaio', kind: 'pc' }]))
  vi.mocked(api.post).mockResolvedValue(bandage)
})

it('form validates name, picks icon and color, shows heal only for consumables and previews the card', async () => {
  const onSave = vi.fn().mockResolvedValue(undefined)
  render(<ItemForm initial={blankItem()} onSave={onSave} onCancel={() => {}} />)
  const save = screen.getByRole('button', { name: 'Salvar item' })
  expect(save).toBeDisabled()
  expect(screen.queryByLabelText(/Cura ao consumir/)).not.toBeInTheDocument()
  await userEvent.type(screen.getByLabelText('Nome'), 'Bandagem')
  await userEvent.selectOptions(screen.getByLabelText('Tipo'), 'consumable')
  await userEvent.click(screen.getByRole('button', { name: 'Bandagem' }))
  expect(screen.getByRole('button', { name: 'Bandagem' })).toHaveAttribute('aria-pressed', 'true')
  await userEvent.clear(screen.getByLabelText(/Cura ao consumir/))
  await userEvent.type(screen.getByLabelText(/Cura ao consumir/), '2')
  await userEvent.type(screen.getByLabelText(/Características/), 'médico, limpo')
  expect(within(screen.getByLabelText('Prévia do cartão')).getByText('Bandagem')).toBeInTheDocument()
  expect(within(screen.getByLabelText('Prévia do cartão')).getByText('Cura 2')).toBeInTheDocument()
  await userEvent.click(save)
  expect(onSave).toHaveBeenCalledWith(expect.objectContaining({ name: 'Bandagem', icon: 'bandage', kind: 'consumable', traits: ['médico', 'limpo'], effects: { heal: 2 } }))
})

it('form shows the attack animation only for weapons and surfaces server errors', async () => {
  const onSave = vi.fn().mockRejectedValue(new Error('Ícone inválido'))
  render(<ItemForm initial={{ ...blankItem(), name: 'Faca' }} onSave={onSave} onCancel={() => {}} />)
  expect(screen.queryByLabelText('Animação de ataque')).not.toBeInTheDocument()
  await userEvent.selectOptions(screen.getByLabelText('Tipo'), 'melee')
  await userEvent.selectOptions(screen.getByLabelText('Animação de ataque'), 'slash')
  await userEvent.click(screen.getByRole('button', { name: 'Salvar item' }))
  expect(await screen.findByRole('alert')).toHaveTextContent('Ícone inválido')
  expect(onSave).toHaveBeenCalledWith(expect.objectContaining({ attackAnimation: 'slash' }))
})

it('filters by name and kind', () => {
  expect(filterItems([bandage, crowbar], 'pé', '')).toEqual([crowbar])
  expect(filterItems([bandage, crowbar], '', 'consumable')).toEqual([bandage])
  expect(filterItems([bandage, crowbar], 'médico', 'tool')).toEqual([])
})

it('panel lists the compendium, filters and refreshes on items.changed', async () => {
  const { handlers } = wrap(<ItemsPanel campaignId="c1" />)
  expect(await screen.findByText('Pé de cabra')).toBeInTheDocument()
  await userEvent.type(screen.getByLabelText('Buscar item'), 'band')
  expect(screen.queryByText('Pé de cabra')).not.toBeInTheDocument()
  act(() => handlers['items.changed']({ items: [{ ...bandage, name: 'Bandagem forte' }] }))
  expect(await screen.findByText('Bandagem forte')).toBeInTheDocument()
})

it('gives an item to a character choosing the backpack', async () => {
  const { send } = wrap(<ItemsPanel campaignId="c1" />)
  await userEvent.click(await screen.findByRole('button', { name: 'Entregar Bandagem' }))
  await waitFor(() => expect(screen.getByRole('option', { name: 'Kaio' })).toBeInTheDocument())
  await userEvent.click(screen.getByRole('button', { name: 'Entregar' }))
  expect(send).toHaveBeenCalledWith('item.give', { itemId: 'i1', characterId: 'ch1', place: 'backpack', qty: 1 })
})

it('gives to the Refúgio without a place and arms the map placement with the difficulty', async () => {
  const { send } = wrap(<ItemsPanel campaignId="c1" />)
  await userEvent.click(await screen.findByRole('button', { name: 'Entregar Pé de cabra' }))
  await waitFor(() => screen.getByRole('option', { name: 'Kaio' }))
  await userEvent.selectOptions(screen.getByLabelText('Para'), 'refuge')
  expect(screen.queryByLabelText('Onde')).not.toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Entregar' }))
  expect(send).toHaveBeenCalledWith('item.give', { itemId: 'i2', target: 'refuge', qty: 1 })
  await userEvent.click(screen.getByRole('button', { name: 'Esconder Pé de cabra no mapa' }))
  await userEvent.clear(screen.getByLabelText(/Dificuldade/))
  await userEvent.type(screen.getByLabelText(/Dificuldade/), '2')
  await userEvent.click(screen.getByRole('button', { name: 'Escolher no mapa' }))
  expect(lootPlacement.get()).toEqual({ itemId: 'i2', name: 'Pé de cabra', difficulty: 2 })
  expect(screen.getByRole('status')).toHaveTextContent(/dificuldade 2/)
  lootPlacement.cancel()
})
