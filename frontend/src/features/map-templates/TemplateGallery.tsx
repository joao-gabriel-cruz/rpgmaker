import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { TemplateCard } from './TemplateCard'
import { TemplatePanel } from './TemplatePanel'
import { ALL_TAGS, fetchPreview, listTemplates, randomSeed, templatesKey, type TemplateInfo } from './templates'

interface Props {
  campaignId: string
  error: string
  onCreate: (name: string, templateId: string, seed: number) => void
}

const chip = (on: boolean) =>
  `rounded-full px-2.5 py-1 text-xs transition-[background-color,transform] active:scale-[0.96] ${on ? 'bg-emerald-600 text-white' : 'bg-zinc-800 text-zinc-300 hover:bg-zinc-700'}`

function usePreview(campaignId: string, id: string, seed: number) {
  return useQuery({ queryKey: [...templatesKey(campaignId), id, seed], queryFn: () => fetchPreview(campaignId, id, seed), staleTime: Infinity }).data
}

function CardWithPreview({ campaignId, info, seed, selected, onSelect }: { campaignId: string; info: TemplateInfo; seed: number; selected: boolean; onSelect: () => void }) {
  return <TemplateCard info={info} preview={usePreview(campaignId, info.id, seed)} selected={selected} onSelect={onSelect} />
}

function SelectedPanel({ campaignId, info, seed, error, onVariation, onCreate }: { campaignId: string; info: TemplateInfo; seed: number; error: string; onVariation: () => void; onCreate: (name: string, id: string, seed: number) => void }) {
  const preview = usePreview(campaignId, info.id, seed)
  return <TemplatePanel key={info.id} info={info} preview={preview} error={error} onVariation={onVariation} onCreate={(name) => onCreate(name, info.id, seed)} />
}

/** Template gallery: tag filter, cards with thumbnails and, for the selected one, a large preview. */
export function TemplateGallery({ campaignId, error, onCreate }: Props) {
  const list = useQuery({ queryKey: templatesKey(campaignId), queryFn: () => listTemplates(campaignId) })
  const [tag, setTag] = useState<string | null>(null)
  const [selected, setSelected] = useState<string | null>(null)
  const [seeds, setSeeds] = useState<Record<string, number>>({})
  const [initialSeed] = useState(randomSeed)
  const seedOf = (id: string) => seeds[id] ?? initialSeed
  if (list.isPending) return <p className="text-sm text-zinc-400">Carregando modelos…</p>
  if (list.isError) return <p role="alert" className="text-sm text-red-400">Não foi possível carregar os modelos.</p>
  const all = list.data
  const tags = ALL_TAGS.filter((t) => all.some((i) => i.tags.includes(t)))
  const shown = tag ? all.filter((i) => i.tags.includes(tag)) : all
  const current = all.find((i) => i.id === selected)
  return (
    <div>
      <div role="group" aria-label="Filtrar por etiqueta" className="mb-3 flex flex-wrap gap-1.5">
        <button type="button" aria-pressed={tag === null} className={chip(tag === null)} onClick={() => setTag(null)}>Todos</button>
        {tags.map((t) => <button key={t} type="button" aria-pressed={tag === t} className={chip(tag === t)} onClick={() => setTag(t)}>{t}</button>)}
      </div>
      <ul className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        {shown.map((i) => <CardWithPreview key={i.id} campaignId={campaignId} info={i} seed={seedOf(i.id)} selected={i.id === selected} onSelect={() => setSelected(i.id)} />)}
      </ul>
      {current && <SelectedPanel campaignId={campaignId} info={current} seed={seedOf(current.id)} error={error}
        onVariation={() => setSeeds((s) => ({ ...s, [current.id]: randomSeed() }))} onCreate={onCreate} />}
    </div>
  )
}
