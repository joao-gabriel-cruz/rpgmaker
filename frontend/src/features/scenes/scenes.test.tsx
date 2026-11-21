import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'
import { createGrid, serializeGrid } from '../map3d/grid'
import { SceneSelector } from './SceneSelector'
import { applyRemoteObjects, applyRemotePaint } from './remote'
import type { MapState, SceneSummary } from './scenes'

afterEach(() => vi.restoreAllMocks())

const scenes: SceneSummary[] = [
  { id: 'a', name: 'Floresta', width: 32, depth: 32, isActive: true },
  { id: 'b', name: 'Caverna', width: 16, depth: 16, isActive: false },
]

function setup() {
  const h = { onSelect: vi.fn(), onCreate: vi.fn(), onRename: vi.fn(), onDelete: vi.fn(), onActivate: vi.fn() }
  render(<SceneSelector scenes={scenes} selectedId="a" {...h} />)
  return h
}

it('creates a scene with name and size', async () => {
  const h = setup()
  await userEvent.type(screen.getByLabelText('Nome da cena'), 'Porto')
  await userEvent.selectOptions(screen.getByLabelText('Tamanho da cena'), '48')
  await userEvent.click(screen.getByRole('button', { name: 'Criar cena' }))
  expect(h.onCreate).toHaveBeenCalledWith('Porto', 48)
})

it('marks the active scene and activates, selects, renames and deletes others', async () => {
  const h = setup()
  expect(screen.getAllByText('Ativa')).toHaveLength(1)
  expect(screen.queryByRole('button', { name: 'Ativar Floresta' })).toBeNull()
  await userEvent.click(screen.getByRole('button', { name: 'Ativar Caverna' }))
  expect(h.onActivate).toHaveBeenCalledWith('b')
  await userEvent.click(screen.getByRole('button', { name: 'Caverna' }))
  expect(h.onSelect).toHaveBeenCalledWith('b')
  vi.spyOn(window, 'prompt').mockReturnValue('Gruta')
  await userEvent.click(screen.getByRole('button', { name: 'Renomear Caverna' }))
  expect(h.onRename).toHaveBeenCalledWith('b', 'Gruta')
  vi.spyOn(window, 'confirm').mockReturnValue(true)
  await userEvent.click(screen.getByRole('button', { name: 'Apagar Caverna' }))
  await waitFor(() => expect(h.onDelete).toHaveBeenCalledWith('b'))
})

const state: MapState = { grid: createGrid(16, 16), objects: [] }

it('applies remote terrain patches only for the scene on screen', () => {
  const ev = { sceneId: 'a', patch: { cells: [{ x: 1, z: 2, terrain: 'stone' as const, height: 5 }] } }
  const next = applyRemotePaint(state, 'a', ev)
  expect(next.grid.cells[2 * 16 + 1]).toEqual({ terrain: 'stone', height: 5 })
  expect(applyRemotePaint(state, 'other', ev)).toBe(state)
})

it('applies remote object patches, defaulting props', () => {
  const wall = { id: 'w', prefab: 'wall', x: 1, z: 1, rotation: 0 as const }
  const added = applyRemoteObjects(state, 'a', { sceneId: 'a', patch: { add: [wall as never], remove: [], update: [] } })
  expect(added.objects).toEqual([{ ...wall, props: {} }])
  const removed = applyRemoteObjects(added, 'a', { sceneId: 'a', patch: { add: [], remove: ['w'], update: [] } })
  expect(removed.objects).toEqual([])
  expect(applyRemoteObjects(state, 'x', { sceneId: 'a', patch: { add: [wall as never], remove: [], update: [] } })).toBe(state)
})

it('keeps the serialized grid format used by the server', () => {
  expect(serializeGrid(createGrid(16, 16)).data).toHaveLength(16 * 16 * 2)
})
