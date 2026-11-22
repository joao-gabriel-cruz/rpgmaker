import { Footprints, Heart, Shield, Swords, Users } from 'lucide-react'
import { LevelTrack } from './LevelTrack'
import { LEVEL_LABELS, LEVEL_TITLES, type LevelKey, type Refuge, type RefugeEditable } from './refuge'
import { Stepper } from './Stepper'

interface Props {
  refuge: Refuge
  isGm: boolean
  onChange: (patch: Partial<RefugeEditable>) => void
}

const ICONS: Record<LevelKey, React.ReactNode> = {
  defense: <Shield size={16} />, morale: <Heart size={16} />, belligerence: <Swords size={16} />,
}
const KEYS: LevelKey[] = ['defense', 'morale', 'belligerence']

function Card({ icon, title, children }: { icon: React.ReactNode; title: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-2 rounded-lg bg-zinc-900 p-3">
      <div className="flex items-center gap-2 text-sm text-zinc-300">{icon}<span>{title}</span></div>
      {children}
    </div>
  )
}

function Population({ refuge, isGm, onChange }: Props) {
  const p = refuge.population
  const set = (patch: Partial<typeof p>) => onChange({ population: { ...p, ...patch } })
  return (
    <Card icon={<Users size={16} />} title="População">
      <div className="flex items-center justify-between text-sm">Nível
        <Stepper label="Nível de População" value={p.level} max={100} editable={isGm} onChange={(level) => set({ level })} /></div>
      <div className="flex items-center justify-between text-sm">Teto
        <Stepper label="Teto de População" value={p.max} min={1} max={100} editable={isGm} onChange={(max) => set({ max })} /></div>
      <p className="text-xs text-zinc-400">
        ~<span aria-label="Habitantes">{refuge.derived.inhabitants}</span> habitantes ·{' '}
        <span aria-label="Trabalhadores">{refuge.derived.workers}</span> trabalhadores
      </p>
      {refuge.derived.overcrowded && (
        <p role="alert" className="text-xs font-medium text-amber-400">Superlotação: a População ({p.level}) passa do teto ({p.max}).</p>
      )}
    </Card>
  )
}

export function RefugeStats({ refuge, isGm, onChange }: Props) {
  return (
    <div className="grid max-w-md grid-cols-1 gap-3">
      {KEYS.map((k) => (
        <LevelTrack key={k} title={LEVEL_TITLES[k]} icon={ICONS[k]} value={refuge[k]} labels={LEVEL_LABELS[k]} isGm={isGm}
          onChange={(v) => onChange({ [k]: v })} />
      ))}
      <Card icon={<Footprints size={16} />} title="Mobilidade">
        <div className="flex items-center justify-between text-sm text-zinc-400">Transporte de População
          <Stepper label="Mobilidade" value={refuge.mobility} max={99} editable={isGm} onChange={(mobility) => onChange({ mobility })} /></div>
      </Card>
      <Population refuge={refuge} isGm={isGm} onChange={onChange} />
    </div>
  )
}
