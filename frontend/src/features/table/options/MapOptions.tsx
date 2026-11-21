import { DrawingOptions } from '../../drawings/DrawingOptions'
import type { SceneDrawings } from '../../drawings/useDrawings'
import { Toolbar } from '../../map3d/Toolbar'
import { GmPanel } from '../../scenes/GmPanel'
import { TokensPanel } from '../../tokens/TokensPanel'
import type { SceneTokens } from '../../tokens/useSceneTokens'
import { FogPanel } from '../../vision/FogPanel'
import type { Vision } from '../../vision/useVision'
import { isEditorTool } from '../tools'
import { useTableTool } from '../useTableTool'

export interface MapOptionsProps {
  campaignId: string
  sceneId: string
  isGm: boolean
  userId: string
  draw: SceneDrawings
  vision: Vision
  tokens: SceneTokens
  draft: boolean
  onCommand: (type: string, extra?: object) => void
}

/**
 * Content of the floating options panel for the active rail tool (portaled in by the map).
 * Delivery 3 refines each branch; add new per-tool option components here.
 */
export function MapOptions(p: MapOptionsProps) {
  const { tool, draftOpen, editor, setEditor, showGrid, toggleGrid } = useTableTool()
  return (
    <div className="flex flex-col gap-4">
      {tool === 'draw' && <DrawingOptions draw={p.draw} isGm={p.isGm} />}
      {p.isGm && isEditorTool(tool) && (
        <Toolbar state={editor} onChange={setEditor} showGrid={showGrid} onToggleGrid={toggleGrid} />
      )}
      {p.isGm && tool === 'fog' && <FogPanel vision={p.vision} />}
      {p.isGm && tool === 'tokens' && <TokensPanel campaignId={p.campaignId} tokens={p.tokens} userId={p.userId} isGm />}
      {p.isGm && draftOpen && <GmPanel campaignId={p.campaignId} sceneId={p.sceneId} draft={p.draft} onCommand={p.onCommand} />}
    </div>
  )
}
