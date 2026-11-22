import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, expect, it, vi } from 'vitest'
import { TableCharacterProvider } from '../hotbar/TableCharacterContext'
import { vitalsOf } from '../hotbar/vitals'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { CharactersSection } from './CharactersSection'
import type { CharacterView } from './characters'

const socket = { status: 'open', send: vi.fn(), subscribe: () => () => undefined } as unknown as CampaignSocketApi
const partial = (id: string, kind: 'pc' | 'npc', name: string, sheet: object) =>
  ({ id, campaignId: 'c1', ownerId: null, ownerName: '', kind, name, sheet, updatedAt: '', issues: [] }) as unknown as CharacterView

afterEach(() => vi.restoreAllMocks())

it('vitalsOf falls back to defaults for missing sheet fields', () => {
  for (const sheet of [{}, { health: {} }, { instincts: {}, tugOfWar: {} }, null]) {
    const v = vitalsOf(partial('a', 'pc', 'Rita', sheet as object))
    expect(v.damage).toBe(0)
    expect(v.boxes).toBe(1)
    expect(v.levelName).toBeTruthy()
  }
  expect(vitalsOf(partial('b', 'npc', 'Lia', {})).boxes).toBe(1)
  expect(vitalsOf(partial('a', 'pc', 'Rita', {})).tug).toBeUndefined()
})

it('the characters section renders PCs and NPCs whose sheet lacks fields', async () => {
  const list = [partial('a', 'pc', 'Rita', {}), partial('b', 'pc', 'Davi', { health: { damage: 3 } }), partial('c', 'npc', 'Dona Lia', {})]
  vi.spyOn(globalThis, 'fetch').mockImplementation(() => Promise.resolve(new Response(JSON.stringify(list), { status: 200 })))
  render(
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
      <MemoryRouter><RealtimeContext.Provider value={socket}><TableCharacterProvider>
        <CharactersSection campaignId="c1" isGm />
      </TableCharacterProvider></RealtimeContext.Provider></MemoryRouter>
    </QueryClientProvider>,
  )
  expect(await screen.findByText('Rita')).toBeInTheDocument()
  expect(screen.getByText('Davi')).toBeInTheDocument()
  expect(screen.getByText('Dona Lia')).toBeInTheDocument()
})
