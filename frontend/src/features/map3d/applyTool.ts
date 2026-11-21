import type { Coord, MapGrid, MapPatch } from './grid'
import { brushCells, erase, fill, lower, paintTerrain, raise, setHeight } from './operations'
import type { ToolState } from './tools'

// Runs the active tool at a cell and returns the resulting patch.
export function applyTool(grid: MapGrid, at: Coord, s: ToolState): MapPatch {
  const cells = brushCells(grid, at, s.tool === 'fill' ? 1 : s.brushSize)
  switch (s.tool) {
    case 'paint':
      return paintTerrain(grid, cells, s.terrain).patch
    case 'raise':
      return raise(grid, cells).patch
    case 'lower':
      return lower(grid, cells).patch
    case 'height':
      return setHeight(grid, cells, s.height).patch
    case 'fill':
      return fill(grid, at, s.terrain).patch
    case 'erase':
      return erase(grid, cells).patch
    default:
      return { cells: [] }
  }
}
