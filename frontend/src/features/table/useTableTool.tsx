import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import { DEFAULT_TOOL_STATE, isObjectTool, TERRAIN_TOOLS, type ToolState } from '../map3d/tools'
import { toolsFor, type TableToolId } from './tools'

export interface TableToolApi {
  tool: TableToolId
  setTool: (t: TableToolId) => void
  /** Terrain editor settings (brush, terrain type, prefab…) shared with the options panel. */
  editor: ToolState
  setEditor: (s: ToolState) => void
  showGrid: boolean
  toggleGrid: () => void
  /** Draft/snapshots panel opened from the rail (GM). */
  draftOpen: boolean
  toggleDraft: () => void
  /** Whether the current scene is in draft mode (published by the map). */
  draftActive: boolean
  setDraftActive: (v: boolean) => void
  /** Character armed in the Tokens tool: the next map click puts its token there. */
  placing: string | null
  setPlacing: (id: string | null) => void
  /** Object picked with the Objects > Select mode (its panel lives in the options panel). */
  selectedObjectId: string | null
  setSelectedObjectId: (id: string | null) => void
  /** DOM node of the floating options panel; map tools portal their controls into it. */
  optionsHost: HTMLElement | null
  setOptionsHost: (el: HTMLElement | null) => void
}

const Ctx = createContext<TableToolApi | null>(null)

export function useTableTool(): TableToolApi {
  const api = useContext(Ctx)
  if (!api) throw new Error('useTableTool must be used inside <TableToolProvider>')
  return api
}

// Keeps the editor's own tool coherent with the rail tool (terrain brushes vs. object tools).
function editorToolFor(rail: TableToolId, current: ToolState['tool']): ToolState['tool'] {
  if (rail === 'wall') return 'wall'
  if (rail === 'objects') return isObjectTool(current) && current !== 'wall' ? current : 'object'
  return TERRAIN_TOOLS.some((t) => t.id === current) ? current : 'paint'
}

/** Single source of truth for the active table tool; the map reads it to decide what a click does. */
export function TableToolProvider({ isGm, children }: { isGm: boolean; children: ReactNode }) {
  const [tool, setToolState] = useState<TableToolId>('move')
  const [editor, setEditor] = useState<ToolState>(DEFAULT_TOOL_STATE)
  const [showGrid, setShowGrid] = useState(true)
  const [draftOpen, setDraftOpen] = useState(false)
  const [draftActive, setDraftActive] = useState(false)
  const [optionsHost, setOptionsHost] = useState<HTMLElement | null>(null)
  const [placing, setPlacing] = useState<string | null>(null)
  const [selectedObjectId, setSelectedObjectId] = useState<string | null>(null)

  const setTool = useCallback((t: TableToolId) => {
    if (!toolsFor(isGm).some((d) => d.id === t)) return
    setToolState(t)
    setPlacing(null)
    if (t === 'terrain' || t === 'objects' || t === 'wall') setEditor((e) => ({ ...e, tool: editorToolFor(t, e.tool) }))
  }, [isGm])
  const toggleGrid = useCallback(() => setShowGrid((v) => !v), [])
  const toggleDraft = useCallback(() => setDraftOpen((v) => !v), [])

  const value = useMemo<TableToolApi>(() => ({
    tool, setTool, editor, setEditor, showGrid, toggleGrid, draftOpen, toggleDraft, draftActive, setDraftActive,
    placing, setPlacing, selectedObjectId, setSelectedObjectId, optionsHost, setOptionsHost,
  }), [tool, setTool, editor, showGrid, toggleGrid, draftOpen, toggleDraft, draftActive, placing, selectedObjectId, optionsHost])
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}
