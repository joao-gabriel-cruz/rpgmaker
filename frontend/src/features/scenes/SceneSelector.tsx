import { useState, type FormEvent } from 'react'
import { SCENE_SIZES, type SceneSummary } from './scenes'

interface Props {
  scenes: SceneSummary[]
  selectedId: string | null
  onSelect: (id: string) => void
  onCreate: (name: string, size: number) => void
  onRename: (id: string, name: string) => void
  onDelete: (id: string) => void
  onActivate: (id: string) => void
}

const btn = 'rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700 disabled:opacity-50'
const input = 'rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm'

function CreateForm({ onCreate }: Pick<Props, 'onCreate'>) {
  const [name, setName] = useState('')
  const [size, setSize] = useState(32)
  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    onCreate(name.trim(), size)
    setName('')
  }
  return (
    <form onSubmit={submit} className="flex flex-wrap items-center gap-2">
      <input aria-label="Nome da cena" placeholder="Nome da nova cena" value={name} maxLength={80}
        onChange={(e) => setName(e.target.value)} className={input} />
      <select aria-label="Tamanho da cena" value={size} onChange={(e) => setSize(Number(e.target.value))} className={input}>
        {SCENE_SIZES.map((s) => <option key={s} value={s}>{s} × {s}</option>)}
      </select>
      <button type="submit" className={btn}>Criar cena</button>
    </form>
  )
}

function SceneRow({ scene, selected, ...p }: { scene: SceneSummary; selected: boolean } & Omit<Props, 'scenes' | 'selectedId' | 'onCreate'>) {
  const rename = () => {
    const name = window.prompt('Novo nome da cena', scene.name)?.trim()
    if (name && name !== scene.name) p.onRename(scene.id, name)
  }
  const remove = () => {
    if (window.confirm(`Apagar a cena "${scene.name}"?`)) p.onDelete(scene.id)
  }
  return (
    <li className={`flex flex-wrap items-center gap-2 rounded px-3 py-2 ${selected ? 'bg-zinc-800' : 'bg-zinc-900'}`}>
      <button onClick={() => p.onSelect(scene.id)} aria-pressed={selected} className="font-medium hover:underline">
        {scene.name}
      </button>
      <span className="text-xs text-zinc-400">{scene.width}×{scene.depth}</span>
      {scene.isActive && <span className="rounded bg-emerald-900 px-2 py-0.5 text-xs text-emerald-300">Ativa</span>}
      <span className="ml-auto flex gap-2">
        {!scene.isActive && <button className={btn} onClick={() => p.onActivate(scene.id)} aria-label={`Ativar ${scene.name}`}>Ativar</button>}
        <button className={btn} onClick={rename} aria-label={`Renomear ${scene.name}`}>Renomear</button>
        <button className={`${btn} hover:bg-red-900`} onClick={remove} aria-label={`Apagar ${scene.name}`}>Apagar</button>
      </span>
    </li>
  )
}

export function SceneSelector({ scenes, selectedId, onCreate, ...rest }: Props) {
  return (
    <div className="flex flex-col gap-3">
      <CreateForm onCreate={onCreate} />
      {scenes.length === 0 && <p className="text-sm text-zinc-400">Nenhuma cena ainda. Crie a primeira.</p>}
      <ul className="flex max-w-3xl flex-col gap-2">
        {scenes.map((s) => <SceneRow key={s.id} scene={s} selected={s.id === selectedId} {...rest} />)}
      </ul>
    </div>
  )
}
