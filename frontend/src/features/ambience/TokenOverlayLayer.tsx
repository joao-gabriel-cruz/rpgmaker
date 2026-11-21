import type { Token } from '../tokens/tokens'
import { ConditionIcons } from './ConditionIcons'
import { conditionsOf } from './conditions'
import type { OverlayRegistry } from './overlayRegistry'
import { usePresenceState } from './PresenceContext'

function Badges({ token }: { token: Token }) {
  const presence = usePresenceState()
  const ids = conditionsOf(token.status)
  const bubble = presence?.bubbles.find((b) => b.characterId === token.characterId)
  const emojis = presence?.reactions.filter((r) => r.characterId === token.characterId) ?? []
  return (
    <div className="flex -translate-x-1/2 -translate-y-full flex-col items-center gap-1">
      {bubble && (
        <p role="status" key={bubble.id} className="bubble-in relative max-w-56 rounded-xl bg-zinc-50 px-2.5 py-1.5 text-center text-xs font-medium leading-snug text-zinc-900 shadow-lg shadow-black/40">
          <span className="sr-only">{token.name} diz: </span>{bubble.text}
          <span aria-hidden className="absolute left-1/2 top-full h-2 w-2 -translate-x-1/2 -translate-y-1 rotate-45 bg-zinc-50" />
        </p>
      )}
      <div className="relative flex h-0 w-0 items-end justify-center">
        {emojis.map((r) => <span key={r.id} aria-hidden className="emoji-rise absolute bottom-0 text-2xl drop-shadow">{r.emoji}</span>)}
      </div>
      <ConditionIcons ids={ids} status={token.status} size={14} />
    </div>
  )
}

/** Speech bubbles, rising emojis and condition icons above each token (positioned by OverlayProjector). */
export function TokenOverlayLayer({ tokens, registry }: { tokens: readonly Token[]; registry: OverlayRegistry }) {
  return (
    <div className="pointer-events-none absolute inset-0 z-[5] overflow-hidden">
      {tokens.map((t) => (
        <div key={t.id} ref={(el) => { if (el) registry.set(t.id, el); else registry.delete(t.id) }} className="absolute left-0 top-0 invisible will-change-transform">
          <Badges token={t} />
        </div>
      ))}
    </div>
  )
}
