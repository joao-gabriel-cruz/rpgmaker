import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import type { CharacterView } from '../characters/characters'
import { useTableCharacterOptional } from '../hotbar/TableCharacterContext'
import type { ActionTarget, GameAction } from './actions'

export interface PendingAction {
  action: GameAction
  character: CharacterView
  /** Set once the target was picked (or immediately for actions without target). */
  target?: ActionTarget
}

export interface ActionFlowApi {
  pending: PendingAction | null
  /** Starts an action: aim at a target on the map when it needs one, else open the composer right away. */
  begin: (action: GameAction, character: CharacterView) => void
  setTarget: (target: ActionTarget) => void
  cancel: () => void
  /** Why the cell under the cursor cannot be targeted (empty when fine). */
  note: string
  setNote: (note: string) => void
}

const Ctx = createContext<ActionFlowApi | null>(null)

export function useActionFlow(): ActionFlowApi {
  const api = useContext(Ctx)
  if (!api) throw new Error('useActionFlow must be used inside <ActionFlowProvider>')
  return api
}

/** Same as `useActionFlow`, null outside the table. */
export const useActionFlowOptional = () => useContext(Ctx)

/** Holds the action being aimed/composed; the map's aim layer and the hotbar's composer both read it. */
export function ActionFlowProvider({ children }: { children: ReactNode }) {
  const [pending, setPending] = useState<PendingAction | null>(null)
  const [note, setNote] = useState('')
  const tc = useTableCharacterOptional()
  const closeComposer = tc?.setComposerOpen
  const begin = useCallback((action: GameAction, character: CharacterView) => {
    closeComposer?.(false)
    setPending({ action, character, ...(action.target === 'none' ? { target: { kind: 'none' as const } } : {}) })
  }, [closeComposer])
  const setTarget = useCallback((target: ActionTarget) => setPending((p) => p && { ...p, target }), [])
  const cancel = useCallback(() => { setPending(null); setNote('') }, [])
  const value = useMemo(() => ({ pending, begin, setTarget, cancel, note, setNote }), [pending, begin, setTarget, cancel, note])
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}
