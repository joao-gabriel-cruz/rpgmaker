import { Modal } from '../table/Modal'
import { CATALOG, materialCost, takenKinds, type Building, type Proposal } from './buildings'

interface Props {
  isGm: boolean
  buildings: Building[]
  proposals: Proposal[]
  onPick: (kind: string) => void
  onClose: () => void
}

/** Catalog of buildings as cards: the Mestre builds directly, players propose. */
export function NewBuildingDialog({ isGm, buildings, proposals, onPick, onClose }: Props) {
  const taken = takenKinds(buildings, proposals)
  return (
    <Modal title="Nova construção" onClose={onClose} wide>
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        {CATALOG.map((d) => (
          <article key={d.kind} aria-label={d.name} className="flex flex-col gap-1 rounded-lg bg-zinc-800 p-3">
            <h3 className="font-semibold text-zinc-100">{d.name}</h3>
            <p className="text-xs text-emerald-300">{d.cost}{d.leveled ? ' × nível' : ''} Pontos de Obra · {materialCost(d, 1)} Material de Construção</p>
            <p className="flex-1 text-sm text-zinc-400">{d.effect}</p>
            <button type="button" disabled={taken.has(d.kind)} onClick={() => { onPick(d.kind); onClose() }}
              className="self-start rounded bg-emerald-600 px-3 py-1 text-sm hover:bg-emerald-500 disabled:bg-zinc-700 disabled:text-zinc-400">
              {taken.has(d.kind) ? 'Já existe ou proposta' : isGm ? 'Construir' : 'Propor'}</button>
          </article>
        ))}
      </div>
    </Modal>
  )
}
