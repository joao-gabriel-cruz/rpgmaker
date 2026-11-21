import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState, type ReactNode } from 'react'
import { afterEach, expect, it, vi } from 'vitest'
import { stubApi } from '../../../test/renderApp'
import { createGrid } from '../../map3d/grid'
import type { MapObject } from '../../map3d/objects/model'
import { DEFAULT_TOOL_STATE, type ToolState } from '../../map3d/tools'
import type { EditHistory } from '../../map3d/useEditHistory'
import type { Vision } from '../../vision/useVision'
import { NO_FOG } from '../../vision/vision'
import { ShortcutsDialog } from '../ShortcutsDialog'
import { TableToolProvider, useTableTool } from '../useTableTool'
import { DraftOptions } from './DraftOptions'
import { FogOptions } from './FogOptions'
import { filterPrefabs } from './prefabCatalog'
import { MapOptions, type MapOptionsProps } from './MapOptions'
import { ObjectOptions } from './ObjectOptions'
import { subToolForKey } from './useOptionKeys'
import { TerrainOptions } from './TerrainOptions'
import { WallOptions } from './WallOptions'
import { TokensOptions } from './TokensOptions'
import type { SceneTokens } from '../../tokens/useSceneTokens'

afterEach(() => vi.unstubAllGlobals())

const history = (over: Partial<EditHistory> = {}): EditHistory => ({
  paint: vi.fn(), editObjects: vi.fn(), undo: vi.fn(), redo: vi.fn(), canUndo: true, canRedo: false, lastKind: 'objects', ...over,
})

// Keeps the editor state live so segmented controls can be observed changing it.
function Terrain({ onState, initial = DEFAULT_TOOL_STATE, hist = history() }: { onState?: (s: ToolState) => void; initial?: ToolState; hist?: EditHistory }) {
  const [state, setState] = useState(initial)
  const [grid, setGrid] = useState(true)
  onState?.(state)
  return (
    <>
      <TerrainOptions editor={state} setEditor={setState} showGrid={grid} toggleGrid={() => setGrid((g) => !g)} history={hist} />
      <output aria-label="grade">{String(grid)}</output>
    </>
  )
}

it('terrain panel: segmented sub-tools change the editor state and only relevant options show', async () => {
  let last = DEFAULT_TOOL_STATE
  render(<Terrain onState={(s) => { last = s }} />)
  expect(screen.getByRole('group', { name: 'Terreno' })).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Pintar' })).toHaveAttribute('aria-pressed', 'true')
  expect(screen.queryByRole('group', { name: 'Altura' })).toBeNull()
  expect(screen.queryByText('Paleta')).toBeNull()
  await userEvent.click(screen.getByRole('button', { name: 'Definir altura' }))
  expect(last.tool).toBe('height')
  await userEvent.click(screen.getByRole('button', { name: 'Aumentar altura' }))
  expect(last.height).toBe(3)
  expect(screen.queryByRole('group', { name: 'Terreno' })).toBeNull()
  await userEvent.click(screen.getByRole('button', { name: 'Pintar' }))
  await userEvent.click(screen.getByRole('button', { name: 'Água' }))
  expect(last.terrain).toBe('water')
  await userEvent.click(screen.getByRole('button', { name: 'Pincel 3×3' }))
  expect(last.brushSize).toBe(3)
  await userEvent.click(screen.getByRole('button', { name: 'Preencher' }))
  expect(screen.queryByRole('group', { name: 'Tamanho do pincel' })).toBeNull()
})

it('terrain panel: grid switch, undo and redo buttons, no object blocks', async () => {
  const hist = history({ canRedo: true })
  render(<Terrain hist={hist} />)
  await userEvent.click(screen.getByRole('switch', { name: 'Mostrar grade' }))
  expect(screen.getByLabelText('grade')).toHaveTextContent('false')
  await userEvent.click(screen.getByRole('button', { name: 'Desfazer' }))
  await userEvent.click(screen.getByRole('button', { name: 'Refazer' }))
  expect(hist.undo).toHaveBeenCalled()
  expect(hist.redo).toHaveBeenCalled()
  expect(screen.queryByText('Paleta')).toBeNull()
  expect(screen.queryByRole('button', { name: 'Árvore' })).toBeNull()
})

