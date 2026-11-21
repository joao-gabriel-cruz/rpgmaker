import { useCallback, useState, type ReactNode } from 'react'
import { HoverCursor } from './HoverCursor'
import type { Coord, MapGrid, MapPatch } from './grid'
import type { MapObject, ObjectPatch } from './objects/model'
import { ObjectGhost, SelectionBox } from './objects/ObjectGhost'
import { ObjectPanel } from './objects/ObjectPanel'
import { useObjectTools, type Selection } from './objects/useObjectTools'
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
  children?: ReactNode
  /** Table mode: tool state lives outside (rail + options panel) and the inline toolbar is hidden. */
  control?: EditorControl
  /** When false, clicks never paint or place objects (the map behaves like the player view). */
  enabled?: boolean
}

export interface EditorControl {
  state: ToolState
  onChange: (next: ToolState) => void
  showGrid: boolean
  /** Selected object, owned by the table so its panel can live in the options panel. */
  selection?: Selection
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

function useEditorState(control?: EditorControl) {
  const [own, setOwn] = useState<ToolState>(DEFAULT_TOOL_STATE)
  const [ownGrid, setOwnGrid] = useState(true)
  const state = control?.state ?? own
  const onChange = control?.onChange
  const update = useCallback((f: (s: ToolState) => ToolState) => {
    if (onChange) onChange(f(state))
    else setOwn(f)
  }, [onChange, state])
  return { state, setState: onChange ?? setOwn, update, showGrid: control?.showGrid ?? ownGrid, toggleGrid: () => setOwnGrid((v) => !v) }
}

export function TerrainEditor({ grid, onPatch, objects = NO_OBJECTS, onObjectPatch = noop, children, control, enabled = true }: Props) {
  const { state, setState, update, showGrid, toggleGrid } = useEditorState(control)
  const { hover, ...paint } = usePainting(grid, state, onPatch)
  const obj = useObjectTools(grid, objects, state, onObjectPatch, control?.selection)
  useShortcuts(update, !control ? 'standalone' : enabled ? 'external' : 'off')
  const objectMode = isObjectTool(state.tool)
  const pickable = enabled && (state.tool === 'select' || state.tool === 'removeObject')
  const size = state.tool === 'fill' ? 1 : state.brushSize
  const all = combineHandlers(objectMode, paint, obj)
  const handlers = enabled ? all : { onUp: all.onUp }
  const at = enabled ? hover : null

  return (
    <div className="flex h-full w-full">
      {!control && <Toolbar state={state} onChange={setState} showGrid={showGrid} onToggleGrid={toggleGrid} />}
      <div className="relative flex-1">
        <TerrainScene grid={grid} showGrid={showGrid} editable={enabled} objects={objects} onObjectPick={pickable ? obj.onPick : undefined} {...handlers}>
          {at && (!objectMode || state.tool === 'wall') && <HoverCursor grid={grid} at={at} size={size} />}
          {enabled && <ObjectOverlay grid={grid} state={state} hover={hover} obj={obj} />}
          {children}
        </TerrainScene>
        {!control && enabled && obj.selected && state.tool === 'select' && (
          <ObjectPanel obj={obj.selected} objects={objects} grid={grid} emit={onObjectPatch} onDone={obj.clearSelection} />
        )}
      </div>
    </div>
  )
}
