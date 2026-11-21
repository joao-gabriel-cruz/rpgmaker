import { APTITUDE_LABELS, INSTINCT_LABELS } from '../sheet-assimilacao/labels'
import { INSTINCT_KEYS, KNOWLEDGE_KEYS, PRACTICE_KEYS, type SheetIssue } from '../sheet-assimilacao/types'
import { Field, IssueList, Section, Stepper, btnCls, inputCls } from '../sheet-assimilacao/ui'
import type { NpcSheetData } from './npc'

interface Props {
  sheet: NpcSheetData
  onChange: (next: NpcSheetData) => void
  readOnly: boolean
  issues: SheetIssue[]
}

const MAX_VALUE = 99

function Stats<K extends 'instincts' | 'knowledge' | 'practices'>({ title, group, keys, labels, p }: {
  title: string
  group: K
  keys: readonly (keyof NpcSheetData[K] & string)[]
  labels: Record<string, string>
  p: Props
}) {
  return (
    <Section title={title}>
      <ul className="grid gap-1 sm:grid-cols-2">
        {keys.map((k) => (
          <li key={k} className="flex items-center justify-between gap-2">
            <span className="text-sm">{labels[k]}</span>
            <Stepper label={labels[k]} value={p.sheet[group][k] as number} min={0} max={MAX_VALUE} disabled={p.readOnly}
              onChange={(v) => p.onChange({ ...p.sheet, [group]: { ...p.sheet[group], [k]: v } })} />
          </li>
        ))}
      </ul>
      <IssueList issues={p.issues} field={group} />
    </Section>
  )
}

function Health({ p }: { p: Props }) {
  const set = (patch: Partial<NpcSheetData['health']>) => p.onChange({ ...p.sheet, health: { ...p.sheet.health, ...patch } })
  return (
    <Section title="Saúde">
      <div className="flex flex-wrap gap-6">
        <label className="flex items-center gap-2 text-sm">Caixas por nível
          <Stepper label="Caixas por nível" value={p.sheet.health.boxesPerLevel} min={1} max={MAX_VALUE} disabled={p.readOnly}
            onChange={(v) => set({ boxesPerLevel: v })} />
        </label>
        <label className="flex items-center gap-2 text-sm">Dano
          <Stepper label="Dano" value={p.sheet.health.damage} min={0} max={MAX_VALUE * 6} disabled={p.readOnly}
            onChange={(v) => set({ damage: v })} />
        </label>
      </div>
      <IssueList issues={p.issues} field="health" />
    </Section>
  )
}

function Threats({ p }: { p: Props }) {
  const set = (threats: string[]) => p.onChange({ ...p.sheet, threats })
  return (
    <Section title="Ameaças e habilidades"
      aside={<button className={btnCls} disabled={p.readOnly} onClick={() => set([...p.sheet.threats, ''])}>Adicionar ameaça</button>}>
      <ul className="flex flex-col gap-2">
        {p.sheet.threats.map((t, i) => (
          <li key={i} className="flex gap-2">
            <textarea aria-label={`Ameaça ${i + 1}`} className={`${inputCls} flex-1`} rows={2} value={t} disabled={p.readOnly}
              onChange={(e) => set(p.sheet.threats.map((x, j) => (j === i ? e.target.value : x)))} />
            <button className={btnCls} aria-label={`Remover ameaça ${i + 1}`} disabled={p.readOnly}
              onClick={() => set(p.sheet.threats.filter((_, j) => j !== i))}>Remover</button>
          </li>
        ))}
        {p.sheet.threats.length === 0 && <li className="text-sm text-zinc-400">Nenhuma ameaça.</li>}
      </ul>
      <IssueList issues={p.issues} field="threats" />
    </Section>
  )
}

function Identity({ p }: { p: Props }) {
  const id = p.sheet.identity
  const set = (patch: Partial<NpcSheetData['identity']>) => p.onChange({ ...p.sheet, identity: { ...id, ...patch } })
  return (
    <Section title="NPC">
      <div className="grid gap-3 sm:grid-cols-2">
        <Field label="Nome">
          <input className={inputCls} value={id.name} disabled={p.readOnly} onChange={(e) => set({ name: e.target.value })} />
        </Field>
        <Field label="Retrato (URL)">
          <input className={inputCls} value={id.portraitUrl ?? ''} disabled={p.readOnly}
            onChange={(e) => set({ portraitUrl: e.target.value || null })} />
        </Field>
        <Field label="Descrição e notas">
          <textarea className={`${inputCls} sm:col-span-2`} rows={4} value={p.sheet.notes} disabled={p.readOnly}
            onChange={(e) => p.onChange({ ...p.sheet, notes: e.target.value })} />
        </Field>
      </div>
      <IssueList issues={p.issues} field="identity" />
      <IssueList issues={p.issues} field="notes" />
    </Section>
  )
}

/** Compact NPC sheet: identity, instincts, aptitudes, health boxes and a threat list. */
export function NpcSheet(p: Props) {
  return (
    <div className="flex flex-col gap-4">
      <Identity p={p} />
      <Stats title="Instintos" group="instincts" keys={INSTINCT_KEYS} labels={INSTINCT_LABELS} p={p} />
      <Stats title="Conhecimentos" group="knowledge" keys={KNOWLEDGE_KEYS} labels={APTITUDE_LABELS} p={p} />
      <Stats title="Práticas" group="practices" keys={PRACTICE_KEYS} labels={APTITUDE_LABELS} p={p} />
      <Health p={p} />
      <Threats p={p} />
    </div>
  )
}
