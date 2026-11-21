import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { createGrid, serializeGrid } from '../map3d/grid'
import { NewSceneDialog } from './NewSceneDialog'

const infos = [
  { id: 'vila-abandonada', name: 'Vila Abandonada', description: 'Casas caídas.', tags: ['urbano', 'perigo'], width: 64, depth: 64, ambience: { timeOfDay: 'dusk', weather: 'clear' } },
  { id: 'rodovia-engarrafada', name: 'Rodovia Engarrafada', description: 'Carros parados.', tags: ['combate'], width: 96, depth: 64, ambience: { timeOfDay: 'day', weather: 'fog' } },
]

let fetched: string[]
const fill = vi.fn()

beforeEach(() => {
  fetched = []
  fill.mockClear()
  vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue({ fillRect: fill, fillStyle: '' } as unknown as CanvasRenderingContext2D)
  vi.stubGlobal('fetch', vi.fn((url: string) => {
    fetched.push(url)
    const body = url.includes('/preview')
      ? { grid: serializeGrid(createGrid(4, 4)), objects: [{ id: 'a', prefab: 'tree', x: 1, z: 1, rotation: 0, props: {} }] }
      : infos
    return Promise.resolve(new Response(JSON.stringify(body), { status: 200 }))
  }))
})

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

function open() {
  const h = { onClose: vi.fn(), onCreateBlank: vi.fn().mockResolvedValue(undefined), onCreateFromTemplate: vi.fn().mockResolvedValue(undefined) }
  render(
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
      <NewSceneDialog campaignId="c1" {...h} />
    </QueryClientProvider>,
  )
  return h
}

const previewSeeds = () => fetched.filter((u) => u.includes('vila-abandonada/preview')).map((u) => u.split('seed=')[1])

it('lists the templates as cards with size, tags and a thumbnail drawn from the grid', async () => {
  open()
  const card = await screen.findByRole('button', { name: 'Modelo Vila Abandonada' })
  expect(within(card).getByText('64×64')).toBeInTheDocument()
  expect(within(card).getByText('urbano')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Modelo Rodovia Engarrafada' })).toBeInTheDocument()
  expect(fetched[0]).toBe('/api/campaigns/c1/map-templates')
  await waitFor(() => expect(fill).toHaveBeenCalledTimes(16 + 1))
  expect(within(card).getByRole('img', { name: 'Miniatura de Vila Abandonada' })).toBeInTheDocument()
})

it('filters the gallery by tag', async () => {
  open()
  await screen.findByRole('button', { name: 'Modelo Vila Abandonada' })
  await userEvent.click(screen.getByRole('button', { name: 'combate' }))
  expect(screen.queryByRole('button', { name: 'Modelo Vila Abandonada' })).toBeNull()
  expect(screen.getByRole('button', { name: 'Modelo Rodovia Engarrafada' })).toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Todos' }))
  expect(screen.getByRole('button', { name: 'Modelo Vila Abandonada' })).toBeInTheDocument()
})

it('reloads the preview with another seed on "Outra variação"', async () => {
  open()
  await userEvent.click(await screen.findByRole('button', { name: 'Modelo Vila Abandonada' }))
  await screen.findByRole('region', { name: 'Prévia de Vila Abandonada' })
  await waitFor(() => expect(previewSeeds()).toHaveLength(1))
  await userEvent.click(screen.getByRole('button', { name: 'Outra variação' }))
  await waitFor(() => expect(previewSeeds()).toHaveLength(2))
  const [a, b] = previewSeeds()
  expect(a).not.toBe(b)
})

it('creates the scene from the selected template sending templateId and the previewed seed', async () => {
  const h = open()
  await userEvent.click(await screen.findByRole('button', { name: 'Modelo Vila Abandonada' }))
  await userEvent.click(await screen.findByRole('button', { name: 'Outra variação' }))
  await waitFor(() => expect(previewSeeds()).toHaveLength(2))
  const seed = Number(previewSeeds()[1])
  const panel = screen.getByRole('region', { name: 'Prévia de Vila Abandonada' })
  const name = within(panel).getByLabelText('Nome da cena')
  await userEvent.clear(name)
  await userEvent.type(name, 'Minha vila')
  await userEvent.click(within(panel).getByRole('button', { name: 'Criar cena' }))
  expect(h.onCreateFromTemplate).toHaveBeenCalledWith('Minha vila', 'vila-abandonada', seed)
  await waitFor(() => expect(h.onClose).toHaveBeenCalled())
})

it('keeps the dialog open and shows the error when creation fails', async () => {
  const h = open()
  h.onCreateFromTemplate.mockRejectedValue(new Error('Falhou feio'))
  await userEvent.click(await screen.findByRole('button', { name: 'Modelo Vila Abandonada' }))
  await userEvent.click(await screen.findByRole('button', { name: 'Criar cena' }))
  expect(await screen.findByRole('alert')).toHaveTextContent('Falhou feio')
  expect(h.onClose).not.toHaveBeenCalled()
})

it('creates a blank scene with a size up to 128', async () => {
  const h = open()
  await userEvent.click(screen.getByRole('tab', { name: 'Em branco' }))
  await userEvent.type(screen.getByLabelText('Nome da cena'), 'Porto')
  await userEvent.selectOptions(screen.getByLabelText('Tamanho da cena'), '128')
  await userEvent.click(screen.getByRole('button', { name: 'Criar cena em branco' }))
  expect(h.onCreateBlank).toHaveBeenCalledWith('Porto', 128)
})
