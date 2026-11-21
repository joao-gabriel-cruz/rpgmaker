import { useCallback, useEffect, useRef, useState } from 'react'
import { applyPatch, type MapGrid, type MapPatch } from './grid'
import { inverseObjects, inversePaint, MAX_HISTORY, mergePaint, type Edit } from './history'
import { applyObjectPatch, isEmptyPatch, type MapObject, type ObjectPatch } from './objects/model'

interface Stacks {
  undo: Edit[]
  redo: Edit[]
}

export interface EditHistory {
  paint: (patch: MapPatch) => void
  editObjects: (patch: ObjectPatch) => void
  undo: () => void
  redo: () => void
  canUndo: boolean
  canRedo: boolean
  /** Kind of the edit that Undo would revert. */
  lastKind: Edit['kind'] | null
}

interface Sink {
  grid: MapGrid
  objects: readonly MapObject[]
  paint: (patch: MapPatch) => void
  editObjects: (patch: ObjectPatch) => void
}

type Side = 'undo' | 'redo'

function useStacks() {
  const [stacks, setStacks] = useState<Stacks>({ undo: [], redo: [] })
  const live = useRef(stacks)
  useEffect(() => { live.current = stacks }, [stacks])
  const commit = (next: Stacks) => {
    live.current = next
    setStacks(next)
  }
  // A new edit drops the redo branch; `merge` folds a brush patch into the open stroke.
  const push = useCallback((edit: Edit, merge: boolean) => {
    const { undo } = live.current
    const top = undo.at(-1)
    const folded = merge && top?.kind === 'paint' && edit.kind === 'paint' ? mergePaint(top, edit.undo, edit.redo) : null
    const list = folded ? [...undo.slice(0, -1), folded] : [...undo, edit]
    commit({ undo: list.slice(-MAX_HISTORY), redo: [] })
  }, [])
  // Moves the newest edit from one stack to the other and returns it.
  const take = useCallback((from: Side, to: Side) => {
    const edit = live.current[from].at(-1)
    if (edit) commit({ ...live.current, [from]: live.current[from].slice(0, -1), [to]: [...live.current[to], edit] })
    return edit
  }, [])
  return { stacks, push, take }
}

interface Shadow {
  grid: MapGrid
  objects: readonly MapObject[]
}
type Ref<T> = { current: T }
type Push = (edit: Edit, merge: boolean) => void

// Records each edit with its inverse, then sends it; `open` marks an unfinished brush stroke.
function useRecorder(sink: Sink, push: Push, shadowRef: Ref<Shadow>, openRef: Ref<boolean>) {
  const { paint: sendPaint, editObjects: sendObjects } = sink
  const paint = useCallback((patch: MapPatch) => {
    const undo = inversePaint(shadowRef.current.grid, patch)
    shadowRef.current.grid = applyPatch(shadowRef.current.grid, patch)
    push({ kind: 'paint', undo, redo: patch }, openRef.current)
    openRef.current = true
    sendPaint(patch)
  }, [push, sendPaint, shadowRef, openRef])
  const editObjects = useCallback((patch: ObjectPatch) => {
    if (isEmptyPatch(patch)) return
    const undo = inverseObjects(shadowRef.current.objects, patch)
    shadowRef.current.objects = applyObjectPatch(shadowRef.current.objects, patch)
    push({ kind: 'objects', undo, redo: patch }, false)
    openRef.current = false
    sendObjects(patch)
  }, [push, sendObjects, shadowRef, openRef])
  return { paint, editObjects }
}

// Sends the stored patch of the newest edit on one side and moves the edit to the other side.
function useReplay(sink: Sink, take: ReturnType<typeof useStacks>['take'], shadowRef: Ref<Shadow>, openRef: Ref<boolean>) {
  const { paint: sendPaint, editObjects: sendObjects } = sink
  const replay = useCallback((from: Side, to: Side) => {
    const edit = take(from, to)
    if (!edit) return
    openRef.current = false
    if (edit.kind === 'paint') {
      shadowRef.current.grid = applyPatch(shadowRef.current.grid, edit[from])
      sendPaint(edit[from])
    } else {
      shadowRef.current.objects = applyObjectPatch(shadowRef.current.objects, edit[from])
      sendObjects(edit[from])
    }
  }, [take, sendPaint, sendObjects, shadowRef, openRef])
  return { undo: useCallback(() => replay('undo', 'redo'), [replay]), redo: useCallback(() => replay('redo', 'undo'), [replay]) }
}

/**
 * Client-side undo/redo of the Mestre's brush strokes and object edits in this session.
 * It sends the inverse patch through the same commands; edits by other people are not tracked.
 */
export function useEditHistory(sink: Sink): EditHistory {
  const { stacks, push, take } = useStacks()
  // Shadow copies so two edits in the same tick (a drag stroke) see each other before React re-renders.
  const shadowRef = useRef<Shadow>({ grid: sink.grid, objects: sink.objects })
  useEffect(() => { shadowRef.current = { grid: sink.grid, objects: sink.objects } }, [sink.grid, sink.objects])
  const openRef = useRef(false)
  // A pointer release closes the current stroke so the next brush patch starts a new undo step.
  useEffect(() => {
    const close = () => { openRef.current = false }
    window.addEventListener('pointerup', close)
    return () => window.removeEventListener('pointerup', close)
  }, [])
  const rec = useRecorder(sink, push, shadowRef, openRef)
  const { undo, redo } = useReplay(sink, take, shadowRef, openRef)
  return { ...rec, undo, redo, canUndo: stacks.undo.length > 0, canRedo: stacks.redo.length > 0,
    lastKind: stacks.undo.at(-1)?.kind ?? null }
}