const grid = createGrid(8, 8)
const door: MapObject = { id: 'd1', prefab: 'door', x: 1, z: 1, rotation: 0, props: { open: false } }

function Objects({ selectedId = null, emit = vi.fn(), initial = { ...DEFAULT_TOOL_STATE, tool: 'object' as const } }: { selectedId?: string | null; emit?: () => void; initial?: ToolState }) {
  const [state, setState] = useState<ToolState>(initial)
  return (
    <>
      <ObjectOptions editor={state} setEditor={setState} grid={grid} objects={[door]} onObjectPatch={emit} selectedId={selectedId} onSelect={() => {}} history={history()} />
      <output aria-label="estado">{`${state.tool}|${state.prefab}|${state.rotation}`}</output>
    </>
  )
}

it('objects panel: catalog filters by search and category, picks a prefab and rotates', async () => {
  render(<Objects />)
  const catalog = within(screen.getByRole('list', { name: 'Catálogo de objetos' }))
  expect(catalog.getAllByRole('button')).toHaveLength(13)
  await userEvent.type(screen.getByRole('searchbox', { name: 'Buscar objeto' }), 'arvore')
  expect(catalog.getAllByRole('button').map((b) => b.getAttribute('aria-label'))).toEqual(['Árvore'])
  await userEvent.clear(screen.getByRole('searchbox'))
  await userEvent.click(screen.getByRole('button', { name: 'Estruturas' }))
  expect(catalog.getAllByRole('button')).toHaveLength(6)
  await userEvent.click(catalog.getByRole('button', { name: 'Porta' }))
  expect(screen.getByLabelText('estado')).toHaveTextContent('object|door|0')
  await userEvent.click(screen.getByRole('button', { name: 'Girar objeto' }))
  expect(screen.getByLabelText('estado')).toHaveTextContent('object|door|90')
})

it('objects panel: modes switch and the selected object offers rotate, door toggle and remove', async () => {
  const emit = vi.fn()
  render(<Objects selectedId="d1" emit={emit} initial={{ ...DEFAULT_TOOL_STATE, tool: 'object' }} />)
  await userEvent.click(screen.getByRole('button', { name: 'Selecionar' }))
  expect(screen.getByLabelText('estado')).toHaveTextContent('select')
  expect(screen.queryByRole('searchbox')).toBeNull()
  const sel = within(screen.getByRole('group', { name: 'Objeto selecionado' }))
  await userEvent.click(sel.getByRole('button', { name: 'Abrir porta' }))
  expect(emit.mock.calls[0][0].update[0].props.open).toBe(true)
  await userEvent.click(sel.getByRole('button', { name: /Girar 90/ }))
  expect(emit.mock.calls[1][0].update[0].rotation).toBe(90)
  await userEvent.click(sel.getByRole('button', { name: 'Remover' }))
  expect(emit.mock.calls[2][0].remove).toEqual(['d1'])
})

it('filters the catalog ignoring accents and case', () => {
  expect(filterPrefabs('ÁRVORE', 'all').map((p) => p.id)).toEqual(['tree'])
  expect(filterPrefabs('', 'objeto').map((p) => p.id)).toEqual(['car', 'crate', 'campfire'])
  expect(filterPrefabs('zzz', 'all')).toEqual([])
})

it('wall panel: shows the instruction and undoes the last stroke only when it was an object edit', async () => {
  const hist = history()
  const { rerender } = render(<WallOptions history={hist} />)
  expect(screen.getByText(/Clique e arraste/)).toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Desfazer último traço' }))
  expect(hist.undo).toHaveBeenCalled()
  rerender(<WallOptions history={history({ lastKind: 'paint' })} />)
  expect(screen.getByRole('button', { name: 'Desfazer último traço' })).toBeDisabled()
})

function vision(over: Partial<Vision> = {}): Vision {
  return {
    state: { ...NO_FOG, mode: 'manual' }, brush: null, setBrush: vi.fn(), paint: vi.fn(), configure: vi.fn(), all: vi.fn(), error: '',
    clearError: vi.fn(), ...over,
  } as Vision
}

