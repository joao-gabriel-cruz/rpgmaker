import { useState } from 'react'
import type { CharacterView } from '../characters/characters'
import { useRealtime } from '../realtime/realtimeContext'
import { fieldCls } from './EffectsFields'
import type { CompendiumItem } from './items'

interface Props {
  item: CompendiumItem
  characters: CharacterView[]
  onClose: () => void
}

type Place = 'body' | 'backpack'

const payloadOf = (itemId: string, target: string, place: Place, qty: number) =>
  target === 'refuge' ? { itemId, target: 'refuge', qty } : { itemId, characterId: target, place, qty }

function PlaceSelect({ place, onChange }: { place: Place; onChange: (p: Place) => void }) {
  return (
    <label className="flex items-center gap-2">Onde
      <select className={fieldCls} value={place} onChange={(e) => onChange(e.target.value as Place)}>
        <option value="backpack">Mochila</option>
        <option value="body">Corpo</option>
      </select>
    </label>
  )
}

/** "Entregar…": pick a character (or the Refúgio), body/backpack and quantity, then send item.give. */
export function GiveForm({ item, characters, onClose }: Props) {
  const { send } = useRealtime()
  const [target, setTarget] = useState(characters[0]?.id ?? 'refuge')
  const [place, setPlace] = useState<Place>('backpack')
  const [qty, setQty] = useState(1)
  const [error, setError] = useState('')
  const give = () => {
    setError('')
    send('item.give', payloadOf(item.id, target, place, qty)).then(onClose, (e: Error) => setError(e.message))
  }
  return (
    <div role="group" aria-label={`Entregar ${item.name}`} className="flex flex-col gap-2 rounded-lg bg-zinc-800 p-2 text-sm">
      <label className="flex items-center gap-2">Para
        <select className={`${fieldCls} flex-1`} value={target} onChange={(e) => setTarget(e.target.value)}>
          {characters.map((c) => <option key={c.id} value={c.id}>{c.name}{c.kind === 'npc' ? ' (NPC)' : ''}</option>)}
          <option value="refuge">Refúgio</option>
        </select>
      </label>
      {target !== 'refuge' && <PlaceSelect place={place} onChange={setPlace} />}
      <label className="flex items-center gap-2">Quantidade
        <input type="number" min={1} max={6} className={`${fieldCls} w-16`} value={qty} onChange={(e) => setQty(Math.min(6, Math.max(1, Number(e.target.value) || 1)))} />
      </label>
      {error && <p role="alert" className="text-red-400">{error}</p>}
      <div className="flex gap-2">
        <button type="button" onClick={give} className="rounded bg-emerald-700 px-3 py-1 hover:bg-emerald-600">Entregar</button>
        <button type="button" onClick={onClose} className="rounded bg-zinc-700 px-3 py-1 hover:bg-zinc-600">Cancelar</button>
      </div>
    </div>
  )
}
