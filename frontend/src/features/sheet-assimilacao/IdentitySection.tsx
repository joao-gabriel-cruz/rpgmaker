import { GENERATION_LABELS } from './labels'
import type { Generation, Sheet } from './types'
import { Field, IssueList, inputCls, Section, TextList, type SectionProps } from './ui'

type Identity = Sheet['identity']

function Portrait({ url }: { url: string | null }) {
  if (url) return <img src={url} alt="Retrato do personagem" className="h-32 w-32 rounded border border-zinc-700 object-cover" />
  return (
    <div className="flex h-32 w-32 items-center justify-center rounded border border-dashed border-zinc-700 text-xs text-zinc-500">
      Sem retrato
    </div>
  )
}

interface FieldsProps {
  id: Identity
  readOnly: boolean
  set: (patch: Partial<Identity>) => void
}

function BasicFields({ id, readOnly, set }: FieldsProps) {
  return (
    <>
      <Field label="Nome">
        <input className={inputCls} value={id.name} disabled={readOnly} onChange={(e) => set({ name: e.target.value })} />
      </Field>
      <Field label="URL do retrato">
        <input className={inputCls} value={id.portraitUrl ?? ''} disabled={readOnly} onChange={(e) => set({ portraitUrl: e.target.value || null })} />
      </Field>
      <Field label="Geração">
        <select className={inputCls} value={id.generation} disabled={readOnly} onChange={(e) => set({ generation: e.target.value as Generation })}>
          {Object.entries(GENERATION_LABELS).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
        </select>
      </Field>
      <Field label="Ocupação">
        <input className={inputCls} value={id.occupation} disabled={readOnly} onChange={(e) => set({ occupation: e.target.value })} />
      </Field>
      <div className="col-span-2">
        <Field label="Evento Marcante">
          <input className={inputCls} value={id.markingEvent} disabled={readOnly} onChange={(e) => set({ markingEvent: e.target.value })} />
        </Field>
      </div>
    </>
  )
}

export function IdentitySection({ sheet, onChange, readOnly, issues }: SectionProps) {
  const id = sheet.identity
  const set = (patch: Partial<Identity>) => onChange({ ...sheet, identity: { ...id, ...patch } })
  return (
    <Section title="Identidade">
      <div className="grid grid-cols-[8rem_1fr] gap-4">
        <Portrait url={id.portraitUrl} />
        <div className="grid grid-cols-2 gap-3">
          <BasicFields id={id} readOnly={readOnly} set={set} />
          <div className="flex flex-col gap-2">
            <TextList label="Propósito pessoal" values={id.personalPurposes} readOnly={readOnly} onChange={(v) => set({ personalPurposes: v })} />
            <IssueList issues={issues} field="identity.personalPurposes" />
          </div>
          <div className="flex flex-col gap-2">
            <TextList label="Propósito coletivo" values={id.collectivePurposes} readOnly={readOnly} onChange={(v) => set({ collectivePurposes: v })} />
            <IssueList issues={issues} field="identity.collectivePurposes" />
          </div>
        </div>
      </div>
    </Section>
  )
}
