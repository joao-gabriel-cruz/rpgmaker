import { useState } from 'react'
import type { CharacterView } from '../characters/characters'
import { useRefugeCommand } from './crises'
import { ExpeditionCard } from './ExpeditionCard'
import { ExpeditionDialog } from './ExpeditionDialog'
import type { Refuge } from './refuge'

interface Props {
  refuge: Refuge
  isGm: boolean
  userId?: string
  /** Characters that may go: every PC for the Mestre, the player's own otherwise. */
  candidates: CharacterView[]
  /** Every character, to name the party and find the leader's owner. */
  everyone: CharacterView[]
}

/** Supply runs: send or propose one, roll it, then let the Mestre spend Adaptations and Pressures. */
export function ExpeditionsSection({ refuge, isGm, userId, candidates, everyone }: Props) {
  const { run, error } = useRefugeCommand()
  const [open, setOpen] = useState(false)
  const list = [...(refuge.expeditions ?? [])].reverse()
  const ownerOf = (id: string) => everyone.find((c) => c.id === id)?.ownerId
  return (
    <section aria-label="Expedições" className="flex max-w-3xl flex-col gap-2">
      <h3 className="text-sm text-zinc-400">Expedições ({list.length})</h3>
      <ul className="flex flex-col gap-2">
        {list.map((e) => (
          <ExpeditionCard key={e.id} e={e} isGm={isGm} party={everyone} run={run}
            canRoll={isGm || (!!userId && ownerOf(e.characterIds[0]) === userId)} canCancel={isGm || e.proposedBy === userId} />
        ))}
        {list.length === 0 && <li className="text-sm text-zinc-400">Nenhuma expedição ainda.</li>}
      </ul>
      {open
        ? <ExpeditionDialog refuge={refuge} characters={candidates} isGm={isGm} onClose={() => setOpen(false)} onSubmit={(p) => run('refuge.expedition.start', p)} />
        : <button className="w-fit rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500" onClick={() => setOpen(true)}>Enviar expedição</button>}
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
    </section>
  )
}
