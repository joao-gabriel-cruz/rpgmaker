import { xpAvailable, xpSpent } from './derive'
import { SUIT_LABELS } from './labels'
import type { Characteristic, SheetAssimilation, Suit } from './types'
import { btnCls, inputCls, IssueList, Section, Stepper, type SectionProps } from './ui'

export function CharacteristicsSection({ sheet, onChange, readOnly, issues }: SectionProps) {
  const list = sheet.characteristics
  const set = (next: Characteristic[]) => onChange({ ...sheet, characteristics: next })
  const patch = (i: number, p: Partial<Characteristic>) => set(list.map((c, j) => (j === i ? { ...c, ...p } : c)))
  return (
    <Section
      title="Características"
      aside={<span className="text-xs text-zinc-400">XP gasto {xpSpent(sheet)} · disponível {xpAvailable(sheet)} / {sheet.xp.total}</span>}
    >
      <ul className="space-y-2">
        {list.map((c, i) => (
          <li key={i} className="rounded border border-zinc-700 bg-zinc-950 p-2">
            <div className="flex items-center gap-2">
              <input className={`${inputCls} flex-1`} aria-label="Nome da característica" placeholder="Nome" value={c.name} disabled={readOnly} onChange={(e) => patch(i, { name: e.target.value })} />
              <Stepper label={`nível de ${c.name || 'característica'}`} value={c.level} min={1} max={9} disabled={readOnly} onChange={(v) => patch(i, { level: v })} />
              <span className="text-xs text-zinc-500">custo {2 * c.level}</span>
              <button className={btnCls} disabled={readOnly} onClick={() => set(list.filter((_, j) => j !== i))}>Remover</button>
            </div>
            <textarea className={`${inputCls} mt-2 w-full`} aria-label="Descrição da característica" placeholder="Descrição" value={c.description} disabled={readOnly} onChange={(e) => patch(i, { description: e.target.value })} />
          </li>
        ))}
      </ul>
      <button className={`${btnCls} mt-2`} disabled={readOnly} onClick={() => set([...list, { name: '', level: 1, description: '' }])}>+ Característica</button>
      <IssueList issues={issues} field="characteristics" />
    </Section>
  )
}

export function AssimilationsSection({ sheet, onChange, readOnly }: SectionProps) {
  const list = sheet.assimilations
  const set = (next: SheetAssimilation[]) => onChange({ ...sheet, assimilations: next })
  const patch = (i: number, p: Partial<SheetAssimilation>) => set(list.map((a, j) => (j === i ? { ...a, ...p } : a)))
  return (
    <Section title="Assimilações">
      <ul className="space-y-2">
        {list.map((a, i) => (
          <li key={i} className="rounded border border-zinc-700 bg-zinc-950 p-2">
            <div className="flex gap-2">
              <input className={`${inputCls} flex-1`} aria-label="Nome da assimilação" placeholder="Nome" value={a.name} disabled={readOnly} onChange={(e) => patch(i, { name: e.target.value })} />
              <select className={inputCls} aria-label="Naipe" value={a.suit} disabled={readOnly} onChange={(e) => patch(i, { suit: e.target.value as Suit })}>
                {Object.entries(SUIT_LABELS).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
              </select>
              <button className={btnCls} disabled={readOnly} onClick={() => set(list.filter((_, j) => j !== i))}>Remover</button>
            </div>
            <textarea className={`${inputCls} mt-2 w-full`} aria-label="Descrição da assimilação" placeholder="Descrição" value={a.description} disabled={readOnly} onChange={(e) => patch(i, { description: e.target.value })} />
          </li>
        ))}
      </ul>
      <button className={`${btnCls} mt-2`} disabled={readOnly} onClick={() => set([...list, { name: '', suit: 'hearts', description: '' }])}>+ Assimilação</button>
    </Section>
  )
}

export function NotesSection({ sheet, onChange, readOnly }: SectionProps) {
  return (
    <Section title="Notas">
      <textarea className={`${inputCls} h-32 w-full`} aria-label="Notas" value={sheet.notes} disabled={readOnly} onChange={(e) => onChange({ ...sheet, notes: e.target.value })} />
    </Section>
  )
}
