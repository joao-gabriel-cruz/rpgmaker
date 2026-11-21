import { createContext, useCallback, useContext, useMemo, useRef, useState, type ReactNode } from 'react'
import type { SceneTokens } from '../tokens/useSceneTokens'

/** Shared state of the table's active character: selection, sheet drawer and roll composer. */
export interface TableCharacterApi {
  selectedId: string | null
  select: (id: string | null) => void
  sheetId: string | null
  openSheet: (id: string) => void
  closeSheet: () => void
  composerOpen: boolean
  setComposerOpen: (open: boolean) => void
  /** The hotbar registers how to repeat the last roll of the active character. */
  registerRollAgain: (fn: (() => void) | null) => void
  rollAgain: () => void
}

const Ctx = createContext<TableCharacterApi | null>(null)
const TokensCtx = createContext<{ tokens: SceneTokens | null; setTokens: (t: SceneTokens | null) => void }>({ tokens: null, setTokens: () => undefined })

export function useTableCharacter(): TableCharacterApi {
  const api = useContext(Ctx)
  if (!api) throw new Error('useTableCharacter must be used inside <TableCharacterProvider>')
  return api
}

/** Same as `useTableCharacter` but null outside the table (e.g. components rendered in isolation). */
export const useTableCharacterOptional = () => useContext(Ctx)

/** Tokens of the scene on screen; published by the map, read by the hotbar. */
export const useSceneTokensBridge = () => useContext(TokensCtx)

export function TableCharacterProvider({ children }: { children: ReactNode }) {
  const [selectedId, select] = useState<string | null>(null)
  const [sheetId, setSheetId] = useState<string | null>(null)
  const [composerOpen, setComposerOpen] = useState(false)
  const [tokens, setTokens] = useState<SceneTokens | null>(null)
  const again = useRef<(() => void) | null>(null)
  const registerRollAgain = useCallback((fn: (() => void) | null) => { again.current = fn }, [])
  const rollAgain = useCallback(() => again.current?.(), [])
  const openSheet = useCallback((id: string) => { select(id); setSheetId(id) }, [])
  const closeSheet = useCallback(() => setSheetId(null), [])
  const value = useMemo<TableCharacterApi>(
    () => ({ selectedId, select, sheetId, openSheet, closeSheet, composerOpen, setComposerOpen, registerRollAgain, rollAgain }),
    [selectedId, sheetId, openSheet, closeSheet, composerOpen, registerRollAgain, rollAgain],
  )
  const tokensValue = useMemo(() => ({ tokens, setTokens }), [tokens])
  return <Ctx.Provider value={value}><TokensCtx.Provider value={tokensValue}>{children}</TokensCtx.Provider></Ctx.Provider>
}
