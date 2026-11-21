import { useCallback, useEffect, useState } from 'react'
import { ApiError } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'
import type { Sheet } from '../sheet-assimilacao/types'
import { Autosaver, type AutosaverHooks } from './Autosaver'
import { saveSheet, type CharacterView, type ServerIssue } from './characters'

export type SaveStatus = 'saved' | 'saving' | 'error'

function rejectionIssues(e: unknown): ServerIssue[] | null {
  return e instanceof ApiError && Array.isArray(e.data.issues) ? (e.data.issues as ServerIssue[]) : null
}

interface Setters {
  sheet: (s: Sheet) => void
  issues: (fn: (cur: ServerIssue[]) => ServerIssue[]) => void
  status: (s: SaveStatus) => void
  error: (e: string) => void
}

function makeHooks(campaignId: string, id: string, set: Setters): AutosaverHooks {
  return {
    save: (s) => saveSheet(campaignId, id, s),
    onSaving: () => set.status('saving'),
    onRemote: (c) => { set.sheet(c.sheet); set.issues(() => c.issues) },
    onSaved: (c, unsaved) => {
      set.issues(() => c.issues)
      set.error('')
      set.status(unsaved ? 'saving' : 'saved')
    },
    onFailed: (e) => {
      set.error(e instanceof Error ? e.message : 'Erro ao salvar')
      set.issues((cur) => rejectionIssues(e) ?? cur)
      set.status('error')
    },
  }
}

/** Local sheet state with debounced autosave and live updates from character.updated. */
export function useCharacterSync(campaignId: string, initial: CharacterView, canEdit: boolean) {
  const { subscribe } = useRealtime()
  const [sheet, setSheet] = useState(initial.sheet)
  const [issues, setIssues] = useState(initial.issues)
  const [status, setStatus] = useState<SaveStatus>('saved')
  const [error, setError] = useState('')
  const [saver] = useState(() => new Autosaver(initial, makeHooks(campaignId, initial.id,
    { sheet: setSheet, issues: setIssues, status: setStatus, error: setError })))
  useEffect(() => subscribe('character.updated', (p) => {
    if ((p as CharacterView).id === initial.id) saver.remote(p as CharacterView)
  }), [subscribe, initial.id, saver])
  useEffect(() => () => void saver.flush(), [saver])
  const onChange = useCallback((next: Sheet) => {
    if (!canEdit) return
    setSheet(next)
    saver.edit(next)
  }, [canEdit, saver])
  return { sheet, issues, status, error, onChange }
}
