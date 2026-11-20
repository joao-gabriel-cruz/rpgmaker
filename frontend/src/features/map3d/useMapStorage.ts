import { useState } from 'react'
import { createGrid, deserializeGrid, serializeGrid, type MapGrid } from './grid'

const STORAGE_KEY = 'rpgmaker.map-demo'

export function useMapStorage() {
  const [grid, setGrid] = useState<MapGrid>(() => createGrid(32, 32))
  const [msg, setMsg] = useState('')

  const save = () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(serializeGrid(grid)))
    setMsg('Mapa salvo.')
  }
  const load = () => {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return setMsg('Nenhum mapa salvo.')
    try {
      setGrid(deserializeGrid(JSON.parse(raw)))
      setMsg('Mapa carregado.')
    } catch {
      setMsg('Mapa salvo inválido.')
    }
  }
  return { grid, setGrid, msg, save, load }
}
