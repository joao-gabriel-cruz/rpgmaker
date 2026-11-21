import { EyeOff, MessageCircle, SmilePlus } from 'lucide-react'
import { useEffect, useRef, useState, type RefObject } from 'react'
import { useActiveCharacter } from '../hotbar/useActiveCharacter'
import { useRealtime } from '../realtime/realtimeContext'
import { SURFACE } from '../table/surface'
import { Kbd } from '../table/Tip'
import { conditionsOf } from './conditions'
import { usePresenceState, type PresenceApi } from './PresenceContext'
import { ReactionWheel } from './ReactionWheel'
import { SpeakBox } from './SpeakBox'

const BTN = 'group flex h-10 items-center gap-1.5 rounded-lg px-3 text-sm font-medium text-zinc-200 transition-[background-color,transform] hover:bg-zinc-800 active:scale-[0.96] disabled:opacity-40 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-400 motion-reduce:transition-none'

function useDockActions(characterId: string | undefined, hidden: boolean) {
  const p = usePresenceState()
  const { send } = useRealtime()
  return {
    say: (text: string) => send('chat.say', { characterId, text }),
    react: (emoji: string, id?: string) => {
      p?.setReactOpen(false)
      p?.closeWheel()
      send('chat.react', { characterId: id ?? characterId, emoji }).catch(() => undefined)
    },
    toggleHidden: () => send('condition.set', { characterId, condition: 'hidden', active: !hidden }).catch(() => undefined),
  }
}

interface ButtonsProps {
  p: PresenceApi
  active: boolean
  isGm: boolean
  hidden: boolean
  btn: RefObject<HTMLButtonElement | null>
  onHide: () => void
}

function DockButtons({ p, active, isGm, hidden, btn, onHide }: ButtonsProps) {
  return (
      <div role="group" aria-label="Presença" className={`flex items-center gap-0.5 rounded-2xl p-1 ${SURFACE}`}>
        <button type="button" disabled={!active} aria-expanded={p.speakOpen} aria-keyshortcuts="'" onClick={() => p.setSpeakOpen(!p.speakOpen)} className={BTN}>
          <MessageCircle size={16} aria-hidden /> Falar<Kbd>'</Kbd>
        </button>
        <button type="button" disabled={!active} aria-haspopup="menu" aria-expanded={p.reactOpen} ref={btn} aria-keyshortcuts="G" onClick={() => p.setReactOpen(!p.reactOpen)} className={BTN}>
          <SmilePlus size={16} aria-hidden /> Reagir<Kbd>G</Kbd>
        </button>
        {!isGm && (
          <button type="button" disabled={!active} aria-pressed={hidden} onClick={onHide} title="Marcar o personagem como Escondido"
            className={`${BTN} ${hidden ? 'bg-zinc-700 text-white' : ''}`}><EyeOff size={16} aria-hidden /> Esconder-me</button>
        )}
      </div>
  )
}

/** Bottom-left dock: speak in character, react with an emoji and (players) hide the character. */
export function PresenceDock({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const p = usePresenceState()
  const { active } = useActiveCharacter(campaignId, isGm)
  const reactBtn = useRef<HTMLButtonElement>(null)
  const [anchor, setAnchor] = useState({ x: 160, y: 520 })
  const reactOpen = p?.reactOpen
  // Center the wheel above the Reagir button, however it was opened (click or the G key).
  useEffect(() => {
    const r = reactBtn.current?.getBoundingClientRect()
    if (reactOpen && r) setAnchor({ x: r.left + r.width / 2, y: r.top - 80 })
  }, [reactOpen])
  const hidden = !!active && conditionsOf({ conditions: active.sheet.conditions }).includes('hidden')
  const act = useDockActions(active?.id, hidden)
  if (!p) return null
  return (
    <div className="pointer-events-auto relative flex flex-col items-start gap-2">
      {p.speakOpen && active && <SpeakBox name={active.name || 'personagem'} onSay={act.say} onClose={() => p.setSpeakOpen(false)} />}
      <DockButtons {...{ p, active: !!active, isGm, hidden, btn: reactBtn, onHide: act.toggleHidden }} />
      {p.reactOpen && active && <ReactionWheel x={anchor.x} y={anchor.y} onPick={(e) => act.react(e)} onClose={() => p.setReactOpen(false)} />}
      {p.wheel && <ReactionWheel x={p.wheel.x} y={p.wheel.y} onPick={(e) => act.react(e, p.wheel?.characterId)} onClose={p.closeWheel} />}
    </div>
  )
}
