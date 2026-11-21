import { useCharacters } from '../characters/characters'
import type { SceneTokens } from './useSceneTokens'

interface Props {
  campaignId: string
  tokens: SceneTokens
  userId: string
}

function Toast({ message, onClose }: { message: string; onClose: () => void }) {
  return (
    <div role="alert" className="flex items-center justify-between gap-2 rounded border border-red-900 bg-red-950 px-3 py-1 text-sm text-red-200">
      <span>{message}</span>
      <button type="button" aria-label="Fechar aviso" className="text-red-300" onClick={onClose}>×</button>
    </div>
  )
}

const btn = 'rounded-md bg-zinc-800 px-2 py-1 text-sm transition-[background-color,transform] hover:bg-zinc-700 active:scale-[0.96] disabled:opacity-50 motion-reduce:transition-none'

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

/** Players: place their own token (the Mestre's roster lives in table/options/TokensOptions). */
export function TokensPanel({ campaignId, tokens, userId }: Props) {
  return <PlayerControls campaignId={campaignId} tokens={tokens} userId={userId} />
}
