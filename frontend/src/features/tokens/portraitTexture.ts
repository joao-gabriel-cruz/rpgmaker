import { CanvasTexture, SRGBColorSpace } from 'three'
import { initials } from './tokens'

const SIZE = 256

function drawInitials(ctx: CanvasRenderingContext2D, name: string, color: string) {
  ctx.fillStyle = color
  ctx.fillRect(0, 0, SIZE, SIZE)
  ctx.fillStyle = '#ffffff'
  ctx.font = 'bold 110px sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(initials(name), SIZE / 2, SIZE / 2 + 6)
}

/** Texture with initials on the player's color; swaps in the portrait once it loads. */
export function portraitTexture(name: string, color: string, url: string | null): CanvasTexture {
  const canvas = document.createElement('canvas')
  canvas.width = canvas.height = SIZE
  const ctx = canvas.getContext('2d')
  const tex = new CanvasTexture(canvas)
  tex.colorSpace = SRGBColorSpace
  if (!ctx) return tex
  drawInitials(ctx, name, color)
  if (url) {
    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => {
      ctx.drawImage(img, 0, 0, SIZE, SIZE)
      tex.needsUpdate = true
    }
    img.src = url
  }
  return tex
}

/** Texture for the name tag floating above the token. */
export function nameTexture(name: string): CanvasTexture {
  const canvas = document.createElement('canvas')
  canvas.width = 512
  canvas.height = 96
  const ctx = canvas.getContext('2d')
  const tex = new CanvasTexture(canvas)
  tex.colorSpace = SRGBColorSpace
  if (!ctx) return tex
  ctx.fillStyle = 'rgba(9,9,11,0.75)'
  ctx.fillRect(0, 0, 512, 96)
  ctx.fillStyle = '#f4f4f5'
  ctx.font = 'bold 52px sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(name.length > 18 ? `${name.slice(0, 17)}…` : name, 256, 52, 490)
  return tex
}
