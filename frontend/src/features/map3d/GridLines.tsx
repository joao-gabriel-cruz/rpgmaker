import { useMemo } from 'react'
import { surfaceHeight, type MapGrid } from './grid'

// Outlines each cell on top of its terrain so the grid is visible over blocks.
export function GridLines({ grid }: { grid: MapGrid }) {
  const positions = useMemo(() => {
    const v: number[] = []
    for (let z = 0; z < grid.depth; z++) {
      for (let x = 0; x < grid.width; x++) {
        const y = surfaceHeight(grid.cells[z * grid.width + x].height) + 0.02
        v.push(x, y, z, x + 1, y, z, x + 1, y, z, x + 1, y, z + 1)
        v.push(x + 1, y, z + 1, x, y, z + 1, x, y, z + 1, x, y, z)
      }
    }
    return new Float32Array(v)
  }, [grid])
  return (
    <lineSegments key={positions.length} raycast={() => null}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" args={[positions, 3]} />
      </bufferGeometry>
      <lineBasicMaterial color="#e4e4e7" transparent opacity={0.28} />
    </lineSegments>
  )
}
