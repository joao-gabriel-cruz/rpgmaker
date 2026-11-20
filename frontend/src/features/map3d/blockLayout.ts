import { cellIndex, type MapGrid } from './grid'
import { TERRAIN_IDS, type TerrainTypeId } from './terrainTypes'

export type BlockPositions = Record<TerrainTypeId, number[]>

function minNeighborHeight(g: MapGrid, x: number, z: number): number {
  let min = Infinity
  for (const [dx, dz] of [[1, 0], [-1, 0], [0, 1], [0, -1]]) {
    const nx = x + dx
    const nz = z + dz
    const inside = nx >= 0 && nz >= 0 && nx < g.width && nz < g.depth
    min = Math.min(min, inside ? g.cells[cellIndex(g, nx, nz)].height : 0)
  }
  return min
}

// Flat [x,y,z,...] block centers per terrain; fully buried blocks are skipped.
export function layoutBlocks(g: MapGrid): BlockPositions {
  const out = Object.fromEntries(TERRAIN_IDS.map((id) => [id, [] as number[]])) as BlockPositions
  for (let z = 0; z < g.depth; z++) {
    for (let x = 0; x < g.width; x++) {
      const { terrain, height } = g.cells[cellIndex(g, x, z)]
      const from = Math.min(minNeighborHeight(g, x, z), height - 1)
      for (let y = Math.max(0, from); y < height; y++) out[terrain].push(x + 0.5, y + 0.5, z + 0.5)
    }
  }
  return out
}
