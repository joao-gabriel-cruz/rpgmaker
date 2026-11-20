import { CanvasTexture, SRGBColorSpace } from 'three'
import { symbolsOf, type DieKind } from './faces'
import { drawSymbols } from './symbolIcons'

const cache = new Map<string, CanvasTexture | null>()

// Transparent decal texture for one face; null when the face has no symbols.
export function faceTexture(kind: DieKind, face: number): CanvasTexture | null {
  const key = `${kind}:${face}`
  if (cache.has(key)) return cache.get(key) ?? null
  const symbols = symbolsOf(kind, face)
  let texture: CanvasTexture | null = null
  if (symbols.length > 0) {
    const canvas = document.createElement('canvas')
    canvas.width = canvas.height = 256
    drawSymbols(canvas, symbols)
    texture = new CanvasTexture(canvas)
    texture.colorSpace = SRGBColorSpace
  }
  cache.set(key, texture)
  return texture
}
