import type { MapGrid } from './grid'
import { TerrainScene } from './TerrainScene'

export function TerrainView({ grid }: { grid: MapGrid }) {
  return (
    <div className="h-full w-full">
      <TerrainScene grid={grid} />
    </div>
  )
}
