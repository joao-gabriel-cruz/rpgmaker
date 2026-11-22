import { useState } from 'react'
import type { CharacterView } from '../characters/characters'
import { EXPEDITION_STATUS_LABEL, type Expedition } from './crises'
import { RESOURCE_GROUPS, RESOURCE_LABEL, type ResourceKey } from './refuge'

type Run = (type: string, payload?: object) => void
const btn = 'rounded bg-zinc-800 px-2 py-1 text-xs hover:bg-zinc-700'
const field = 'rounded bg-zinc-800 px-2 py-1 text-xs'
const RESOURCES = RESOURCE_GROUPS.flatMap((g) => g.keys)

function Extra({ e, run }: { e: Expedition; run: Run }) {
  const [resource, setResource] = useState<ResourceKey>('food')
  return (
    <div className="flex flex-wrap items-center gap-2 text-xs">
      <span>{e.extras} Adaptação(ões): +1 de</span>
      <select aria-label="Recurso extra" value={resource} onChange={(ev) => setResource(ev.target.value as ResourceKey)} className={field}>
        {RESOURCES.map((k) => <option key={k} value={k}>{RESOURCE_LABEL[k]}</option>)}
      </select>
      <button className={btn} onClick={() => run('refuge.expedition.extra', { id: e.id, resource })}>Dar extra</button>
    </div>
  )
}

function Complication({ e, party, run }: { e: Expedition; party: CharacterView[]; run: Run }) {
  const [kind, setKind] = useState('injury')
  const [characterId, setCharacterId] = useState(e.characterIds[0])
  const [resource, setResource] = useState<ResourceKey>('water')
  return (
    <div className="flex flex-wrap items-center gap-2 text-xs">
      <span>{e.complications} Pressão(ões):</span>
      <select aria-label="Complicação" value={kind} onChange={(ev) => setKind(ev.target.value)} className={field}>
        <option value="injury">Ferimento</option><option value="encounter">Encontro</option><option value="loss">Perda</option>
      </select>
      {kind === 'injury' && (
        <select aria-label="Ferido" value={characterId} onChange={(ev) => setCharacterId(ev.target.value)} className={field}>
          {e.characterIds.map((id) => <option key={id} value={id}>{party.find((c) => c.id === id)?.name ?? 'Personagem'}</option>)}
        </select>
      )}
      {kind === 'loss' && (
        <select aria-label="Recurso perdido" value={resource} onChange={(ev) => setResource(ev.target.value as ResourceKey)} className={field}>
          {RESOURCES.map((k) => <option key={k} value={k}>{RESOURCE_LABEL[k]}</option>)}
        </select>
      )}
      <button className={btn} onClick={() => run('refuge.expedition.complication', { id: e.id, kind, characterId, resource, amount: 1, note: 'Algo se meteu no caminho.' })}>Aplicar</button>
    </div>
  )
}

function RollButtons({ e, run }: { e: Expedition; run: Run }) {
  return (
    <div className="flex gap-2">
      <button className={btn} onClick={() => run('refuge.expedition.roll', { id: e.id, instinct: 'perception' })}>Rolar Percepção + Sobrevivência</button>
      <button className={btn} onClick={() => run('refuge.expedition.roll', { id: e.id, instinct: 'sagacity' })}>Rolar Sagacidade + Sobrevivência</button>
    </div>
  )
}

interface Props { e: Expedition; isGm: boolean; canRoll: boolean; canCancel: boolean; party: CharacterView[]; run: Run }

export function ExpeditionCard({ e, isGm, canRoll, canCancel, party, run }: Props) {
  return (
    <li data-testid="expedition-card" className="flex flex-col gap-2 rounded bg-zinc-900 px-3 py-2">
      <div className="flex flex-wrap items-center gap-2">
        <span className="font-medium">Expedição de {e.leaderName}</span>
        <span className="text-xs text-zinc-400">{RESOURCE_LABEL[e.resource as ResourceKey]} · semana {e.week}{e.regionNote && ` · ${e.regionNote}`}</span>
        <span className="ml-auto rounded bg-zinc-800 px-2 text-xs">{EXPEDITION_STATUS_LABEL[e.status]}</span>
      </div>
      <p className="text-xs text-zinc-400">Equipe: {e.characterIds.map((id) => party.find((c) => c.id === id)?.name ?? '?').join(', ')}</p>
      {e.log.map((l, i) => <p key={i} className="text-sm text-zinc-300">{l}</p>)}
      {isGm && e.status === 'proposed' && <button className={`${btn} w-fit`} onClick={() => run('refuge.expedition.approve', { id: e.id })}>Aprovar</button>}
      {canRoll && e.status === 'ready' && <RollButtons e={e} run={run} />}
      {isGm && e.status === 'done' && e.extras > 0 && <Extra e={e} run={run} />}
      {isGm && e.status === 'done' && e.complications > 0 && <Complication e={e} party={party} run={run} />}
      {canCancel && e.status !== 'rolling' && e.status !== 'done' && <button className={`${btn} w-fit`} onClick={() => run('refuge.expedition.cancel', { id: e.id })}>Cancelar</button>}
    </li>
  )
}
