import type { Symbol } from './faces'

export const SYMBOL_COLORS: Record<Symbol, string> = {
  success: '#22c55e',
  adaptation: '#38bdf8',
  pressure: '#ef4444',
}

type Pt = [number, number]

export function layoutSymbols(count: number): Pt[] {
  const layouts: Pt[][] = [
    [],
    [[0.5, 0.5]],
    [[0.28, 0.5], [0.72, 0.5]],
    [[0.5, 0.28], [0.27, 0.72], [0.73, 0.72]],
    [[0.28, 0.28], [0.72, 0.28], [0.28, 0.72], [0.72, 0.72]],
  ]
  return layouts[count] ?? []
}

function poly(ctx: CanvasRenderingContext2D, pts: Pt[]) {
  ctx.beginPath()
  pts.forEach(([x, y], i) => (i ? ctx.lineTo(x, y) : ctx.moveTo(x, y)))
  ctx.closePath()
}

// Success: filled circle with a check mark.
function drawSuccess(ctx: CanvasRenderingContext2D, r: number) {
  ctx.beginPath()
  ctx.arc(0, 0, r, 0, Math.PI * 2)
  ctx.fill()
  ctx.beginPath()
  ctx.moveTo(-r * 0.5, 0)
  ctx.lineTo(-r * 0.12, r * 0.4)
  ctx.lineTo(r * 0.52, -r * 0.35)
  ctx.stroke()
}

// Adaptation: hexagon with a two-way horizontal arrow.
function drawAdaptation(ctx: CanvasRenderingContext2D, r: number) {
  const hex: Pt[] = Array.from({ length: 6 }, (_, i) => [r * Math.cos((i * Math.PI) / 3), r * Math.sin((i * Math.PI) / 3)])
  poly(ctx, hex)
  ctx.fill()
  ctx.beginPath()
  ctx.moveTo(-r * 0.55, 0)
  ctx.lineTo(r * 0.55, 0)
  ctx.moveTo(-r * 0.55, 0)
  ctx.lineTo(-r * 0.25, -r * 0.28)
  ctx.moveTo(-r * 0.55, 0)
  ctx.lineTo(-r * 0.25, r * 0.28)
  ctx.moveTo(r * 0.55, 0)
  ctx.lineTo(r * 0.25, -r * 0.28)
  ctx.moveTo(r * 0.55, 0)
  ctx.lineTo(r * 0.25, r * 0.28)
  ctx.stroke()
}

// Pressure: spiky four-point star.
function drawPressure(ctx: CanvasRenderingContext2D, r: number) {
  const pts: Pt[] = Array.from({ length: 8 }, (_, i) => {
    const rad = i % 2 === 0 ? r * 1.1 : r * 0.38
    const a = (i * Math.PI) / 4 - Math.PI / 2
    return [rad * Math.cos(a), rad * Math.sin(a)]
  })
  poly(ctx, pts)
  ctx.fill()
}

const DRAW = { success: drawSuccess, adaptation: drawAdaptation, pressure: drawPressure }

export function drawSymbols(canvas: HTMLCanvasElement, symbols: Symbol[]) {
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const size = canvas.width
  ctx.clearRect(0, 0, size, size)
  const r = size * (symbols.length > 1 ? 0.19 : 0.3)
  symbols.forEach((symbol, i) => {
    const [px, py] = layoutSymbols(symbols.length)[i]
    ctx.save()
    ctx.translate(px * size, py * size)
    ctx.fillStyle = SYMBOL_COLORS[symbol]
    ctx.strokeStyle = '#0a0a0a'
    ctx.lineWidth = r * 0.16
    ctx.lineCap = 'round'
    ctx.lineJoin = 'round'
    DRAW[symbol](ctx, r)
    ctx.restore()
  })
}
