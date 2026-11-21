import { useMemo } from 'react'
import type { MapGrid } from '../grid'
import type { MapObject } from './model'
import { ObjectInstance } from './ObjectInstance'
import { WallInstances } from './WallInstances'

interface Props {
  grid: MapGrid
  objects: readonly MapObject[]
  /** Only set by tools that select/remove objects; otherwise clicks pass through to the terrain. */
  onPick?: (id: string) => void
}

export function ObjectsLayer({ grid, objects, onPick }: Props) {
  const walls = useMemo(() => objects.filter((o) => o.prefab === 'wall'), [objects])
  return (
    <group>
      <WallInstances walls={walls} grid={grid} onPick={onPick} />
      {objects.map((o) => o.prefab !== 'wall' && <ObjectInstance key={o.id} obj={o} grid={grid} objects={objects} onPick={onPick} />)}
    </group>
  )
}
