import { useState } from 'react'
import type { CharacterView } from '../characters/characters'
import { CharacterAvatar } from '../miniature/CharacterAvatar'
import { ROLE_LABEL, ROLES, useRefugeCommand, type ResidentRole } from './crises'
import type { Refuge } from './refuge'

const field = 'rounded bg-zinc-800 px-2 py-1 text-sm'

function AddResident({ npcs, run }: { npcs: CharacterView[]; run: (type: string, payload?: object) => void }) {
  const [characterId, setCharacterId] = useState('')
  const [role, setRole] = useState<ResidentRole>('other')
  const [note, setNote] = useState('')
  const chosen = characterId || npcs[0]?.id || ''
  return (
    <form className="flex flex-wrap items-end gap-2" onSubmit={(e) => { e.preventDefault(); if (chosen) { run('refuge.resident.add', { characterId: chosen, role, note }); setNote('') } }}>
      <select aria-label="NPC morador" value={chosen} onChange={(e) => setCharacterId(e.target.value)} className={field}>
        {npcs.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
      </select>
      <select aria-label="Função do morador" value={role} onChange={(e) => setRole(e.target.value as ResidentRole)} className={field}>
        {ROLES.map((r) => <option key={r} value={r}>{ROLE_LABEL[r]}</option>)}
      </select>
      <input aria-label="Nota do morador" placeholder="Nota" maxLength={500} value={note} onChange={(e) => setNote(e.target.value)} className={field} />
      <button disabled={!chosen} className="rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500 disabled:opacity-40">Adicionar morador</button>
    </form>
  )
}

/** Notable residents: campaign NPCs with a role, shown beside the abstract Population. */
export function ResidentsSection({ refuge, isGm, everyone }: { refuge: Refuge; isGm: boolean; everyone: CharacterView[] }) {
  const { run, error } = useRefugeCommand()
  const residents = refuge.residents ?? []
  const free = everyone.filter((c) => c.kind === 'npc' && !residents.some((r) => r.characterId === c.id))
  return (
    <section aria-label="Moradores" className="flex max-w-3xl flex-col gap-2">
      <h3 className="text-sm text-zinc-400">Moradores notáveis ({residents.length})</h3>
      <ul className="flex flex-col gap-2">
        {residents.map((r) => {
          const c = everyone.find((x) => x.id === r.characterId)
          return (
            <li key={r.id} data-testid="resident" className="flex items-center gap-3 rounded bg-zinc-900 px-3 py-2">
              <CharacterAvatar name={r.name} appearance={c?.sheet?.appearance} portraitUrl={c?.sheet?.identity?.portraitUrl} size="h-10 w-10" />
              <span className="flex flex-col"><span className="font-medium">{r.name}</span><span className="text-xs text-zinc-400">{ROLE_LABEL[r.role]}{r.note && ` · ${r.note}`}</span></span>
              {isGm && <button className="ml-auto rounded bg-zinc-800 px-2 py-1 text-xs hover:bg-zinc-700" onClick={() => run('refuge.resident.remove', { id: r.id })}>Remover</button>}
            </li>
          )
        })}
        {residents.length === 0 && <li className="text-sm text-zinc-400">Nenhum morador notável.</li>}
      </ul>
      {isGm && <AddResident npcs={free} run={run} />}
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
    </section>
  )
}