it('fog panel: mode, brush and reveal/hide all call the vision commands', async () => {
  const v = vision({ brush: 'reveal' })
  render(<FogOptions vision={v} />)
  await userEvent.click(screen.getByRole('button', { name: 'Sem névoa' }))
  expect(v.configure).toHaveBeenCalledWith('off', 12)
  await userEvent.click(screen.getByRole('button', { name: 'Esconder área' }))
  expect(v.setBrush).toHaveBeenCalledWith('hide')
  await userEvent.click(screen.getByRole('button', { name: 'Revelar tudo' }))
  expect(v.all).toHaveBeenCalledWith(true)
  expect(screen.queryByRole('slider')).toBeNull()
})

it('fog panel: the radius slider only exists in dynamic mode and commits on release', () => {
  const v = vision({ brush: 'reveal', state: { ...NO_FOG, mode: 'dynamic', visionRadius: 8 } })
  render(<FogOptions vision={v} />)
  const slider = screen.getByRole('slider', { name: 'Raio de visão' })
  fireEvent.change(slider, { target: { value: '15' } })
  fireEvent.pointerUp(slider)
  expect(v.configure).toHaveBeenCalledWith('dynamic', 15)
})

it('fog panel: arms the reveal brush when none is active', () => {
  const v = vision()
  render(<FogOptions vision={v} />)
  expect(v.setBrush).toHaveBeenCalledWith('reveal')
})

const wrap = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>{children}</QueryClientProvider>
)

it('draft panel: inline name field saves a state without window.prompt, fire and delete work', async () => {
  const calls: string[] = []
  stubApi((m, path, body) => {
    calls.push(`${m} ${path} ${body ?? ''}`)
    return { status: 200, body: path.endsWith('/snapshots') && m === 'GET' ? [{ id: 's1', name: 'erupção', createdAt: '' }] : {} }
  })
  const prompt = vi.spyOn(window, 'prompt')
  const onCommand = vi.fn()
  render(<DraftOptions campaignId="c" sceneId="a" draft={false} onCommand={onCommand} />, { wrapper: wrap })
  await screen.findByText('erupção')
  expect(screen.getByRole('button', { name: 'Salvar estado atual' })).toBeDisabled()
  await userEvent.type(screen.getByLabelText('Nome do estado salvo'), 'Noite{Enter}')
  expect(calls.some((c) => c.startsWith('POST') && c.includes('Noite'))).toBe(true)
  expect(prompt).not.toHaveBeenCalled()
  await userEvent.click(screen.getByRole('button', { name: 'Disparar erupção' }))
  expect(onCommand).toHaveBeenCalledWith('map.snapshot.apply', { snapshotId: 's1' })
  await userEvent.click(screen.getByRole('switch', { name: 'Rascunho' }))
  expect(onCommand).toHaveBeenCalledWith('map.draft.start')
  await userEvent.click(screen.getByRole('button', { name: 'Apagar erupção' }))
  expect(calls.some((c) => c.startsWith('DELETE') && c.includes('s1'))).toBe(true)
})

it('draft panel: Aplicar and Descartar only while a draft is active', async () => {
  stubApi(() => ({ status: 200, body: [] }))
  const onCommand = vi.fn()
  const { rerender } = render(<DraftOptions campaignId="c" sceneId="a" draft={false} onCommand={onCommand} />, { wrapper: wrap })
  expect(screen.queryByRole('button', { name: 'Aplicar' })).toBeNull()
  rerender(<DraftOptions campaignId="c" sceneId="a" draft onCommand={onCommand} />)
  await userEvent.click(screen.getByRole('button', { name: 'Aplicar' }))
  await userEvent.click(screen.getByRole('button', { name: 'Descartar' }))
  expect(onCommand.mock.calls).toEqual([['map.draft.apply'], ['map.draft.discard']])
})

