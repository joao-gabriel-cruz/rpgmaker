import { useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { AssimilacaoSheet, type AssimilacaoSheetProps } from '../sheet-assimilacao/AssimilacaoSheet'
import type { RollRequest } from '../sheet-assimilacao/types'
import { ModifierPanel } from './ModifierPanel'
import { modifiersPayload, NO_MODIFIERS, useRollState, type Modifiers } from './modifiers'

type Props = Omit<AssimilacaoSheetProps, 'onRollRequest'> & { campaignId: string; characterId: string; isGm: boolean }

/** The sheet wired to `roll.request`, with the modifiers panel; the Mestre can roll in secret. */
export function RollableSheet({ campaignId, characterId, isGm, ...sheetProps }: Props) {
  const { send } = useRealtime()
  const [secret, setSecret] = useState(false)
  const [error, setError] = useState('')
  const [mods, setMods] = useState<Modifiers>(NO_MODIFIERS)
  const state = useRollState(campaignId, characterId)
  const onRoll = (req: RollRequest) => {
    setError('')
    const body = { characterId, ...req, ...modifiersPayload(mods, isGm), secret: isGm && secret }
    send('roll.request', body).then(() => setMods(NO_MODIFIERS)).catch((e: Error) => setError(e.message))
  }
  const setDifficulty = (d12: number) => {
    setError('')
    send('roll.difficulty.set', { characterId, d12 }).catch((e: Error) => setError(e.message))
  }
  return (
    <>
      {!sheetProps.readOnly && <ModifierPanel value={mods} onChange={setMods} state={state} isGm={isGm} onDifficulty={setDifficulty} />}
      {isGm && !sheetProps.readOnly && (
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={secret} onChange={(e) => setSecret(e.target.checked)} /> Rolagem secreta
        </label>
      )}
      {error && <p role="alert" className="rounded bg-red-950 px-3 py-2 text-sm text-red-300">{error}</p>}
      <AssimilacaoSheet {...sheetProps} onRollRequest={onRoll} />
    </>
  )
}
