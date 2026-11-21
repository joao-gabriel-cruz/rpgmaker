import { useState } from 'react'
import { HoverCursor } from './HoverCursor'
import type { Coord, MapGrid, MapPatch } from './grid'
import type { MapObject, ObjectPatch } from './objects/model'
import { ObjectGhost, SelectionBox } from './objects/ObjectGhost'
import { ObjectPanel } from './objects/ObjectPanel'
import { useObjectTools } from './objects/useObjectTools'
import { TerrainScene } from './TerrainScene'
import { Toolbar } from './Toolbar'
import { usePainting } from './usePainting'
import { useShortcuts } from './useShortcuts'
import { DEFAULT_TOOL_STATE, isObjectTool, type ToolState } from './tools'

interface Props {
  grid: MapGrid
  onPatch: (patch: MapPatch) => void
  objects?: readonly MapObject[]
  onObjectPatch?: (patch: ObjectPatch) => void
}

const NO_OBJECTS: readonly MapObject[] = []
const noop = () => {}

type Obj = ReturnType<typeof useObjectTools>

function ObjectOverlay({ grid, state, hover, obj }: { grid: MapGrid; state: ToolState; hover: Coord | null; obj: Obj }) {
  const ghost: MapObject | null =
    hover && state.tool === 'object'
      ? { id: 'ghost', prefab: state.prefab, x: hover.x, z: hover.z, rotation: state.rotation, props: { variant: state.variant } }
      : null
  return (
    <>
      {ghost && <ObjectGhost grid={grid} obj={ghost} />}
      {obj.wallPreview.map((w) => <ObjectGhost key={w.id} grid={grid} obj={w} />)}
      {obj.selected && state.tool === 'select' && <SelectionBox grid={grid} obj={obj.selected} />}
    </>
  )
}

type Paint = Omit<ReturnType<typeof usePainting>, 'hover'>

// Routes pointer events to the object tools or the terrain brush.
function combineHandlers(objectMode: boolean, paint: Paint, obj: Obj) {
  return {
    onCellDown: objectMode ? obj.onCellDown : paint.onCellDown,
    onCellMove: (c: Coord) => {
      paint.onCellMove(c)
      obj.onCellMove(c)
    },
    onUp: () => {
      paint.onUp()
      obj.onUp()
    },
  }
}

export function TerrainEditor({ grid, onPatch, objects = NO_OBJECTS, onObjectPatch = noop }: Props) {
  const [state, setState] = useState<ToolState>(DEFAULT_TOOL_STATE)
  const [showGrid, setShowGrid] = useState(true)
  const { hover, ...paint } = usePainting(grid, state, onPatch)
  const obj = useObjectTools(grid, objects, state, onObjectPatch)
  useShortcuts(setState)
  const objectMode = isObjectTool(state.tool)
  const pickable = state.tool === 'select' || state.tool === 'removeObject'
  const size = state.tool === 'fill' ? 1 : state.brushSize

  const handlers = combineHandlers(objectMode, paint, obj)

  return (
    <div className="flex h-full w-full">
      <Toolbar state={state} onChange={setState} showGrid={showGrid} onToggleGrid={() => setShowGrid((v) => !v)} />
      <div className="relative flex-1">
        <TerrainScene grid={grid} showGrid={showGrid} editable objects={objects} onObjectPick={pickable ? obj.onPick : undefined} {...handlers}>
          {hover && (!objectMode || state.tool === 'wall') && <HoverCursor grid={grid} at={hover} size={size} />}
          <ObjectOverlay grid={grid} state={state} hover={hover} obj={obj} />
        </TerrainScene>
        {obj.selected && state.tool === 'select' && (
          <ObjectPanel obj={obj.selected} objects={objects} grid={grid} emit={onObjectPatch} onDone={obj.clearSelection} />
        )}
      </div>
    </div>
  )
}
