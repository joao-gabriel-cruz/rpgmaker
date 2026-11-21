import { AmbienceBadge } from './AmbienceBadge'
import { MapThumbnail } from './MapThumbnail'
import type { TemplateInfo, TemplatePreview } from './templates'

interface Props {
  info: TemplateInfo
  preview: TemplatePreview | undefined
  selected: boolean
  onSelect: () => void
}

/** Gallery card: thumbnail, name, short hook, size, tags and the suggested ambience. */
export function TemplateCard({ info, preview, selected, onSelect }: Props) {
  return (
    <li>
      <button type="button" onClick={onSelect} aria-pressed={selected} aria-label={`Modelo ${info.name}`}
        className={`flex h-full w-full flex-col gap-2 rounded-xl bg-zinc-800/60 p-2 text-left ring-1 transition-[background-color,transform,box-shadow] hover:bg-zinc-800 active:scale-[0.98] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-400 ${selected ? 'ring-2 ring-emerald-400' : 'ring-white/10'}`}>
        <MapThumbnail preview={preview} scale={2} label={`Miniatura de ${info.name}`} className="aspect-[3/2] object-cover" />
        <span className="flex items-center justify-between gap-2">
          <span className="text-sm font-semibold">{info.name}</span>
          <AmbienceBadge ambience={info.ambience} />
        </span>
        <span className="line-clamp-3 text-xs text-zinc-400">{info.description}</span>
        <span className="mt-auto flex flex-wrap items-center gap-1 text-[11px]">
          <span className="rounded bg-zinc-700 px-1.5 py-0.5 text-zinc-200">{info.width}×{info.depth}</span>
          {info.tags.map((t) => <span key={t} className="rounded bg-zinc-900 px-1.5 py-0.5 text-zinc-400">{t}</span>)}
        </span>
      </button>
    </li>
  )
}
