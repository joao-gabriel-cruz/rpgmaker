import { Crosshair, X } from 'lucide-react'
import { RollComposer, type RollOption, type RollPayload } from '../rolls/RollComposer'
import type { PoolSelection } from '../rolls/pool'
import { useRealtime } from '../realtime/realtimeContext'
import { useEffect } from 'react'
import { useActionFlow } from './ActionFlow'
import { bonusSource, keptBonusFor } from './actions'

/** Esc cancels aiming; this strip tells the player what is being aimed and offers the same exit as a button. */
export function AimHint() {
  const { pending, cancel, note } = useActionFlow()
  const aiming = !!pending && !pending.target
  useEffect(() => {
    if (!aiming) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') cancel() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [aiming, cancel])
  if (!pending || pending.target) return null
  return (
    <div role="status" className="flex items-center gap-2 px-3 pb-1 pt-1 text-sm text-emerald-300">
      <Crosshair size={16} aria-hidden />
      <span>{pending.action.name}: escolha o alvo no mapa{pending.action.range ? ` (até ${pending.action.range} células)` : ''}. <kbd className="font-mono text-xs text-zinc-300">Esc</kbd> cancela.</span>
      {note && <span aria-live="polite" className="text-amber-300">{note}</span>}
      <button type="button" aria-label="Cancelar mira" onClick={cancel} className="ml-auto rounded p-1 text-zinc-400 hover:bg-zinc-800 hover:text-white"><X size={14} aria-hidden /></button>
    </div>
  )
}

/** The roll composer pre-filled by the action: default roll marked, the others as chips, modifiers available. */
export function ActionComposer({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const { pending, cancel } = useActionFlow()
  const { send } = useRealtime()
  if (!pending?.target) return null
  const { action, character, target } = pending
  const alternatives: RollOption[] = action.rolls.map((r) => ({ instinct: r.instinct, aptitude: r.aptitude, label: r.label }))
  const bonus = keptBonusFor(character.sheet, action.category)
  const source = bonus ? bonusSource(character.sheet, action.category) : ''
  const submit = (body: RollPayload, sel: PoolSelection) => {
    const rollChoice = Math.max(0, action.rolls.findIndex((r) => r.instinct === sel.instinct && r.aptitude === sel.aptitude))
    const { effort, useOccupation, useMarkingEvent, secondInstinct } = body
    return send('action.perform', {
      characterId: character.id, actionId: action.id, target, rollChoice,
      modifiers: { effort, useOccupation, useMarkingEvent, ...(secondInstinct ? { secondInstinct } : {}) },
    })
  }
  return (
    <RollComposer key={`${action.id}:${character.id}`} campaignId={campaignId} character={character} isGm={isGm} title={`${character.name}: ${action.name}`}
      preset={{ instinct: action.rolls[0].instinct, aptitude: action.rolls[0].aptitude }} alternatives={alternatives} lockToAlternatives
      bonusKept={bonus} hint={bonus ? `+${bonus} dado mantido${source ? ` (${source})` : ''}` : undefined} submit={submit} onClose={cancel} />
  )
}
