import type { ReactNode } from 'react'
import type { MapGrid } from './grid'
import type { MapObject } from './objects/model'
import { TerrainScene } from './TerrainScene'

interface Props {
  grid: MapGrid
  objects?: readonly MapObject[]
  environment?: ReactNode
  children?: ReactNode
}

export function TerrainView({ grid, objects, environment, children }: Props) {
  return (
    <div className="h-full w-full">
      <TerrainScene grid={grid} objects={objects ?? []} environment={environment}>{children}</TerrainScene>
    </div>
  )
}
