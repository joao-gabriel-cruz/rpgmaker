import { BrickWall, Crosshair, Droplets, Fuel, Gem, Leaf, PawPrint, Pill, Shirt, Sprout, Trees, Wheat, TriangleAlert, type LucideIcon } from 'lucide-react'
import { RESOURCE_GROUPS, RESOURCE_LABEL, type Refuge, type RefugeEditable, type ResourceKey } from './refuge'
import { Stepper } from './Stepper'

const ICONS: Record<ResourceKey, LucideIcon> = {
  water: Droplets, food: Wheat, medicine: Pill, clothing: Shirt, plants: Leaf, animals: PawPrint,
  wood: Trees, biomass: Sprout, minerals: Gem, fuel: Fuel, ammo: Crosshair, buildingMaterial: BrickWall,
}

interface Props {
  refuge: Refuge
  isGm: boolean
  onChange: (patch: Partial<RefugeEditable>) => void
}

function Bar({ refuge, isGm, onChange }: Props) {
  const { reservesTotal: total, reservesOverflow: over } = refuge.derived
  const pct = refuge.reservesMax > 0 ? Math.min(100, (total / refuge.reservesMax) * 100) : 100
  return (
    <div className="flex flex-col gap-1 rounded-lg bg-zinc-900 p-3">
      <div className="flex items-center justify-between text-sm text-zinc-300">
        <span>Reservas: <span aria-label="Total de Reservas" className="font-semibold">{total}</span> / {refuge.reservesMax}</span>
        <Stepper label="Teto de Reservas" value={refuge.reservesMax} editable={isGm} onChange={(reservesMax) => onChange({ reservesMax })} />
      </div>
      <div role="progressbar" aria-label="Ocupação das Reservas" aria-valuenow={total} aria-valuemax={refuge.reservesMax}
        className="h-3 overflow-hidden rounded-full bg-zinc-700">
        <div className={`h-full ${over > 0 ? 'bg-red-500' : 'bg-emerald-500'}`} style={{ width: `${pct}%` }} />
      </div>
      {over > 0 && (
        <p role="alert" className="flex items-center gap-1 text-sm text-red-400"><TriangleAlert size={14} />
          Excedente de {over}: o que passa do teto se perde na próxima semana.</p>
      )}
    </div>
  )
}

function Group({ title, keys, refuge, isGm, onChange }: Props & { title: string; keys: ResourceKey[] }) {
  return (
    <div className="flex flex-col gap-1">
      <h4 className="text-xs uppercase tracking-wide text-zinc-500">{title}</h4>
      {keys.map((k) => {
        const Icon = ICONS[k]
        return (
          <div key={k} className="flex items-center gap-2 rounded bg-zinc-900 px-3 py-2 text-sm text-zinc-300">
            <Icon size={16} className="text-zinc-400" />{RESOURCE_LABEL[k]}
            <span className="ml-auto">
              <Stepper label={RESOURCE_LABEL[k]} value={refuge.resources[k]} editable={isGm}
                onChange={(v) => onChange({ resources: { ...refuge.resources, [k]: v } })} />
            </span>
          </div>
        )
      })}
    </div>
  )
}

export function ReservesPanel(props: Props) {
  const { refuge, isGm, onChange } = props
  return (
    <section className="flex max-w-md flex-col gap-3">
      <h3 className="text-sm text-zinc-400">Reservas e recursos</h3>
      <Bar {...props} />
      <div className="grid grid-cols-1 gap-3">
        {RESOURCE_GROUPS.map((g) => <Group key={g.title} {...props} title={g.title} keys={g.keys} />)}
      </div>
      <label className="flex items-center gap-2 text-sm text-zinc-300">
        <input type="checkbox" disabled={!isGm} checked={refuge.hasWaterSource} onChange={(e) => onChange({ hasWaterSource: e.target.checked })} />
        <Droplets size={16} className="text-sky-400" />Fonte de Água
      </label>
    </section>
  )
}
