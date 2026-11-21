import type { MapGrid } from './grid'
import type { MapObject } from './objects/model'
import { TerrainScene } from './TerrainScene'

export function TerrainView({ grid, objects }: { grid: MapGrid; objects?: readonly MapObject[] }) {
  return (
    <div className="h-full w-full">
      <TerrainScene grid={grid} objects={objects ?? []} />
    </div>
  )
}
