import { useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { AssimilacaoSheet, type AssimilacaoSheetProps } from '../sheet-assimilacao/AssimilacaoSheet'
import type { RollRequest } from '../sheet-assimilacao/types'

type Props = Omit<AssimilacaoSheetProps, 'onRollRequest'> & { characterId: string; isGm: boolean }

/** The sheet wired to `roll.request`; the Mestre can roll in secret. */
export function RollableSheet({ characterId, isGm, ...sheetProps }: Props) {
  const { send } = useRealtime()
  const [secret, setSecret] = useState(false)
  const [error, setError] = useState('')
  const onRoll = (req: RollRequest) => {
    setError('')
    send('roll.request', { characterId, ...req, secret: isGm && secret }).catch((e: Error) => setError(e.message))
  }
  return (
    <>
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
