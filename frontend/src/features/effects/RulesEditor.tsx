import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { EFFECT_LABEL, TRIGGER_LABEL, type EffectKind, type Rule, type RuleSet, type Trigger } from './effects'

const field = 'rounded bg-zinc-800 px-2 py-1 text-sm'
const btn = 'rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700'

function RuleRow({ r, onChange, onRemove, name }: { r: Rule; name: string; onChange: (r: Rule) => void; onRemove: () => void }) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <select aria-label={`Gatilho ${name}`} className={field} value={r.trigger}
        onChange={(e) => onChange({ ...r, trigger: e.target.value as Trigger })}>
        {Object.entries(TRIGGER_LABEL).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
      </select>
      <select aria-label={`Efeito ${name}`} className={field} value={r.effect}
        onChange={(e) => onChange({ ...r, effect: e.target.value as EffectKind, amount: e.target.value === 'notice' ? 0 : Math.max(1, r.amount) })}>
        {Object.entries(EFFECT_LABEL).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
      </select>
      <input aria-label={`Quantidade ${name}`} type="number" min={0} className={`${field} w-16`} value={r.amount}
        disabled={r.effect === 'notice'} onChange={(e) => onChange({ ...r, amount: Number(e.target.value) })} />
      <input aria-label={`Mensagem ${name}`} className={`${field} min-w-40 flex-1`} maxLength={200} value={r.message}
        onChange={(e) => onChange({ ...r, message: e.target.value })} />
      <button className={btn} onClick={onRemove} aria-label={`Remover regra ${name}`}>Remover</button>
    </div>
  )
}

function SetRow({ campaignId, set, onSaved }: { campaignId: string; set: RuleSet; onSaved: (s: RuleSet) => void }) {
  const [rules, setRules] = useState(set.rules)
  const [error, setError] = useState('')
  const base = `/campaigns/${campaignId}/effects/rules/${set.kind}/${set.key}`
  const run = (p: Promise<RuleSet>) => p.then((s) => { setError(''); onSaved(s) }, (e: Error) => setError(e.message))
  return (
    <tr className="border-t border-zinc-800 align-top">
      <td className="py-2 pr-3 text-sm">{set.label}{set.blocksMovement ? ' (bloqueia)' : ''}</td>
      <td className="flex flex-col gap-1 py-2">
        {rules.map((r, i) => (
          <RuleRow key={i} r={r} name={`${set.label} ${i + 1}`}
            onChange={(n) => setRules(rules.map((x, j) => (j === i ? n : x)))}
            onRemove={() => setRules(rules.filter((_, j) => j !== i))} />
        ))}
        <div className="flex gap-2">
          <button className={btn} onClick={() => setRules([...rules, { trigger: 'on_enter', effect: 'notice', amount: 0, message: '' }])}
            aria-label={`Adicionar regra ${set.label}`}>Adicionar</button>
          <button className={btn} onClick={() => run(api.put<RuleSet>(base, { rules }))} aria-label={`Salvar ${set.label}`}>Salvar</button>
          {!set.isDefault && <button className={btn} onClick={() => run(api.delete<RuleSet>(base))} aria-label={`Restaurar padrão ${set.label}`}>Restaurar padrão</button>}
        </div>
        {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      </td>
    </tr>
  )
}

/** Mestre table to edit the effects of each terrain type and prefab; blocking flags come from the catalog. */
export function RulesEditor({ campaignId }: { campaignId: string }) {
  const [sets, setSets] = useState<RuleSet[]>([])
  useEffect(() => {
    api.get<{ rules: RuleSet[] }>(`/campaigns/${campaignId}/effects/rules`).then((r) => setSets(Array.isArray(r?.rules) ? r.rules : []), () => {})
  }, [campaignId])
  const saved = (s: RuleSet) => setSets((cur) => cur.map((c) => (c.kind === s.kind && c.key === s.key ? s : c)))

  return (
    <details className="rounded bg-zinc-900 px-4 py-3">
      <summary className="cursor-pointer text-lg font-semibold">Efeitos de terreno</summary>
      <p className="mt-1 text-sm text-zinc-400">Sugestões para o Mestre; bloqueio de movimento e visão vêm do catálogo e não são editáveis.</p>
      <table className="mt-2 w-full text-left">
        <tbody>{sets.map((s) => <SetRow key={`${s.kind}:${s.key}:${s.isDefault}:${JSON.stringify(s.rules)}`} campaignId={campaignId} set={s} onSaved={saved} />)}</tbody>
      </table>
    </details>
  )
}
