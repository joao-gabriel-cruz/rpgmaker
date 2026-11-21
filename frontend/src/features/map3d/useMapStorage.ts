import { useState } from 'react'
import { createGrid, deserializeGrid, serializeGrid, type MapGrid, type SerializedGrid } from './grid'
import { sanitizeObjects, type MapObject } from './objects/model'

const STORAGE_KEY = 'rpgmaker.map-demo'

interface Saved {
  grid: SerializedGrid
  objects: unknown
}

// Older saves stored only the serialized grid.
function parse(raw: string): { grid: MapGrid; objects: MapObject[] } {
  const json = JSON.parse(raw)
  const saved: Saved = 'grid' in json ? json : { grid: json, objects: [] }
  const grid = deserializeGrid(saved.grid)
  return { grid, objects: sanitizeObjects(saved.objects, grid) }
}

export function useMapStorage() {
  const [grid, setGrid] = useState<MapGrid>(() => createGrid(32, 32))
  const [objects, setObjects] = useState<readonly MapObject[]>([])
  const [msg, setMsg] = useState('')

  const save = () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ grid: serializeGrid(grid), objects }))
    setMsg('Mapa salvo.')
  }
  const load = () => {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return setMsg('Nenhum mapa salvo.')
    try {
      const loaded = parse(raw)
      setGrid(loaded.grid)
      setObjects(loaded.objects)
      setMsg('Mapa carregado.')
    } catch {
      setMsg('Mapa salvo inválido.')
    }
  }
  return { grid, setGrid, objects, setObjects, msg, save, load }
}
