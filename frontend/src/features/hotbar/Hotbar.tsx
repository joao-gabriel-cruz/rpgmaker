import { Dices, FileText, MapPin, RotateCcw } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import type { CharacterView } from '../characters/characters'
import { useRealtime } from '../realtime/realtimeContext'
import { ConditionIcons } from '../ambience/ConditionIcons'
import { conditionsOf } from '../ambience/conditions'
import { RollComposer, type RollPayload } from '../rolls/RollComposer'
import { APTITUDE_LABELS, INSTINCT_LABELS } from '../sheet-assimilacao/labels'
import { SURFACE } from '../table/surface'
import { Kbd } from '../table/Tip'
import { HealthBar, Portrait, StateBadges, TugBar } from './HotbarStatus'
import { useSceneTokensBridge, useTableCharacter } from './TableCharacterContext'
import { useActiveCharacter } from './useActiveCharacter'
import { vitalsOf } from './vitals'

const ACTION = 'flex h-10 items-center gap-1.5 rounded-lg px-3 text-sm font-medium transition-[background-color,color,transform] active:scale-[0.96] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-emerald-400'
const lastText = (p: RollPayload) => `${INSTINCT_LABELS[p.instinct]}${p.aptitude ? ` + ${APTITUDE_LABELS[p.aptitude]}` : ''}`

function Selector({ options, value, onChange }: { options: CharacterView[]; value: string; onChange: (id: string) => void }) {
  return (
    <select aria-label="Personagem ativo" value={value} onChange={(e) => onChange(e.target.value)}
      className="max-w-40 truncate rounded-md bg-zinc-800 px-1.5 py-0.5 text-sm font-semibold focus-visible:outline-2 focus-visible:outline-emerald-400">
      {!value && <option value="">Escolher personagem…</option>}
      {options.map((c) => <option key={c.id} value={c.id}>{c.name || 'Sem nome'}</option>)}
    </select>
  )
}

// Roll / roll again / open sheet / place token; the last roll is remembered per character.
function useHotbarActions(campaignId: string, active: CharacterView | undefined) {
  const { send } = useRealtime()
  const tc = useTableCharacter()
  const [last, setLast] = useState<Record<string, RollPayload>>({})
  const [error, setError] = useState('')
  const prev = active ? last[active.id] : undefined
  const again = useCallback(() => {
    if (!prev) return
    setError('')
    send('roll.request', { ...prev, useOccupation: false, useMarkingEvent: false }).catch((e: Error) => setError(e.message))
  }, [prev, send])
  const { registerRollAgain } = tc
  useEffect(() => {
    registerRollAgain(again)
    return () => registerRollAgain(null)
  }, [registerRollAgain, again])
  const remember = (p: RollPayload) => setLast((l) => ({ ...l, [p.characterId]: p }))
  return { prev, again, remember, error, campaignId }
}

function Buttons({ active, canRoll, a, canPlace }: { active: CharacterView; canRoll: boolean; a: ReturnType<typeof useHotbarActions>; canPlace: boolean }) {
  const tc = useTableCharacter()
  const { tokens } = useSceneTokensBridge()
  const sheetOpen = tc.sheetId === active.id
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {canRoll && (
        <button type="button" aria-haspopup="dialog" aria-expanded={tc.composerOpen} onClick={() => tc.setComposerOpen(!tc.composerOpen)}
          className={`${ACTION} bg-emerald-600 text-white hover:bg-emerald-500`}><Dices size={18} aria-hidden /> Rolar<Kbd>Espaço</Kbd></button>
      )}
      {canRoll && a.prev && (
        <button type="button" onClick={a.again} title={`Rolar de novo: ${lastText(a.prev)}`}
          className={`${ACTION} bg-zinc-800 text-zinc-100 hover:bg-zinc-700`}><RotateCcw size={16} aria-hidden /> Rolar de novo<Kbd>Enter</Kbd></button>
      )}
      <button type="button" aria-pressed={sheetOpen} onClick={() => (sheetOpen ? tc.closeSheet() : tc.openSheet(active.id))}
        className={`${ACTION} ${sheetOpen ? 'bg-zinc-700 text-white' : 'bg-zinc-800 text-zinc-100 hover:bg-zinc-700'}`}><FileText size={16} aria-hidden /> Ficha</button>
      {canPlace && tokens && (
        <button type="button" onClick={() => tokens.place(active.id)} className={`${ACTION} bg-zinc-800 text-zinc-100 hover:bg-zinc-700`}>
          <MapPin size={16} aria-hidden /> {active.ownerId ? 'Colocar meu token' : 'Colocar na cena'}</button>
      )}
    </div>
  )
}

function Identity({ options, active, isGm }: { options: CharacterView[]; active?: CharacterView; isGm: boolean }) {
  const tc = useTableCharacter()
  const v = active && vitalsOf(active)
  return (
    <>
      {active && <Portrait character={active} />}
      <div className="flex min-w-0 flex-col gap-1">
        {options.length > 1 || isGm
          ? <Selector options={options} value={active?.id ?? ''} onChange={tc.select} />
          : <span className="truncate text-sm font-semibold">{active?.name}</span>}
        {v && <span className="flex flex-wrap items-center gap-1"><StateBadges v={v} />{active && <ConditionIcons ids={conditionsOf(active.sheet)} status={active.sheet} />}</span>}
      </div>
      {v && <HealthBar v={v} />}
      {v && <TugBar v={v} />}
    </>
  )
}

/** Bottom-center hotbar: active character, vitals, roll composer, sheet drawer and token placement. */
export function Hotbar({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const { options, active, userId } = useActiveCharacter(campaignId, isGm)
  const tc = useTableCharacter()
  const { tokens } = useSceneTokensBridge()
  const a = useHotbarActions(campaignId, active)
  if (options.length === 0 || (!active && !isGm)) return null
  const canRoll = !!active && active.kind !== 'npc'
  const canPlace = !!active && !!tokens && !tokens.tokens.some((t) => t.characterId === active.id) && (isGm || active.ownerId === userId)
  return (
    <section aria-label="Hotbar do personagem" className={`pointer-events-auto relative max-w-[calc(100%-1.5rem)] rounded-2xl p-1 ${SURFACE}`}>
      {tc.composerOpen && canRoll && (
        <div className="absolute bottom-full left-1/2 mb-2 -translate-x-1/2">
          <RollComposer key={active.id} campaignId={campaignId} character={active} isGm={isGm} onRoll={a.remember}
            preset={a.prev && { instinct: a.prev.instinct, aptitude: a.prev.aptitude }} onClose={() => tc.setComposerOpen(false)} />
        </div>
      )}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-2 rounded-xl px-3 py-2">
        <Identity options={options} active={active} isGm={isGm} />
        {active && <Buttons active={active} canRoll={canRoll} a={a} canPlace={canPlace} />}
      </div>
      {a.error && <p role="alert" className="px-3 pb-1 text-xs text-red-300">{a.error}</p>}
    </section>
  )
}
