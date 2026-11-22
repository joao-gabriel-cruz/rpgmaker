import { useState } from 'react'
import type { CharacterView } from '../characters/characters'
import { ABUNDANCE_LABEL, NATURAL_KEYS, RESOURCE_GROUPS, RESOURCE_LABEL, type Refuge, type ResourceKey } from './refuge'

const field = 'rounded bg-zinc-800 px-2 py-1 text-sm'
const RESOURCES = RESOURCE_GROUPS.flatMap((g) => g.keys)

interface Props {
  refuge: Refuge
  characters: CharacterView[]
  isGm: boolean
  onSubmit: (p: { characterIds: string[]; resource: ResourceKey; regionNote: string }) => void
  onClose: () => void
}

function hint(refuge: Refuge, key: ResourceKey) {
  const natural = (NATURAL_KEYS as string[]).includes(key)
  return natural ? ABUNDANCE_LABEL[refuge.region.abundance[key as keyof typeof refuge.region.abundance]] : 'Escasso (sem região)'
}

/** Form to send (Mestre) or propose (player) a supply run; the first chosen character leads. */
export function ExpeditionDialog({ refuge, characters, isGm, onSubmit, onClose }: Props) {
  const [chosen, setChosen] = useState<string[]>([])
  const [resource, setResource] = useState<ResourceKey>('water')
  const [note, setNote] = useState('')
  const toggle = (id: string) => setChosen((c) => (c.includes(id) ? c.filter((x) => x !== id) : [...c, id]))
  return (
    <form role="dialog" aria-label="Enviar expedição" className="flex flex-col gap-2 rounded border border-zinc-700 bg-zinc-900 p-3"
      onSubmit={(e) => { e.preventDefault(); if (chosen.length) { onSubmit({ characterIds: chosen, resource, regionNote: note.trim() }); onClose() } }}>
      <fieldset className="flex flex-wrap gap-3">
        <legend className="text-xs text-zinc-400">Quem vai (o primeiro escolhido lidera)</legend>
        {characters.map((c) => (
          <label key={c.id} className="flex items-center gap-1 text-sm">
            <input type="checkbox" checked={chosen.includes(c.id)} onChange={() => toggle(c.id)} />{c.name}
          </label>
        ))}
        {characters.length === 0 && <span className="text-sm text-zinc-400">Nenhum personagem disponível.</span>}
      </fieldset>
      <label className="flex flex-col gap-1 text-xs text-zinc-400">Recurso procurado
        <select aria-label="Recurso procurado" value={resource} onChange={(e) => setResource(e.target.value as ResourceKey)} className={field}>
          {RESOURCES.map((k) => <option key={k} value={k}>{RESOURCE_LABEL[k]} · {hint(refuge, k)}</option>)}
        </select>
      </label>
      <input aria-label="Região ou observação" placeholder="Para onde? (opcional)" maxLength={300} value={note} onChange={(e) => setNote(e.target.value)} className={field} />
      <div className="flex gap-2">
        <button disabled={chosen.length === 0} className="rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500 disabled:opacity-40">{isGm ? 'Enviar' : 'Propor'}</button>
        <button type="button" onClick={onClose} className="rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700">Cancelar</button>
      </div>
    </form>
  )
}
