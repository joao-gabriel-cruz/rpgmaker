import { useActiveCharacter } from '../hotbar/useActiveCharacter'
import { useRealtime } from '../realtime/realtimeContext'
import { Hint, Section, Segmented } from '../table/options/ui'
import { TIMES, WEATHERS } from './ambience'
import { autoConditions, CONDITIONS, conditionsOf, type ConditionId } from './conditions'
import type { SceneAmbience } from './useSceneAmbience'

function ConditionToggles({ campaignId }: { campaignId: string }) {
  const { send } = useRealtime()
  const { active } = useActiveCharacter(campaignId, true)
  if (!active) return <Hint>Selecione um personagem na hotbar ou no Grupo para aplicar condições.</Hint>
  const status = { conditions: active.sheet.conditions, health: active.sheet.health, tugOfWar: active.sheet.tugOfWar, instincts: active.sheet.instincts }
  const on = conditionsOf(status)
  const auto = autoConditions(status)
  const toggle = (id: ConditionId) => send('condition.set', { characterId: active.id, condition: id, active: !on.includes(id) }).catch(() => undefined)
  return (
    <>
      <p className="truncate text-sm text-zinc-100">{active.name}</p>
      <div className="grid grid-cols-2 gap-1">
        {CONDITIONS.map((c) => {
          const Icon = c.icon
          const isOn = on.includes(c.id)
          const locked = auto.includes(c.id) && !status.conditions?.includes(c.id)
          return (
            <button key={c.id} type="button" aria-pressed={isOn} disabled={locked} title={locked ? 'Automático pela ficha' : undefined} onClick={() => toggle(c.id)}
              className={`flex items-center gap-1.5 rounded-md px-2 py-1.5 text-xs transition-colors disabled:cursor-default motion-reduce:transition-none focus-visible:outline-2 focus-visible:outline-emerald-400 ${isOn ? 'bg-zinc-700 text-white ring-1 ring-white/20' : 'bg-zinc-800 text-zinc-300 hover:bg-zinc-700'}`}>
              <Icon size={14} aria-hidden className={c.tone} /> {c.label}
            </button>
          )
        })}
      </div>
    </>
  )
}

/** Ambiente tool (Mestre): time of day and weather of the scene, and conditions of the selected character. */
export function AmbienceOptions({ campaignId, scene }: { campaignId: string; scene: SceneAmbience }) {
  const { ambience, set, error } = scene
  return (
    <div className="flex flex-col gap-4">
      <Section title="Hora do dia">
        <Segmented label="Hora do dia" value={ambience.timeOfDay} options={TIMES.map((t) => ({ id: t.id, label: t.label, icon: t.icon }))} onChange={(timeOfDay) => set({ timeOfDay })} />
      </Section>
      <Section title="Clima">
        <Segmented label="Clima" value={ambience.weather} options={WEATHERS.map((w) => ({ id: w.id, label: w.label, icon: w.icon }))} onChange={(weather) => set({ weather })} />
      </Section>
      {error && <p role="alert" className="text-xs text-red-300">{error}</p>}
      <Hint>À noite, personagens com lanterna ou tocha e fogueiras iluminam ao redor. Ao editar terreno o mapa fica claro.</Hint>
      <Section title="Condições do personagem"><ConditionToggles campaignId={campaignId} /></Section>
    </div>
  )
}
