import { CanvasTexture, SRGBColorSpace } from 'three'
import { isNumbered, symbolsOf, type DieKind } from './faces'
import { drawSymbols } from './symbolIcons'

function drawNumber(canvas: HTMLCanvasElement, face: number) {
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  ctx.fillStyle = '#18181b'
  ctx.font = 'bold 150px sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(String(face), 128, 134)
  if (face === 6 || face === 9) ctx.fillRect(84, 214, 88, 10)
}

const cache = new Map<string, CanvasTexture | null>()

// Transparent decal texture for one face; null when the face has no symbols.
export function faceTexture(kind: DieKind, face: number): CanvasTexture | null {
  const key = `${kind}:${face}`
  if (cache.has(key)) return cache.get(key) ?? null
  const symbols = symbolsOf(kind, face)
  let texture: CanvasTexture | null = null
  if (symbols.length > 0 || isNumbered(kind)) {
    const canvas = document.createElement('canvas')
    canvas.width = canvas.height = 256
    if (isNumbered(kind)) drawNumber(canvas, face)
    else drawSymbols(canvas, symbols)
    texture = new CanvasTexture(canvas)
    texture.colorSpace = SRGBColorSpace
  }
  cache.set(key, texture)
  return texture
}
