import { useCharacters } from '../characters/characters'
import { SYMBOL_LABELS, newId, type Activation, type ActivationSymbol, type Threat } from './conflict'
import { Group, NumberField, input, patchAt, smallBtn } from './editorParts'

const POOL0 = { d6: 0, d10: 0, d12: 0 }

function NpcSelect({ campaignId, value, onChange }: { campaignId: string; value?: string; onChange: (v?: string) => void }) {
  const { data } = useCharacters(campaignId)
  const npcs = (data ?? []).filter((c) => c.kind === 'npc')
  return (
    <select aria-label="NPC vinculado" value={value ?? ''} onChange={(e) => onChange(e.target.value || undefined)} className={input}>
      <option value="">Sem NPC</option>
      {npcs.map((n) => <option key={n.id} value={n.id}>{n.name}</option>)}
    </select>
  )
}

export function ThreatsEditor({ campaignId, threats, onChange }: { campaignId: string; threats: Threat[]; onChange: (t: Threat[]) => void }) {
  return (
    <Group title="Ameaças" onAdd={() => onChange([...threats, { id: newId(), name: '', pool: { ...POOL0, d6: 1 } }])}>
      {threats.map((t, i) => {
        const pool = t.pool ?? POOL0
        const set = (p: Partial<Threat>) => onChange(patchAt(threats, i, p))
        return (
          <div key={t.id} className="flex flex-wrap items-end gap-2">
            <input aria-label="Nome da ameaça" placeholder="Nome" value={t.name} onChange={(e) => set({ name: e.target.value })} className={input} />
            <NumberField label="d6" value={pool.d6} onChange={(n) => set({ pool: { ...pool, d6: n } })} />
            <NumberField label="d10" value={pool.d10} onChange={(n) => set({ pool: { ...pool, d10: n } })} />
            <NumberField label="d12" value={pool.d12} onChange={(n) => set({ pool: { ...pool, d12: n } })} />
            <NpcSelect campaignId={campaignId} value={t.npcCharacterId} onChange={(v) => set({ npcCharacterId: v })} />
            <input aria-label="Notas da ameaça" placeholder="Notas" value={t.notes ?? ''} onChange={(e) => set({ notes: e.target.value })} className={input} />
            <button type="button" className={smallBtn} onClick={() => onChange(threats.filter((_, j) => j !== i))}>Remover</button>
          </div>
        )
      })}
    </Group>
  )
}

export function ActivationsEditor({ activations, threats, onChange }: { activations: Activation[]; threats: Threat[]; onChange: (a: Activation[]) => void }) {
  return (
    <Group title="Ativações" onAdd={() => onChange([...activations, { id: newId(), symbol: 'pressure', threshold: 3, text: '', source: 'players' }])}>
      {activations.map((a, i) => {
        const set = (p: Partial<Activation>) => onChange(patchAt(activations, i, p))
        return (
          <div key={a.id} className="flex flex-wrap items-end gap-2">
            <NumberField label="A cada" min={1} value={a.threshold} onChange={(n) => set({ threshold: n })} />
            <select aria-label="Símbolo" value={a.symbol} onChange={(e) => set({ symbol: e.target.value as ActivationSymbol })} className={input}>
              {(Object.keys(SYMBOL_LABELS) as ActivationSymbol[]).map((s) => <option key={s} value={s}>{SYMBOL_LABELS[s]}</option>)}
            </select>
            <select aria-label="Origem" value={a.source} onChange={(e) => set({ source: e.target.value })} className={input}>
              <option value="players">Jogadores</option>
              {threats.map((t) => <option key={t.id} value={`threat:${t.id}`}>{t.name || 'Ameaça'}</option>)}
            </select>
            <input aria-label="Texto da ativação" placeholder="O que acontece" value={a.text} onChange={(e) => set({ text: e.target.value })} className={`${input} flex-1`} />
            <button type="button" className={smallBtn} onClick={() => onChange(activations.filter((_, j) => j !== i))}>Remover</button>
          </div>
        )
      })}
    </Group>
  )
}
