import { useEffect, useMemo, useRef, type ReactNode } from 'react'
import { MapOptions } from '../table/options/MapOptions'
import { OptionsPortal } from '../table/OptionsPortal'
import { SURFACE } from '../table/surface'
import { drawModeOf, isEditorTool } from '../table/tools'
import { useTableTool } from '../table/useTableTool'
import { TerrainEditor } from '../map3d/TerrainEditor'
import { TerrainView } from '../map3d/TerrainView'
import { useMe } from '../auth/auth'
import { useLiveScene } from './useLiveScene'
import type { SceneControls } from './useSceneControls'
import { DrawingLayer } from '../drawings/DrawingLayer'
import { useDrawings } from '../drawings/useDrawings'
import { TokenLayer } from '../tokens/TokenLayer'
import { TokensError, TokensPanel } from '../tokens/TokensPanel'
import { useSceneTokens, type SceneTokens } from '../tokens/useSceneTokens'
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
  userId: string
  isGm: boolean
}

// The in-canvas layers; the rail tool decides which of them react to the pointer.
function MapLayers({ grid, tokens, userId, isGm, draw, fog }: LayersProps) {
  const { tool } = useTableTool()
  return (
    <>
      <TokenLayer grid={grid} tokens={tokens} userId={userId} isGm={isGm} pingTool={tool === 'ping'}
        dragEnabled={tool === 'move' || tool === 'tokens'} />
      <DrawingLayer grid={grid} draw={draw} />
      <VisionLayers grid={grid} vision={fog} isGm={isGm} />
    </>
  )
}

function SceneWithTokens(p: WithTokens) {
  const { campaignId, sceneId, isGm, state, paint, editObjects, command, draft } = p
  const { tool, editor, setEditor, showGrid, setDraftActive } = useTableTool()
  const userId = useMe().data?.id ?? ''
  const { vision, fog, tokens, draw } = useMapTools(p, userId)
  useEffect(() => {
    setDraftActive(draft)
    return () => setDraftActive(false)
  }, [draft, setDraftActive])
  const layer = <MapLayers {...{ grid: state.grid, tokens, userId, isGm, draw, fog }} />
  return (
    <>
      <div className="absolute left-1/2 top-3 z-10 -translate-x-1/2"><TokensError tokens={tokens} /></div>
      {!isGm && <PlayerTokenPrompt campaignId={campaignId} tokens={tokens} userId={userId} />}
      <OptionsPortal>
        <MapOptions {...{ campaignId, sceneId, isGm, userId, draw, vision, tokens, draft, onCommand: command }} />
      </OptionsPortal>
      {isGm
        ? <TerrainEditor grid={state.grid} onPatch={paint} objects={state.objects} onObjectPatch={editObjects}
            control={{ state: editor, onChange: setEditor, showGrid }} enabled={isEditorTool(tool)}>{layer}</TerrainEditor>
        : <TerrainView grid={state.grid} objects={state.objects}>{layer}</TerrainView>}
    </>
  )
}

// Players have no Tokens tool: the "place my token" button floats at the bottom-left of the map.
function PlayerTokenPrompt(p: { campaignId: string; tokens: SceneTokens; userId: string }) {
  return <div className={`absolute bottom-3 left-3 z-10 empty:hidden rounded-xl p-1 ${SURFACE}`}><TokensPanel {...p} isGm={false} /></div>
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
