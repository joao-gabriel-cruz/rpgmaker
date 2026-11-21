import { cellIndex, surfaceHeight, type Coord, type MapGrid } from './grid'

interface Props {
  grid: MapGrid
  at: Coord
  size: number
}

export function HoverCursor({ grid, at, size }: Props) {
  const cell = grid.cells[cellIndex(grid, at.x, at.z)]
  if (!cell) return null
  return (
    <mesh position={[at.x + 0.5, surfaceHeight(cell.height) + 0.05, at.z + 0.5]} raycast={() => null}>
      <boxGeometry args={[size, 0.1, size]} />
      <meshBasicMaterial color="#34d399" transparent opacity={0.45} />
    </mesh>
  )
}
