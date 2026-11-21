import { Volume2, VolumeX } from 'lucide-react'
import { useState } from 'react'
import { ICON_BTN, SURFACE } from '../table/surface'
import { levelOf } from './engine'
import { useSound } from './SoundContext'

/** Top bar button: mute toggle plus a volume slider popover, per player. */
export function SoundControl() {
  const sound = useSound()
  const [open, setOpen] = useState(false)
  if (!sound) return null
  const { settings, setVolume, setMuted } = sound
  const silent = levelOf(settings) === 0
  const Icon = silent ? VolumeX : Volume2
  return (
    <div className="relative">
      <button type="button" aria-label="Som" aria-expanded={open} aria-haspopup="dialog" title="Volume e mudo" onClick={() => setOpen((o) => !o)} className={`${ICON_BTN} h-9 w-9`}>
        <Icon size={18} strokeWidth={1.5} aria-hidden />
      </button>
      {open && (
        <div role="dialog" aria-label="Controle de som" onKeyDown={(e) => e.key === 'Escape' && setOpen(false)}
          className={`absolute right-0 top-full z-40 mt-2 flex w-56 flex-col gap-2 rounded-xl p-3 ${SURFACE}`}>
          <label className="flex items-center justify-between text-xs font-medium text-zinc-300" htmlFor="sound-volume">
            Volume <span className="tabular-nums text-zinc-400">{Math.round(settings.volume * 100)}%</span>
          </label>
          <input id="sound-volume" type="range" min={0} max={100} step={5} value={Math.round(settings.volume * 100)}
            onChange={(e) => setVolume(Number(e.target.value) / 100)} className="accent-emerald-500" />
          <button type="button" aria-pressed={settings.muted} onClick={() => setMuted(!settings.muted)}
            className="rounded-md bg-zinc-800 px-2 py-1.5 text-xs text-zinc-200 hover:bg-zinc-700 aria-pressed:bg-red-900 aria-pressed:text-red-100">
            {settings.muted ? 'Som desligado — ativar' : 'Desligar o som'}
          </button>
        </div>
      )}
    </div>
  )
}
