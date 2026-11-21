import { Plus } from 'lucide-react'
import type { SceneSummary } from './scenes'

interface Props {
  scenes: SceneSummary[]
  selectedId: string | null
  onSelect: (id: string) => void
  onNew: () => void
  onRename: (id: string, name: string) => void
  onDelete: (id: string) => void
  onActivate: (id: string) => void
}

const btn = 'rounded-lg bg-zinc-800 px-2.5 py-1 text-xs transition-[background-color,transform] hover:bg-zinc-700 active:scale-[0.96] disabled:opacity-50'

function NewSceneButton({ onNew }: Pick<Props, 'onNew'>) {
  return (
    <button type="button" onClick={onNew} className="flex items-center justify-center gap-1.5 rounded-lg bg-emerald-600 px-3 py-1.5 text-sm font-medium text-white transition-[background-color,transform] hover:bg-emerald-500 active:scale-[0.96]">
      <Plus size={16} aria-hidden />Nova cena
    </button>
  )
}

function SceneRow({ scene, selected, ...p }: { scene: SceneSummary; selected: boolean } & Omit<Props, 'scenes' | 'selectedId' | 'onNew'>) {
  const rename = () => {
    const name = window.prompt('Novo nome da cena', scene.name)?.trim()
    if (name && name !== scene.name) p.onRename(scene.id, name)
  }
  const remove = () => {
    if (window.confirm(`Apagar a cena "${scene.name}"?`)) p.onDelete(scene.id)
  }
  return (
    <li className={`flex flex-wrap items-center gap-2 rounded-lg px-2.5 py-1.5 ${selected ? 'bg-zinc-800' : 'bg-zinc-900/60'}`}>
      <button onClick={() => p.onSelect(scene.id)} aria-pressed={selected} className="text-sm font-medium hover:underline">
        {scene.name}
      </button>
      <span className="text-xs text-zinc-400">{scene.width}×{scene.depth}</span>
      {scene.isActive && <span className="rounded bg-emerald-900 px-2 py-0.5 text-xs text-emerald-300">Ativa</span>}
      <span className="ml-auto flex gap-1.5">
        {!scene.isActive && <button className={btn} onClick={() => p.onActivate(scene.id)} aria-label={`Ativar ${scene.name}`}>Ativar</button>}
        <button className={btn} onClick={rename} aria-label={`Renomear ${scene.name}`}>Renomear</button>
        <button className={`${btn} hover:bg-red-900`} onClick={remove} aria-label={`Apagar ${scene.name}`}>Apagar</button>
      </span>
    </li>
  )
}

export function SceneSelector({ scenes, selectedId, onNew, ...rest }: Props) {
  return (
    <div className="flex flex-col gap-3">
      <NewSceneButton onNew={onNew} />
      {scenes.length === 0 && <p className="text-sm text-zinc-400">Nenhuma cena ainda. Crie a primeira.</p>}
      <ul className="flex flex-col gap-1.5">
        {scenes.map((s) => <SceneRow key={s.id} scene={s} selected={s.id === selectedId} {...rest} />)}
      </ul>
    </div>
  )
}
