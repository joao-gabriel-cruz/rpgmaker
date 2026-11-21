import { Crosshair, LocateFixed, MapPinPlus, Trash2, Wand2 } from 'lucide-react'
import { useCharacters, type CharacterView } from '../../characters/characters'
import type { SceneTokens } from '../../tokens/useSceneTokens'
import { useTableTool } from '../useTableTool'
import { BTN, Hint, Section } from './ui'

const ROW_BTN = 'flex h-7 w-7 items-center justify-center rounded-md text-zinc-300 transition-[background-color,color,transform] hover:bg-zinc-700 hover:text-white active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-emerald-400 motion-reduce:transition-none'

function Avatar({ c }: { c: CharacterView }) {
  const url = c.sheet?.identity?.portraitUrl
  return url
    ? <img src={url} alt="" className="h-8 w-8 shrink-0 rounded-full object-cover ring-1 ring-white/10" />
    : <span aria-hidden className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-zinc-700 text-xs font-medium text-zinc-200 ring-1 ring-white/10">
        {c.name.slice(0, 2).toUpperCase()}
      </span>
}

interface RowProps {
  c: CharacterView
  tokenId: string | undefined
  armed: boolean
  onArm: () => void
  onAuto: () => void
  onRemove: () => void
}

function Row({ c, tokenId, armed, onArm, onAuto, onRemove }: RowProps) {
  return (
    <li className={`flex items-center gap-2 rounded-lg p-1.5 ${armed ? 'bg-emerald-600/20 ring-1 ring-emerald-500' : 'bg-zinc-800/60'}`}>
      <Avatar c={c} />
      <span className="min-w-0 flex-1">
        <span className="block truncate text-sm text-zinc-100">{c.name}</span>
        <span className="block text-[11px] text-zinc-400">{c.kind === 'npc' ? 'NPC' : 'Personagem'}{tokenId ? ' · na cena' : ''}</span>
      </span>
      {tokenId ? (
        <button type="button" aria-label={`Remover token de ${c.name}`} title="Remover da cena" onClick={onRemove} className={`${ROW_BTN} hover:text-red-300`}>
          <Trash2 size={15} strokeWidth={1.5} aria-hidden />
        </button>
      ) : (
        <>
          <button type="button" aria-label={`Colocar ${c.name} na cena`} title="Colocar na cena (clique no mapa)" aria-pressed={armed} onClick={onArm} className={`${ROW_BTN} ${armed ? 'bg-emerald-600 text-white' : ''}`}>
            <MapPinPlus size={15} strokeWidth={armed ? 2.25 : 1.5} aria-hidden />
          </button>
          <button type="button" aria-label={`Colocar ${c.name} automaticamente`} title="Colocar automaticamente" onClick={onAuto} className={ROW_BTN}>
            <Wand2 size={15} strokeWidth={1.5} aria-hidden />
          </button>
        </>
      )}
    </li>
  )
}

/** Tokens tool: roster with avatars; "colocar" arms a click on the map, "automático" picks a free cell. */
export function TokensOptions({ campaignId, tokens }: { campaignId: string; tokens: SceneTokens }) {
  const chars = useCharacters(campaignId)
  const { placing, setPlacing } = useTableTool()
  const list = Array.isArray(chars.data) ? chars.data : []
  const tokenOf = (id: string) => tokens.tokens.find((t) => t.characterId === id)?.id
  const armedName = list.find((c) => c.id === placing)?.name
  return (
    <div className="flex flex-col gap-3">
      <Section title="Personagens e NPCs">
        {list.length === 0 && <Hint>Nenhum personagem na campanha ainda.</Hint>}
        <ul aria-label="Personagens" className="flex max-h-64 flex-col gap-1 overflow-y-auto pr-0.5">
          {list.map((c) => (
            <Row key={c.id} c={c} tokenId={tokenOf(c.id)} armed={placing === c.id}
              onArm={() => setPlacing(placing === c.id ? null : c.id)}
              onAuto={() => tokens.place(c.id)} onRemove={() => tokens.remove(tokenOf(c.id)!)} />
          ))}
        </ul>
        {armedName && <p role="status" className="text-xs text-emerald-300">Clique no mapa para posicionar {armedName}. Esc cancela.</p>}
      </Section>
      <button type="button" onClick={tokens.focusAll} className={`${BTN} flex items-center justify-center gap-1.5 py-1.5`}>
        <Crosshair size={14} strokeWidth={1.5} aria-hidden />Focar todos aqui
      </button>
      <Hint><LocateFixed size={12} aria-hidden className="mr-1 inline" />Alt+clique no mapa (ou a ferramenta Ping) chama a atenção de todos.</Hint>
    </div>
  )
}