function Dispatcher({ tool, draftOpen = false }: { tool: Parameters<ReturnType<typeof useTableTool>['setTool']>[0]; draftOpen?: boolean }) {
  const t = useTableTool()
  if (t.tool !== tool) t.setTool(tool)
  if (draftOpen && !t.draftOpen) t.toggleDraft()
  const p = {
    campaignId: 'c', sceneId: 'a', isGm: true, grid, objects: [], onObjectPatch: vi.fn(), history: history(), draw: {}, draft: false,
    vision: vision({ brush: 'reveal' }), tokens: { tokens: [], focusAll: vi.fn(), place: vi.fn(), remove: vi.fn() }, onCommand: vi.fn(),
  } as unknown as MapOptionsProps
  return <MapOptions {...p} />
}

const mount = (tool: Parameters<typeof Dispatcher>[0]['tool']) => {
  stubApi(() => ({ status: 200, body: [] }))
  return render(<TableToolProvider isGm><Dispatcher tool={tool} /></TableToolProvider>, { wrapper: wrap })
}

it('each tool shows only its own options in the panel', () => {
  const has = (n: string | RegExp) => screen.queryByRole('group', { name: n }) ?? screen.queryByRole('button', { name: n })
  const a = mount('terrain')
  expect(has('Ferramenta de terreno')).not.toBeNull()
  expect(has('Modo dos objetos')).toBeNull()
  expect(has('Modo da névoa')).toBeNull()
  a.unmount()
  const b = mount('objects')
  expect(has('Modo dos objetos')).not.toBeNull()
  expect(has('Ferramenta de terreno')).toBeNull()
  expect(has('Mostrar grade')).toBeNull()
  b.unmount()
  const c = mount('fog')
  expect(has('Modo da névoa')).not.toBeNull()
  expect(has('Modo dos objetos')).toBeNull()
  c.unmount()
  mount('tokens')
  expect(has('Focar todos aqui')).not.toBeNull()
  expect(has('Modo da névoa')).toBeNull()
})

it('number keys pick sub-tools in Terrain and Objects only, and the dialog lists the new shortcuts', () => {
  expect(subToolForKey('terrain', '3')).toBe('lower')
  expect(subToolForKey('terrain', '6')).toBe('erase')
  expect(subToolForKey('objects', '2')).toBe('select')
  expect(subToolForKey('objects', '5')).toBeUndefined()
  expect(subToolForKey('fog', '1')).toBeUndefined()
  render(<ShortcutsDialog isGm onClose={() => {}} />)
  for (const t of [/Desfazer a última edição/, /Refazer/, /Centralizar no mapa/, /Alternar visão de cima/, /Sub-ferramenta do Terreno/]) {
    expect(screen.getByText(t)).toBeInTheDocument()
  }
})

it('tokens panel: lists characters, arms a map click, places automatically, removes and focuses everyone', async () => {
  const chars = [
    { id: 'c1', name: 'Ana', kind: 'pc', ownerId: 'u1', sheet: { identity: { portraitUrl: null } } },
    { id: 'c2', name: 'Saqueador', kind: 'npc', ownerId: null, sheet: { identity: { portraitUrl: null } } },
  ]
  stubApi(() => ({ status: 200, body: chars }))
  const tokens = { tokens: [{ id: 't2', characterId: 'c2' }], place: vi.fn(), remove: vi.fn(), focusAll: vi.fn() } as unknown as SceneTokens
  const armed = vi.fn()
  function Probe() {
    const { placing } = useTableTool()
    armed(placing)
    return null
  }
  render(<TableToolProvider isGm><Probe /><TokensOptions campaignId="c" tokens={tokens} /></TableToolProvider>, { wrapper: wrap })
  await userEvent.click(await screen.findByRole('button', { name: 'Colocar Ana na cena' }))
  expect(armed).toHaveBeenLastCalledWith('c1')
  expect(screen.getByRole('status')).toHaveTextContent('Clique no mapa para posicionar Ana')
  await userEvent.click(screen.getByRole('button', { name: 'Colocar Ana automaticamente' }))
  expect(tokens.place).toHaveBeenCalledWith('c1')
  await userEvent.click(screen.getByRole('button', { name: 'Remover token de Saqueador' }))
  expect(tokens.remove).toHaveBeenCalledWith('t2')
  await userEvent.click(screen.getByRole('button', { name: 'Focar todos aqui' }))
  expect(tokens.focusAll).toHaveBeenCalled()
})
