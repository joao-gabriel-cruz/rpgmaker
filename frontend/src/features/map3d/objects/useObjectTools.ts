import { useState } from 'react'
import type { Coord, MapGrid } from '../grid'
import type { ToolState } from '../tools'
import { addObject, isEmptyPatch, removeObject, type MapObject, type ObjectPatch } from './model'
import { drawWall } from './walls'

interface Wall {
  from: Coord
  to: Coord
}

/** Selected object id kept outside (the table shows its panel next to the rail). */
export type Selection = [string | null, (id: string | null) => void]

export type ObjectPatchHandler = (patch: ObjectPatch) => void

// Click/drag behaviour of the object tools (place, wall, select, remove).
export function useObjectTools(
  grid: MapGrid,
  objects: readonly MapObject[],
  state: ToolState,
  emit: ObjectPatchHandler,
  selection?: Selection,
) {
  const own = useState<string | null>(null)
  const [selectedId, setSelectedId] = selection ?? own
  const [wall, setWall] = useState<Wall | null>(null)
  const send = (patch: ObjectPatch) => !isEmptyPatch(patch) && emit(patch)

  const wallPreview = wall
    ? drawWall(objects, grid, wall.from, wall.to).patch.add.map((w) => ({ ...w, id: `preview-${w.x}-${w.z}` }))
    : []
  const selected = objects.find((o) => o.id === selectedId) ?? null

  return {
    selected,
    wallPreview,
    onCellDown: (at: Coord) => {
      if (state.tool === 'object') {
        send(addObject(objects, grid, {
            prefab: state.prefab,
            x: at.x,
            z: at.z,
            rotation: state.rotation,
            props: state.prefab === 'tree' ? { variant: state.variant } : undefined,
          }).patch)
      }
      if (state.tool === 'wall') setWall({ from: at, to: at })
      if (state.tool === 'select') setSelectedId(null)
    },
    onCellMove: (at: Coord) => wall && setWall({ ...wall, to: at }),
    onUp: () => {
      if (!wall) return
      send(drawWall(objects, grid, wall.from, wall.to).patch)
      setWall(null)
    },
    onPick: (id: string) => {
      if (state.tool === 'select') setSelectedId(id)
      if (state.tool === 'removeObject') send(removeObject(objects, id).patch)
    },
    clearSelection: () => setSelectedId(null),
  }
}
