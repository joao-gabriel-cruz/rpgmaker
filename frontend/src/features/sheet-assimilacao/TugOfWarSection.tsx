import { Section, Stepper, IssueList, type SectionProps } from './ui'

function Rope({ determination }: { determination: number }) {
  return (
    <div className="flex h-6 gap-0.5" role="img" aria-label={`Cabo de guerra: Determinação ${determination}, Assimilação ${10 - determination}`}>
      {Array.from({ length: 10 }, (_, i) => (
        <div key={i} className={`flex-1 rounded-sm ${i < determination ? 'bg-emerald-500' : 'bg-fuchsia-600'}`} />
      ))}
    </div>
  )
}

interface SideProps {
  title: string
  color: string
  level: number
  points: number
  disabled: boolean
  onLevel: (v: number) => void
  onPoints: (v: number) => void
}

function Side({ title, color, level, points, disabled, onLevel, onPoints }: SideProps) {
  return (
    <div className="space-y-2">
      <h3 className={color}>{title}</h3>
      <div className="flex items-center justify-between">Nível
        <Stepper label={`nível de ${title}`} value={level} min={0} max={10} disabled={disabled} onChange={onLevel} />
      </div>
      <div className="flex items-center justify-between">Pontos
        <Stepper label={`pontos de ${title}`} value={points} min={0} max={level} disabled={disabled} onChange={onPoints} />
      </div>
    </div>
  )
}

export function TugOfWarSection({ sheet, onChange, readOnly, issues }: SectionProps) {
  const t = sheet.tugOfWar
  const set = (patch: Partial<typeof t>) => onChange({ ...sheet, tugOfWar: { ...t, ...patch } })
  const setDetermination = (level: number) => {
    const asm = 10 - level
    set({
      determinationLevel: level,
      assimilationLevel: asm,
      determinationPoints: Math.min(t.determinationPoints, level),
      assimilationPoints: Math.min(t.assimilationPoints, asm),
    })
  }
  const badge = t.determinationPoints === 0 ? <span className="rounded bg-red-900 px-2 text-xs text-red-200">Suscetível</span> : null
  return (
    <Section title="Cabo de Guerra" aside={badge}>
      <Rope determination={t.determinationLevel} />
      <div className="mt-3 grid grid-cols-2 gap-4 text-sm">
        <Side title="Determinação" color="text-emerald-400" level={t.determinationLevel} points={t.determinationPoints} disabled={readOnly} onLevel={setDetermination} onPoints={(v) => set({ determinationPoints: v })} />
        <Side title="Assimilação" color="text-fuchsia-400" level={t.assimilationLevel} points={t.assimilationPoints} disabled={readOnly} onLevel={(v) => setDetermination(10 - v)} onPoints={(v) => set({ assimilationPoints: v })} />
      </div>
      <IssueList issues={issues} field="tugOfWar" />
    </Section>
  )
}
