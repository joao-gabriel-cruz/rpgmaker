import { useState } from 'react'
import { useCharacters } from '../characters/characters'
import type { SceneTokens } from './useSceneTokens'

interface Props {
  campaignId: string
  tokens: SceneTokens
  userId: string
  isGm: boolean
}

function Toast({ message, onClose }: { message: string; onClose: () => void }) {
  return (
    <div role="alert" className="flex items-center justify-between gap-2 rounded border border-red-900 bg-red-950 px-3 py-1 text-sm text-red-200">
      <span>{message}</span>
      <button type="button" aria-label="Fechar aviso" className="text-red-300" onClick={onClose}>×</button>
    </div>
  )
}

const btn = 'rounded bg-zinc-800 px-2 py-1 text-sm hover:bg-zinc-700 disabled:opacity-50'

function GmControls({ campaignId, tokens }: { campaignId: string; tokens: SceneTokens }) {
  const chars = useCharacters(campaignId)
  const [picked, setPicked] = useState('')
  const onScene = new Set(tokens.tokens.map((t) => t.characterId))
  const free = (Array.isArray(chars.data) ? chars.data : []).filter((c) => !onScene.has(c.id))
  const chosen = free.find((c) => c.id === picked)?.id ?? free[0]?.id ?? ''
  return (
    <div className="flex flex-wrap items-center gap-2">
      <select aria-label="Personagem" className="rounded bg-zinc-900 px-2 py-1 text-sm" value={chosen} onChange={(e) => setPicked(e.target.value)}>
        {free.length === 0 && <option value="">Todos os personagens já estão na cena</option>}
        {free.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
      </select>
      <button type="button" className={btn} disabled={!chosen} onClick={() => tokens.place(chosen)}>Colocar na cena</button>
      <button type="button" className={btn} onClick={tokens.focusAll}>Focar todos aqui</button>
      {tokens.tokens.map((t) => (
        <span key={t.id} className="flex items-center gap-1 rounded bg-zinc-900 px-2 py-1 text-xs">
          {t.name}
          <button type="button" aria-label={`Remover token de ${t.name}`} className="text-zinc-400 hover:text-red-400" onClick={() => tokens.remove(t.id)}>×</button>
        </span>
      ))}
    </div>
  )
}

function PlayerControls({ campaignId, tokens, userId }: { campaignId: string; tokens: SceneTokens; userId: string }) {
  const chars = useCharacters(campaignId)
  const onScene = new Set(tokens.tokens.map((t) => t.characterId))
  const mine = (Array.isArray(chars.data) ? chars.data : []).filter((c) => c.ownerId === userId && !onScene.has(c.id))
  if (mine.length === 0) return null
  return (
    <div className="flex flex-wrap items-center gap-2">
      {mine.map((c) => (
        <button key={c.id} type="button" className={btn} onClick={() => tokens.place(c.id)}>
          Colocar meu token{mine.length > 1 ? ` (${c.name})` : ''}
        </button>
      ))}
    </div>
  )
}

/** Error toast for failed token commands; shown over the map whatever the active tool. */
export function TokensError({ tokens }: { tokens: SceneTokens }) {
  return tokens.error ? <Toast message={tokens.error} onClose={tokens.clearError} /> : null
}

/** Mestre: placing tokens and focusing everyone. Players: placing their own token. */
export function TokensPanel({ campaignId, tokens, userId, isGm }: Props) {
  if (!isGm) return <PlayerControls campaignId={campaignId} tokens={tokens} userId={userId} />
  return (
    <div className="flex flex-col gap-2">
      <GmControls campaignId={campaignId} tokens={tokens} />
      <p className="text-xs text-zinc-500">Alt+clique no mapa (ou a ferramenta Ping) chama a atenção de todos.</p>
    </div>
  )
}
