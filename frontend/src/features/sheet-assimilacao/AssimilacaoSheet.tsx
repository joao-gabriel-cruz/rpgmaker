import { useState } from 'react'
import { AttributesSection } from './AttributesSection'
import { AssimilationsSection, CharacteristicsSection, NotesSection } from './CharacteristicsSection'
import { HealthSection } from './HealthSection'
import { IdentitySection } from './IdentitySection'
import { InventorySection } from './InventorySection'
import { TugOfWarSection } from './TugOfWarSection'
import type { RollRequest, Sheet, SheetIssue } from './types'
import { validateSheet } from './validate'
import { APTITUDE_LABELS, INSTINCT_LABELS } from './labels'

export interface AssimilacaoSheetProps {
  sheet: Sheet
  onChange: (next: Sheet) => void
  readOnly?: boolean
  issues?: SheetIssue[]
  onRollRequest?: (req: RollRequest) => void
  /** When set, each inventory item gets a "Guardar no Refúgio" button. */
  onDepositItem?: (place: 'body' | 'backpack', index: number) => void
}

function RollBar({ selection, onRoll }: { selection: Partial<RollRequest>; onRoll: () => void }) {
  const { instinct, aptitude } = selection
  return (
    <div className="sticky bottom-2 flex items-center justify-between rounded-lg border border-emerald-800 bg-zinc-900 p-3 text-sm">
      <span>{instinct ? INSTINCT_LABELS[instinct] : 'Instinto'} + {aptitude ? APTITUDE_LABELS[aptitude] : 'Aptidão'}</span>
      <button className="rounded bg-emerald-600 px-3 py-1 font-medium text-white disabled:opacity-40" disabled={!instinct || !aptitude} onClick={onRoll}>Rolar</button>
    </div>
  )
}

export function AssimilacaoSheet({ sheet, onChange, readOnly = false, issues, onRollRequest, onDepositItem }: AssimilacaoSheetProps) {
  const [selection, setSelection] = useState<Partial<RollRequest>>({})
  const common = { sheet, onChange, readOnly, issues: issues ?? validateSheet(sheet) }
  const rollable = !readOnly && !!onRollRequest
  const roll = () => {
    if (!selection.instinct || !selection.aptitude) return
    onRollRequest?.({ instinct: selection.instinct, aptitude: selection.aptitude })
    setSelection({})
  }
  return (
    <div className="mx-auto grid max-w-6xl grid-cols-1 gap-4 text-zinc-100 lg:grid-cols-2">
      <div className="lg:col-span-2"><IdentitySection {...common} /></div>
      <div className="space-y-4">
        <AttributesSection {...common} selection={selection} rollable={rollable} onSelect={setSelection} />
        {rollable && <RollBar selection={selection} onRoll={roll} />}
      </div>
      <div className="space-y-4">
        <TugOfWarSection {...common} />
        <HealthSection {...common} />
      </div>
      <div className="lg:col-span-2"><InventorySection {...common} onDeposit={readOnly ? undefined : onDepositItem} /></div>
      <CharacteristicsSection {...common} />
      <div className="space-y-4">
        <AssimilationsSection {...common} />
        <NotesSection {...common} />
      </div>
    </div>
  )
}
