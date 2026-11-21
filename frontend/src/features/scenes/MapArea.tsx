import { useCallback, useEffect, useMemo, useRef, type ReactNode } from 'react'
import { useSceneTokensBridge, useTableCharacter } from '../hotbar/TableCharacterContext'
import { MapOptions } from '../table/options/MapOptions'
import { OptionsPortal } from '../table/OptionsPortal'
import { drawModeOf, isEditorTool } from '../table/tools'
import { useTableTool } from '../table/useTableTool'
import { CameraBridge } from '../table/camera/CameraBridge'
import { CameraControls } from '../table/camera/CameraControls'
import type { CameraHandle } from '../table/camera/cameraActions'
import { TerrainEditor } from '../map3d/TerrainEditor'
import { useEditHistory } from '../map3d/useEditHistory'
import { useHistoryKeys } from '../map3d/useHistoryKeys'
import { TerrainView } from '../map3d/TerrainView'
import { useMe } from '../auth/auth'
import { useLiveScene } from './useLiveScene'
import type { SceneControls } from './useSceneControls'
import { DrawingLayer } from '../drawings/DrawingLayer'
import { useDrawings } from '../drawings/useDrawings'
import { LootLayer } from '../items/LootLayer'
import { useSceneLoot, type SceneLoot } from '../items/loot'
import { TokenLayer } from '../tokens/TokenLayer'
import { TokensError } from '../tokens/TokensPanel'
import { useSceneTokens } from '../tokens/useSceneTokens'
import { useShownTokens, useVision } from '../vision/useVision'
import { VisionLayers } from '../vision/VisionLayers'
import type { MapState } from './scenes'

type Live = ReturnType<typeof useLiveScene>

interface SceneProps {
  campaignId: string
  sceneId: string
  isGm: boolean
}

function SceneCanvas({ campaignId, sceneId, isGm }: SceneProps) {
  const { state, error, paint, editObjects, command, replacedAt } = useLiveScene(campaignId, sceneId)
  const box = useRef<HTMLDivElement>(null)
  useEffect(() => {
    // Short fade/scale-in so a map.replaced does not look like a glitch.
    if (replacedAt) box.current?.animate?.([{ opacity: 0.2, transform: 'scale(0.97)' }, { opacity: 1, transform: 'scale(1)' }], { duration: 400, easing: 'ease-out' })
  }, [replacedAt])
  return (
    <div className="relative h-full w-full">
      {error && <p role="alert" className="absolute left-1/2 top-3 z-20 -translate-x-1/2 rounded-lg bg-red-950/90 px-3 py-1 text-sm text-red-300 shadow-lg">{error}</p>}
      {isGm && state?.draft && (
        <p role="status" className="absolute left-1/2 top-3 z-10 -translate-x-1/2 rounded-lg bg-amber-900/90 px-4 py-1.5 text-sm font-medium text-amber-100 shadow-lg ring-1 ring-amber-400/40">
          Editando rascunho — jogadores não veem
        </p>
      )}
      <div ref={box} className="h-full w-full">
        {state && <SceneWithTokens {...{ campaignId, sceneId, isGm, state, paint, editObjects, command, draft: !!state.draft }} />}
        {!state && !error && <Empty>Carregando mapa…</Empty>}
      </div>
    </div>
  )
}

interface WithTokens extends SceneProps, Pick<Live, 'paint' | 'editObjects' | 'command'> {
  state: MapState
  draft: boolean
}

// Everything the active rail tool needs: drawings, fog and tokens, with the tool deciding their mode.
function useMapTools({ campaignId, sceneId, isGm, state }: WithTokens, userId: string) {
  const { tool } = useTableTool()
  const vision = useVision(campaignId, sceneId, state.grid, isGm)
  const tokens = useShownTokens(useSceneTokens(campaignId, sceneId, state.grid), vision, state.grid, userId, isGm)
  const drawing = useDrawings(campaignId, sceneId, userId)
  const draw = useMemo(() => ({ ...drawing, mode: drawModeOf(tool) }), [drawing, tool])
  const fog = useMemo(() => ({ ...vision, brush: tool === 'fog' ? vision.brush : null }), [vision, tool])
  return { vision, fog, tokens, draw }
}

type Tools = ReturnType<typeof useMapTools>

interface LayersProps extends Pick<Tools, 'tokens' | 'draw' | 'fog'> {
  grid: MapState['grid']
  objects: MapState['objects']
  loot: SceneLoot
  userId: string
  isGm: boolean
  camera: CameraHandle
  onSelect: (characterId: string) => void
}

