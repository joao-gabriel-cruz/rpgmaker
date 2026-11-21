import { ANIMATION_LABEL, CATEGORY_LABEL, isWeapon, RESOURCE_LABEL, type ActionCategory, type AttackAnimation, type ItemDraft, type ItemEffects, type RefugeResourceKey } from './items'

export const fieldCls = 'rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm text-zinc-100 focus:border-emerald-500 focus:outline-none'

interface Props {
  draft: ItemDraft
  onChange: (next: ItemDraft) => void
}

function KeptBonusField({ effects, set }: { effects: ItemEffects; set: (e: ItemEffects) => void }) {
  const b = effects.keptBonus
  return (
    <fieldset className="flex flex-wrap items-center gap-2 text-sm">
      <label className="flex items-center gap-1">
        <input type="checkbox" checked={!!b} onChange={(e) => set({ ...effects, keptBonus: e.target.checked ? { actionCategory: 'exploration', amount: 1 } : undefined })} />
        Bônus de dado mantido
      </label>
      {b && (
        <>
          <select aria-label="Categoria do bônus" className={fieldCls} value={b.actionCategory}
            onChange={(e) => set({ ...effects, keptBonus: { ...b, actionCategory: e.target.value as ActionCategory } })}>
            {Object.entries(CATEGORY_LABEL).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
          </select>
          <select aria-label="Valor do bônus" className={fieldCls} value={b.amount} onChange={(e) => set({ ...effects, keptBonus: { ...b, amount: Number(e.target.value) } })}>
            <option value={1}>+1</option>
            <option value={2}>+2</option>
          </select>
        </>
      )}
    </fieldset>
  )
}

function ConsumableFields({ effects, set }: { effects: ItemEffects; set: (e: ItemEffects) => void }) {
  const r = effects.refugeResource
  return (
    <div className="flex flex-col gap-2 text-sm">
      <label className="flex items-center gap-2">Cura ao consumir (1–6)
        <input type="number" min={0} max={6} className={`${fieldCls} w-16`} value={effects.heal ?? 0}
          onChange={(e) => set({ ...effects, heal: Number(e.target.value) || undefined })} />
      </label>
      <label className="flex items-center gap-2">Recurso do Refúgio
        <select className={fieldCls} value={r?.resource ?? ''} onChange={(e) => set({ ...effects, refugeResource: e.target.value ? { resource: e.target.value as RefugeResourceKey, amount: r?.amount ?? 1 } : undefined })}>
          <option value="">Nenhum</option>
          {Object.entries(RESOURCE_LABEL).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
        {r && <input type="number" min={1} max={10} aria-label="Quantidade do recurso" className={`${fieldCls} w-16`} value={r.amount}
          onChange={(e) => set({ ...effects, refugeResource: { ...r, amount: Number(e.target.value) || 1 } })} />}
      </label>
    </div>
  )
}

/** Effects block of the item form: kept bonus, heal/refuge resource (consumables) and attack animation (weapons). */
export function EffectsFields({ draft, onChange }: Props) {
  const set = (effects: ItemEffects) => onChange({ ...draft, effects })
  return (
    <div className="flex flex-col gap-2">
      <KeptBonusField effects={draft.effects} set={set} />
      {draft.kind === 'consumable' && <ConsumableFields effects={draft.effects} set={set} />}
      {isWeapon(draft.kind) && (
        <label className="flex items-center gap-2 text-sm">Animação de ataque
          <select className={fieldCls} value={draft.attackAnimation ?? ''} onChange={(e) => onChange({ ...draft, attackAnimation: (e.target.value || undefined) as AttackAnimation | undefined })}>
            <option value="">Nenhuma</option>
            {Object.entries(ANIMATION_LABEL).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
          </select>
        </label>
      )}
    </div>
  )
}
