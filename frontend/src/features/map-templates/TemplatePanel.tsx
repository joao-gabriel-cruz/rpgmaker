import { Dices } from 'lucide-react'
import { useEffect, useRef, useState, type FormEvent } from 'react'
import { AmbienceBadge } from './AmbienceBadge'
import { MapThumbnail } from './MapThumbnail'
import type { TemplateInfo, TemplatePreview } from './templates'

interface Props {
  info: TemplateInfo
  preview: TemplatePreview | undefined
  error: string
  onVariation: () => void
  onCreate: (name: string) => void
}

const btn = 'rounded-lg px-3 py-1.5 text-sm font-medium transition-[background-color,transform] active:scale-[0.96] disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-400'

/** Larger preview of the selected template with the variation button, scene name and create button. */
export function TemplatePanel({ info, preview, error, onVariation, onCreate }: Props) {
  const [name, setName] = useState(info.name)
  const box = useRef<HTMLElement>(null)
  useEffect(() => {
    box.current?.scrollIntoView?.({ block: 'nearest' })
  }, [])
  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (name.trim()) onCreate(name.trim())
  }
  return (
    <section ref={box} aria-label={`Prévia de ${info.name}`} className="mt-4 grid gap-3 rounded-xl bg-zinc-950/60 p-3 ring-1 ring-white/10 sm:grid-cols-[minmax(0,1fr)_16rem]">
      <MapThumbnail preview={preview} scale={4} label={`Prévia de ${info.name}`} className="max-h-64 object-contain" />
      <form onSubmit={submit} className="flex flex-col gap-2">
        <h3 className="flex items-center justify-between text-base font-semibold">{info.name}<AmbienceBadge ambience={info.ambience} /></h3>
        <p className="text-xs text-zinc-400">{info.description}</p>
        <button type="button" onClick={onVariation} className={`${btn} flex items-center justify-center gap-1.5 bg-zinc-800 hover:bg-zinc-700`}>
          <Dices size={16} aria-hidden />Outra variação
        </button>
        <label className="mt-auto text-xs text-zinc-400">Nome da cena
          <input value={name} maxLength={80} onChange={(e) => setName(e.target.value)}
            className="mt-1 block w-full rounded-lg border border-zinc-700 bg-zinc-900 px-2 py-1.5 text-sm text-zinc-100" />
        </label>
        {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
        <button type="submit" disabled={!name.trim() || !preview} className={`${btn} bg-emerald-600 text-white hover:bg-emerald-500`}>Criar cena</button>
      </form>
    </section>
  )
}