// The in-canvas layers; the rail tool decides which of them react to the pointer.
function MapLayers({ grid, objects, loot, tokens, userId, isGm, draw, fog, camera, onSelect }: LayersProps) {
  const { tool, placing, setPlacing } = useTableTool()
  const place = useCallback((c: { x: number; z: number }) => {
    if (placing) tokens.place(placing, c)
    setPlacing(null)
  }, [placing, tokens, setPlacing])
  return (
    <>
      <TokenLayer grid={grid} tokens={tokens} userId={userId} isGm={isGm} pingTool={tool === 'ping'}
        dragEnabled={tool === 'move' || tool === 'tokens'} onPlaceAt={tool === 'tokens' && placing ? place : null} onSelect={onSelect} />
      <LootLayer grid={grid} objects={objects} tokens={tokens.tokens} loot={loot} userId={userId} isGm={isGm} />
      <CameraBridge handleRef={camera} />
      <DrawingLayer grid={grid} draw={draw} />
      <VisionLayers grid={grid} vision={fog} isGm={isGm} />
    </>
  )
}

// The Mestre's edits go through the undo history; players never edit.
function useEditing({ state, paint, editObjects, isGm }: WithTokens) {
  const history = useEditHistory({ grid: state.grid, objects: state.objects, paint, editObjects })
  useHistoryKeys(history, isGm)
  return history
}

function usePublishDraft(draft: boolean, setDraftActive: (v: boolean) => void) {
  useEffect(() => {
    setDraftActive(draft)
    return () => setDraftActive(false)
  }, [draft, setDraftActive])
}

// The hotbar reads the scene's tokens (to offer "place my token") from the table context.
function usePublishTokens(tokens: Tools['tokens']) {
  const { setTokens } = useSceneTokensBridge()
  useEffect(() => {
    setTokens(tokens)
    return () => setTokens(null)
  }, [tokens, setTokens])
}

function SceneWithTokens(p: WithTokens) {
  const { campaignId, sceneId, isGm, state, command, draft } = p
  const { tool, editor, setEditor, showGrid, setDraftActive, selectedObjectId, setSelectedObjectId } = useTableTool()
  const userId = useMe().data?.id ?? ''
  const { vision, fog, tokens, draw } = useMapTools(p, userId)
  const history = useEditing(p)
  const camera = useRef<CameraHandle['current']>(null) as CameraHandle
  usePublishDraft(draft, setDraftActive)
  usePublishTokens(tokens)
  const { select } = useTableCharacter()
  const loot = useSceneLoot(campaignId, sceneId)
  const layer = <MapLayers {...{ grid: state.grid, objects: state.objects, loot, tokens, userId, isGm, draw, fog, camera, onSelect: select }} />
  const mine = tokens.tokens.find((t) => t.ownerId === userId)
  const opts = { campaignId, sceneId, isGm, grid: state.grid, objects: state.objects, onObjectPatch: history.editObjects,
    history, draw, vision, tokens, draft, onCommand: command }
  return (
    <>
      <div className="absolute left-1/2 top-3 z-10 -translate-x-1/2"><TokensError tokens={tokens} /></div>
      <OptionsPortal>
        <MapOptions {...opts} />
      </OptionsPortal>
      <CameraControls handle={camera} onCenterMap={() => tokens.focus(state.grid.width / 2, state.grid.depth / 2)}
        onCenterToken={mine ? () => tokens.focus(mine.x + 0.5, mine.z + 0.5) : undefined} />
      {isGm
        ? <TerrainEditor grid={state.grid} onPatch={history.paint} objects={state.objects} onObjectPatch={history.editObjects}
            control={{ state: editor, onChange: setEditor, showGrid, selection: [selectedObjectId, setSelectedObjectId] }}
            enabled={isEditorTool(tool)}>{layer}</TerrainEditor>
        : <TerrainView grid={state.grid} objects={state.objects}>{layer}</TerrainView>}
    </>
  )
}

function Empty({ children }: { children: ReactNode }) {
  return <p className="flex h-full items-center justify-center p-4 text-zinc-400">{children}</p>
}

/** The map: fills its parent; shows the scene picked in the top bar. Must render inside CampaignSocketProvider. */
export function MapArea({ campaignId, isGm, scenes }: { campaignId: string; isGm: boolean; scenes: SceneControls }) {
  if (scenes.loading) return <Empty>Carregando…</Empty>
  if (!scenes.selectedId) {
    return <Empty>{isGm ? 'Nenhuma cena ainda. Crie a primeira no seletor de cenas.' : 'O Mestre ainda não ativou nenhuma cena.'}</Empty>
  }
  return <SceneCanvas key={scenes.selectedId} campaignId={campaignId} sceneId={scenes.selectedId} isGm={isGm} />
}
