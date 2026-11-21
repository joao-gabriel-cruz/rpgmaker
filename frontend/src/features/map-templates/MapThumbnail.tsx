import { useEffect, useRef } from 'react'
import type { TemplatePreview } from './templates'
import { drawThumbnail } from './thumbnail'

interface Props {
  preview: TemplatePreview | undefined
  /** Pixels per cell of the canvas bitmap (CSS scales it to the box). */
  scale: number
  label: string
  className?: string
}

/** Top-down map thumbnail drawn on a 2D canvas from the grid and objects. */
export function MapThumbnail({ preview, scale, label, className = '' }: Props) {
  const ref = useRef<HTMLCanvasElement>(null)
  useEffect(() => {
    const ctx = ref.current?.getContext('2d')
    if (preview && ctx) drawThumbnail(ctx, preview.grid, preview.objects, scale)
  }, [preview, scale])
  const w = (preview?.grid.width ?? 64) * scale
  const h = (preview?.grid.depth ?? 64) * scale
  return (
    <canvas ref={ref} role="img" aria-label={label} width={w} height={h} style={{ imageRendering: 'pixelated' }}
      className={`block w-full rounded-lg bg-zinc-800 ${preview ? '' : 'animate-pulse'} ${className}`} />
  )
}
