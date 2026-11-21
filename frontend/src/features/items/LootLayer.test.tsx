import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { beforeEach, expect, it, vi } from 'vitest'
import { api } from '../../lib/api'
import { createGrid } from '../map3d/grid'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import type { Token } from '../tokens/tokens'
import { LootLayer } from './LootLayer'
import { objectAt, pickerFor, useSceneLoot, visibleLoot, type Loot } from './loot'

vi.mock('../../lib/api', () => ({ api: { get: vi.fn() } }))
vi.mock('@react-three/drei', () => ({ Html: ({ children }: { children: ReactNode }) => <div>{children}</div> }))
vi.mock('@react-three/fiber', () => ({ useThree: (sel: (s: unknown) => unknown) => sel({ gl: { domElement: document.createElement('canvas') }, camera: {} }) }))

const item = { name: 'Pé de cabra', icon: 'wrench', color: '#aabbcc' }
const hiddenLoot: Loot = { id: 'l1', sceneId: 's1', x: 2, z: 2, objectId: 'o1', itemId: 'i1', item, qty: 1, difficulty: 2, hidden: true, foundBy: null }
const openLoot: Loot = { ...hiddenLoot, id: 'l2', x: 3, z: 2, hidden: false, difficulty: 0 }
const mine = { id: 't1', characterId: 'ch1', x: 3, z: 3, ownerId: 'u1' } as Token
const far = { ...mine, id: 't2', x: 9, z: 9 } as Token
const grid = createGrid(12, 12)

function Harness({ isGm, tokens, userId }: { isGm: boolean; tokens: Token[]; userId: string }) {
  const loot = useSceneLoot('c1', 's1')
  return <LootLayer grid={grid} objects={[]} tokens={tokens} loot={loot} userId={userId} isGm={isGm} />
}

function mount(isGm: boolean, tokens: Token[], send = vi.fn().mockResolvedValue(undefined)) {
  const handlers: Record<string, (p: unknown) => void> = {}
  const socket = { status: 'open', send, subscribe: (t: string, h: (p: unknown) => void) => { handlers[t] = h; return () => {} } } as unknown as CampaignSocketApi
  render(<RealtimeContext.Provider value={socket}><Harness isGm={isGm} tokens={tokens} userId="u1" /></RealtimeContext.Provider>)
  return { handlers, send }
}

beforeEach(() => vi.mocked(api.get).mockResolvedValue({ loot: [hiddenLoot, openLoot] }))

it('visibleLoot hides hidden loot from players only', () => {
  expect(visibleLoot([hiddenLoot, openLoot], true)).toHaveLength(2)
  expect(visibleLoot([hiddenLoot, openLoot], false)).toEqual([openLoot])
})

it('the Mestre sees hidden loot ghosted with its difficulty', async () => {
  mount(true, [])
  const marker = await screen.findByTestId('loot-l1')
  expect(marker).toHaveAttribute('data-hidden', 'true')
  expect(marker).toHaveTextContent('Dif. 2')
  expect(screen.getByTestId('loot-l2')).toHaveAttribute('data-hidden', 'false')
})

it('players only see revealed loot and pick it up when adjacent', async () => {
  const { send } = mount(false, [mine])
  await screen.findByTestId('loot-l2')
  expect(screen.queryByTestId('loot-l1')).not.toBeInTheDocument()
  expect(screen.queryByText(/Dif\./)).not.toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Pegar Pé de cabra' }))
  expect(send).toHaveBeenCalledWith('loot.pickup', { lootId: 'l2', characterId: 'ch1' })
})

it('a player far from the loot cannot pick it up', async () => {
  mount(false, [far])
  await screen.findByTestId('loot-l2')
  expect(screen.getByRole('button', { name: 'Pé de cabra' })).toBeDisabled()
  expect(pickerFor(openLoot, [far], 'u1')).toBeUndefined()
  expect(pickerFor(hiddenLoot, [mine], 'u1')).toBeUndefined()
})

it('loot.found reveals the leftover for everyone and flashes who found it', async () => {
  const { handlers } = mount(false, [])
  await screen.findByTestId('loot-l2')
  act(() => handlers['loot.found']({ loot: { ...hiddenLoot, hidden: false }, characterName: 'Kaio', left: 1, item }))
  expect(screen.getByTestId('loot-l1')).toHaveAttribute('data-hidden', 'false')
  expect(screen.getByText('Kaio achou Pé de cabra!')).toBeInTheDocument()
  act(() => handlers['loot.found']({ loot: hiddenLoot, characterName: 'Kaio', left: 0, item }))
  expect(screen.queryByTestId('loot-l1')).not.toBeInTheDocument()
})

it('objectAt resolves the object covering a cell', () => {
  const crate = { id: 'o1', prefab: 'crate', x: 4, z: 4, rotation: 0, props: {} } as never
  expect(objectAt([crate], 4, 4)?.id).toBe('o1')
  expect(objectAt([crate], 5, 4)).toBeUndefined()
})
