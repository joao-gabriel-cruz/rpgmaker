import { useState, type FormEvent } from 'react'
import { SCENE_SIZES } from '../scenes/scenes'

const input = 'mt-1 block w-full rounded-lg border border-zinc-700 bg-zinc-900 px-2 py-1.5 text-sm text-zinc-100'

/** Blank scene: just a name and a square size. */
export function BlankSceneForm({ error, onCreate }: { error: string; onCreate: (name: string, size: number) => void }) {
  const [name, setName] = useState('')
  const [size, setSize] = useState(32)
  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (name.trim()) onCreate(name.trim(), size)
  }
  return (
    <form onSubmit={submit} className="flex max-w-sm flex-col gap-3">
      <label className="text-xs text-zinc-400">Nome da cena
        <input value={name} maxLength={80} placeholder="Nome da nova cena" onChange={(e) => setName(e.target.value)} className={input} />
      </label>
      <label className="text-xs text-zinc-400">Tamanho da cena
        <select value={size} onChange={(e) => setSize(Number(e.target.value))} className={input}>
          {SCENE_SIZES.map((s) => <option key={s} value={s}>{s} × {s}</option>)}
        </select>
      </label>
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      <button type="submit" disabled={!name.trim()}
        className="self-start rounded-lg bg-emerald-600 px-3 py-1.5 text-sm font-medium text-white transition-[background-color,transform] hover:bg-emerald-500 active:scale-[0.96] disabled:opacity-50">Criar cena em branco</button>
    </form>
  )
}
