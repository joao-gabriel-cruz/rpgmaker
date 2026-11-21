import { AmbienceOptions } from '../../ambience/AmbienceOptions'
import type { SceneAmbience } from '../../ambience/useSceneAmbience'
import { DrawingOptions } from '../../drawings/DrawingOptions'
import type { SceneDrawings } from '../../drawings/useDrawings'
import type { MapGrid } from '../../map3d/grid'
import type { MapObject, ObjectPatch } from '../../map3d/objects/model'
import type { EditHistory } from '../../map3d/useEditHistory'
import type { SceneTokens } from '../../tokens/useSceneTokens'
import type { Vision } from '../../vision/useVision'
import { useTableTool } from '../useTableTool'
import { DraftOptions } from './DraftOptions'
import { FogOptions } from './FogOptions'
import { ObjectOptions } from './ObjectOptions'
import { TerrainOptions } from './TerrainOptions'
import { TokensOptions } from './TokensOptions'
import { useOptionKeys } from './useOptionKeys'
import { WallOptions } from './WallOptions'

export interface MapOptionsProps {
  campaignId: string
  sceneId: string
  isGm: boolean
  grid: MapGrid
  objects: readonly MapObject[]
  onObjectPatch: (patch: ObjectPatch) => void
  history: EditHistory
  draw: SceneDrawings
  vision: Vision
  tokens: SceneTokens
  draft: boolean
  ambience: SceneAmbience
  onCommand: (type: string, extra?: object) => void
}

/**
 * Content of the floating options panel for the active rail tool (portaled in by the map).
 * One small component per tool; add new per-tool options here.
 */
export function MapOptions(p: MapOptionsProps) {
  const { tool, draftOpen, editor, setEditor, showGrid, toggleGrid, selectedObjectId, setSelectedObjectId } = useTableTool()
  useOptionKeys(tool, editor, setEditor, p.isGm)
  return (
    <div className="flex flex-col gap-4">
      {tool === 'draw' && <DrawingOptions draw={p.draw} isGm={p.isGm} />}
      {p.isGm && tool === 'terrain' && (
        <TerrainOptions {...{ editor, setEditor, showGrid, toggleGrid, history: p.history }} />
      )}
      {p.isGm && tool === 'objects' && (
        <ObjectOptions {...{ editor, setEditor, grid: p.grid, objects: p.objects, onObjectPatch: p.onObjectPatch,
          history: p.history, selectedId: selectedObjectId, onSelect: setSelectedObjectId }} />
      )}
      {p.isGm && tool === 'wall' && <WallOptions history={p.history} />}
      {p.isGm && tool === 'fog' && <FogOptions vision={p.vision} />}
      {p.isGm && tool === 'ambience' && <AmbienceOptions campaignId={p.campaignId} scene={p.ambience} />}
      {p.isGm && tool === 'tokens' && <TokensOptions campaignId={p.campaignId} tokens={p.tokens} />}
      {p.isGm && draftOpen && <DraftOptions campaignId={p.campaignId} sceneId={p.sceneId} draft={p.draft} onCommand={p.onCommand} />}
    </div>
  )
}
