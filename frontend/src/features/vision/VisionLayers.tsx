import type { MapGrid } from '../map3d/grid'
import { FogBrush } from './FogBrush'
import { FogLayer } from './FogLayer'
import type { Vision } from './useVision'

/** In-canvas part of the fog feature: the cover and (for the Mestre) the brush. */
export function VisionLayers({ grid, vision, isGm }: { grid: MapGrid; vision: Vision; isGm: boolean }) {
  return (
    <>
      <FogLayer grid={grid} state={vision.state} isGm={isGm} />
      {isGm && <FogBrush grid={grid} brush={vision.brush} onApply={vision.paint} />}
    </>
  )
}
