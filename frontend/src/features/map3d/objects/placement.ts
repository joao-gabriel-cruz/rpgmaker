import { getCell, type MapGrid } from '../grid'
import { footprint, occupiedCells, type MapObject } from './model'

// Objects stand on the tallest terrain cell under their footprint.
export function baseHeight(grid: MapGrid, o: MapObject): number {
  return Math.max(0, ...occupiedCells(o).map((c) => getCell(grid, c.x, c.z)?.height ?? 0))
}

export function placement(grid: MapGrid, o: MapObject) {
  const { w, d } = footprint(o.prefab, o.rotation)
  return { cx: o.x + w / 2, cz: o.z + d / 2, y: baseHeight(grid, o), w, d }
}
