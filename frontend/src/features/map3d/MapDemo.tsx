import { applyPatch } from './grid'
import { applyObjectPatch } from './objects/model'
import { TerrainEditor } from './TerrainEditor'
import { useMapStorage } from './useMapStorage'

const btn = 'rounded bg-zinc-800 px-2 py-1 hover:bg-zinc-700'

export function MapDemo() {
  const { grid, setGrid, objects, setObjects, msg, save, load } = useMapStorage()
  return (
    <div className="flex h-screen flex-col bg-zinc-950 text-zinc-100">
      <header className="flex items-center gap-2 border-b border-zinc-800 p-2 text-sm">
        <strong>Demo do mapa 3D</strong>
        <button className={btn} onClick={save}>Salvar</button>
        <button className={btn} onClick={load}>Carregar</button>
        <span className="text-zinc-400">{msg}</span>
        <span className="ml-auto text-zinc-500">Clique: usar ferramenta · R: girar objeto · Botão direito: girar câmera · Scroll: zoom</span>
      </header>
      <div className="min-h-0 flex-1">
        <TerrainEditor
          grid={grid}
          onPatch={(p) => setGrid((g) => applyPatch(g, p))}
          objects={objects}
          onObjectPatch={(p) => setObjects((o) => applyObjectPatch(o, p))}
        />
      </div>
    </div>
  )
}
