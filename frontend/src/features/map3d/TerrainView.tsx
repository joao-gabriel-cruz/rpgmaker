import type { ReactNode } from 'react'
import type { MapGrid } from './grid'
import type { MapObject } from './objects/model'
import { TerrainScene } from './TerrainScene'

export function TerrainView({ grid, objects, children }: { grid: MapGrid; objects?: readonly MapObject[]; children?: ReactNode }) {
  return (
    <div className="h-full w-full">
      <TerrainScene grid={grid} objects={objects ?? []}>{children}</TerrainScene>
    </div>
  )
}
