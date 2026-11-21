import { CanvasTexture, SRGBColorSpace } from 'three'

const cache = new Map<string, CanvasTexture>()

/** Texture with one glyph (emoji or short text) with a dark outline; cached per glyph and color. */
export function glyphTexture(text: string, color = '#ffffff', bubble = false): CanvasTexture {
  const key = `${text}|${color}|${bubble}`
  let tex = cache.get(key)
  if (tex) return tex
  const canvas = document.createElement('canvas')
  canvas.width = 256
  canvas.height = 256
  const ctx = canvas.getContext('2d')
  if (ctx) draw(ctx, text, color, bubble)
  tex = new CanvasTexture(canvas)
  tex.colorSpace = SRGBColorSpace
  cache.set(key, tex)
  return tex
}

function draw(ctx: CanvasRenderingContext2D, text: string, color: string, bubble: boolean) {
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  if (bubble) {
    ctx.fillStyle = '#fafafa'
    ctx.strokeStyle = '#18181b'
    ctx.lineWidth = 8
    ctx.beginPath()
    ctx.roundRect(24, 30, 208, 140, 40)
    ctx.moveTo(90, 168)
    ctx.lineTo(70, 226)
    ctx.lineTo(130, 168)
    ctx.fill()
    ctx.stroke()
    ctx.font = '88px sans-serif'
    ctx.fillStyle = color
    ctx.fillText(text, 128, 104)
    return
  }
  ctx.font = 'bold 150px sans-serif'
  ctx.lineWidth = 16
  ctx.lineJoin = 'round'
  ctx.strokeStyle = '#18181b'
  ctx.strokeText(text, 128, 134)
  ctx.fillStyle = color
  ctx.fillText(text, 128, 134)
}
