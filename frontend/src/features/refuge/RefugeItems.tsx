import { useState } from 'react'
import type { CharacterView } from '../characters/characters'
import { BACKPACK_LIMIT, BODY_LIMIT, type Item } from '../sheet-assimilacao/types'

export type Place = 'body' | 'backpack'

interface Props {
  items: Item[]
  characters: CharacterView[]
  error: string
  onTake: (characterId: string, to: Place, refugeIndex: number) => void
}

const selectCls = 'rounded bg-zinc-800 px-2 py-1 text-sm'

function TakeForm({ characters, onTake }: { characters: CharacterView[]; onTake: (id: string, to: Place) => void }) {
  const [charId, setCharId] = useState('')
  const [to, setTo] = useState<Place>('backpack')
  const chosen = charId || characters[0]?.id || ''
  return (
    <span className="ml-auto flex items-center gap-2">
      <select aria-label="Personagem" className={selectCls} value={chosen} onChange={(e) => setCharId(e.target.value)}>
        {characters.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
      </select>
      <select aria-label="Destino" className={selectCls} value={to} onChange={(e) => setTo(e.target.value as Place)}>
        <option value="backpack">Mochila (máx. {BACKPACK_LIMIT})</option>
        <option value="body">Corpo (máx. {BODY_LIMIT})</option>
      </select>
      <button disabled={!chosen} className="rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500 disabled:opacity-40"
        onClick={() => onTake(chosen, to)}>Pegar</button>
    </span>
  )
}

export function RefugeItems({ items, characters, error, onTake }: Props) {
  return (
    <div className="flex max-w-3xl flex-col gap-2">
      <h3 className="text-sm text-zinc-400">Inventário do Refúgio ({items.length})</h3>
      <ul className="flex flex-col gap-2">
        {items.map((item, i) => (
          <li key={i} className="flex flex-wrap items-center gap-3 rounded bg-zinc-900 px-3 py-2">
            <span className="font-medium">{item.name || 'Item sem nome'}</span>
            <span className="text-xs text-zinc-400">Qualidade {item.quality} · Escassez {item.scarcity} · Pressão {item.pressure}</span>
            {item.traits?.map((t) => <span key={t} className="rounded-full bg-zinc-800 px-2 py-0.5 text-xs">{t}</span>)}
            {characters.length > 0 && <TakeForm characters={characters} onTake={(id, to) => onTake(id, to, i)} />}
          </li>
        ))}
        {items.length === 0 && <li className="text-sm text-zinc-400">Nenhum item guardado.</li>}
      </ul>
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
    </div>
  )
}
