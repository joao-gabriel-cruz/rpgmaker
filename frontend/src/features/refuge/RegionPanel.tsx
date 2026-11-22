import { ChevronDown, ChevronRight, Mountain } from 'lucide-react'
import { useState } from 'react'
import { LevelTrack } from './LevelTrack'
import {
  ABUNDANCE_LABEL, NATURAL_KEYS, REGION_LABEL, REGION_SCALE, RESOURCE_LABEL,
  type Abundance, type NaturalKey, type Refuge, type Region, type RefugeEditable, type RegionKey,
} from './refuge'

interface Props {
  refuge: Refuge
  isGm: boolean
  onChange: (patch: Partial<RefugeEditable>) => void
}

const REGION_KEYS = Object.keys(REGION_LABEL) as RegionKey[]

function AbundanceGrid({ region, isGm, onChange }: { region: Region; isGm: boolean; onChange: (k: NaturalKey, v: Abundance) => void }) {
  return (
    <div className="grid grid-cols-1 gap-2">
      {NATURAL_KEYS.map((k) => (
        <div key={k} className="flex items-center justify-between gap-2 rounded bg-zinc-900 px-3 py-2 text-sm text-zinc-300">
          {RESOURCE_LABEL[k]}
          <select aria-label={`Abundância de ${RESOURCE_LABEL[k]}`} disabled={!isGm} value={region.abundance[k]}
            onChange={(e) => onChange(k, e.target.value as Abundance)} className="rounded bg-zinc-800 px-2 py-1 text-sm">
            {Object.entries(ABUNDANCE_LABEL).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
          </select>
        </div>
      ))}
    </div>
  )
}

/** Collapsible region section: six 0-6 attributes plus abundance per natural resource. */
export function RegionPanel({ refuge, isGm, onChange }: Props) {
  const [open, setOpen] = useState(true)
  const region = refuge.region
  const set = (patch: Partial<Region>) => onChange({ region: { ...region, ...patch } })
  const setAbundance = (k: NaturalKey, v: Abundance) => set({ abundance: { ...region.abundance, [k]: v } })
  return (
    <section className="flex max-w-md flex-col gap-3">
      <button className="flex items-center gap-2 text-left text-sm text-zinc-400 hover:text-zinc-200" aria-expanded={open} onClick={() => setOpen(!open)}>
        {open ? <ChevronDown size={16} /> : <ChevronRight size={16} />}<Mountain size={16} />Região do Refúgio
      </button>
      {open && (
        <>
          <div className="grid grid-cols-1 gap-3">
            {REGION_KEYS.map((k) => (
              <LevelTrack key={k} title={REGION_LABEL[k]} value={region[k]} labels={REGION_SCALE} isGm={isGm} onChange={(v) => set({ [k]: v })} />
            ))}
          </div>
          <AbundanceGrid region={region} isGm={isGm} onChange={setAbundance} />
        </>
      )}
    </section>
  )
}
