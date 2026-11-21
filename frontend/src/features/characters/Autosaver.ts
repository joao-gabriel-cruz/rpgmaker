import type { Sheet } from '../sheet-assimilacao/types'
import type { CharacterView } from './characters'

export const AUTOSAVE_DELAY_MS = 600

/** Debounce in use; tests widen it so a slow machine cannot fire the save mid-assertion. */
export const autosaveTiming = { delayMs: AUTOSAVE_DELAY_MS }

export interface AutosaverHooks {
  save: (sheet: Sheet) => Promise<CharacterView>
  /** A newer server version should replace the local sheet. */
  onRemote: (c: CharacterView) => void
  /** `unsaved` is true when the user edited again while this save was in flight. */
  onSaved: (c: CharacterView, unsaved: boolean) => void
  onFailed: (e: unknown) => void
  onSaving: () => void
}

const stamp = (c: CharacterView) => Date.parse(c.updatedAt)

/** Debounced saver; remote versions arriving during a local edit wait until it is saved. */
export class Autosaver {
  private dirty = false
  private saving = false
  private latest: Sheet
  private version: number
  private pending: CharacterView | null = null
  private timer: ReturnType<typeof setTimeout> | undefined
  private hooks: AutosaverHooks

  constructor(initial: CharacterView, hooks: AutosaverHooks) {
    this.latest = initial.sheet
    this.version = stamp(initial)
    this.hooks = hooks
  }

  edit(sheet: Sheet) {
    this.latest = sheet
    this.dirty = true
    this.hooks.onSaving()
    this.schedule()
  }

  remote(c: CharacterView) {
    if (stamp(c) <= this.version) return
    if (this.dirty || this.saving) this.pending = c
    else this.adopt(c)
  }

  /** Saves right away if there is an unsaved edit (also used on unmount). */
  async flush() {
    clearTimeout(this.timer)
    this.timer = undefined
    if (this.saving || !this.dirty) return
    this.saving = true
    this.dirty = false
    const failed = await this.attempt()
    this.saving = false
    if (this.dirty && !failed) this.schedule()
    this.applyPending()
  }

  private async attempt() {
    try {
      const saved = await this.hooks.save(this.latest)
      this.version = stamp(saved)
      this.hooks.onSaved(saved, this.dirty)
      return false
    } catch (e) {
      this.dirty = true
      this.hooks.onFailed(e)
      return true
    }
  }

  private schedule() {
    clearTimeout(this.timer)
    this.timer = setTimeout(() => void this.flush(), autosaveTiming.delayMs)
  }

  private applyPending() {
    const p = this.pending
    if (this.dirty || !p) return
    this.pending = null
    if (stamp(p) > this.version) this.adopt(p)
  }

  private adopt(c: CharacterView) {
    this.version = stamp(c)
    this.latest = c.sheet
    this.hooks.onRemote(c)
  }
}
