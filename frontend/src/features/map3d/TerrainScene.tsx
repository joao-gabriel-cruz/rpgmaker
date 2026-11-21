import { Canvas, type ThreeEvent } from '@react-three/fiber'
import { OrbitControls } from '@react-three/drei'
import type { ReactNode } from 'react'
import { MOUSE } from 'three'
import type { Coord, MapGrid } from './grid'
import { GridLines } from './GridLines'
import { ObjectsLayer } from './objects/ObjectsLayer'
import type { MapObject } from './objects/model'
import { TerrainBlocks } from './TerrainBlocks'

export interface PointerHandlers {
  onCellDown?: (c: Coord) => void
  onCellMove?: (c: Coord) => void
  onUp?: () => void
}

interface Props extends PointerHandlers {
  grid: MapGrid
  showGrid?: boolean
  editable?: boolean
  objects?: readonly MapObject[]
  onObjectPick?: (id: string) => void
  children?: ReactNode
}

// Nudges the hit point inward along the face normal so side hits map to the block hit.
function toCell(e: ThreeEvent<PointerEvent>): Coord {
  const n = e.object.type === 'InstancedMesh' ? e.face?.normal : undefined
  return {
    x: Math.floor(e.point.x - (n?.x ?? 0) * 0.01),
    z: Math.floor(e.point.z - (n?.z ?? 0) * 0.01),
  }
}

function cellHandlers(onDown?: (c: Coord) => void, onMove?: (c: Coord) => void) {
  return {
    down: (e: ThreeEvent<PointerEvent>) => {
      if (e.button !== 0) return
      e.stopPropagation()
      onDown?.(toCell(e))
    },
    move: (e: ThreeEvent<PointerEvent>) => {
      e.stopPropagation()
      onMove?.(toCell(e))
    },
  }
}

function Ground({ grid }: { grid: MapGrid }) {
  return (
    <mesh rotation-x={-Math.PI / 2} position={[grid.width / 2, -0.01, grid.depth / 2]}>
      <planeGeometry args={[grid.width, grid.depth]} />
      <meshStandardMaterial color="#0f0f12" />
    </mesh>
  )
}

function Controls({ grid, editable }: { grid: MapGrid; editable?: boolean }) {
  const buttons = editable
    ? { RIGHT: MOUSE.ROTATE, MIDDLE: MOUSE.DOLLY }
    : { LEFT: MOUSE.ROTATE, MIDDLE: MOUSE.DOLLY, RIGHT: MOUSE.PAN }
  return (
    <OrbitControls
      target={[grid.width / 2, 1, grid.depth / 2]}
      maxPolarAngle={Math.PI * 0.49}
      mouseButtons={buttons}
    />
  )
}

export function TerrainScene({ grid, showGrid, editable, objects, onObjectPick, children, onCellDown, onCellMove, onUp }: Props) {
  const cx = grid.width / 2
  const cz = grid.depth / 2
  const span = Math.max(grid.width, grid.depth)
  const { down, move } = cellHandlers(onCellDown, onCellMove)
  return (
    <Canvas
      camera={{ position: [cx + span * 0.8, span * 0.8, cz + span * 0.8], fov: 50, far: 1000 }}
      onContextMenu={(e) => e.preventDefault()}
      onPointerUp={onUp}
      onPointerLeave={onUp}
    >
      <color attach="background" args={['#18181b']} />
      <ambientLight intensity={0.8} />
      <directionalLight position={[cx + 30, 60, cz + 20]} intensity={1.6} />
      <group onPointerDown={editable ? down : undefined} onPointerMove={editable ? move : undefined}>
        <TerrainBlocks grid={grid} />
        <Ground grid={grid} />
      </group>
      {objects && <ObjectsLayer grid={grid} objects={objects} onPick={onObjectPick} />}
      {showGrid && <GridLines grid={grid} />}
      {children}
      <Controls grid={grid} editable={editable} />
    </Canvas>
  )
}
