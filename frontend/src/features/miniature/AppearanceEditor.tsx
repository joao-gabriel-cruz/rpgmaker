import { lazy, Suspense } from 'react'
import { Section } from '../sheet-assimilacao/ui'
import { AccessoryControls, BodyControls, OutfitControls } from './AppearanceControls'
import { DEFAULT_APPEARANCE, normalizeAppearance, randomAppearance, type Appearance } from './appearance'
import { btnCls } from './fields'

const MiniaturePreview = lazy(() => import('./MiniaturePreview'))

interface Props {
  appearance: Appearance | null | undefined
  onChange: (next: Appearance) => void
  readOnly?: boolean
  /** Shown on the preview so marks can be seen while editing. */
  assimilationLevel?: number | null
}

/** Character page section: big rotating preview plus controls per category. */
export function AppearanceEditor({ appearance, onChange, readOnly = false, assimilationLevel }: Props) {
  const a = normalizeAppearance(appearance)
  const keep = { tokenColor: a.tokenColor, infected: a.infected }
  const actions = readOnly ? null : (
    <div className="flex gap-2">
      <button className={btnCls} onClick={() => onChange(randomAppearance(Math.random, keep))}>Aleatório</button>
      <button className={btnCls} onClick={() => onChange({ ...DEFAULT_APPEARANCE, ...keep })}>Restaurar padrão</button>
    </div>
  )
  return (
    <Section title="Aparência" aside={actions}>
      <div className="grid gap-4 lg:grid-cols-[minmax(240px,1fr)_2fr]">
        <div className="h-72 overflow-hidden rounded border border-zinc-800 lg:h-96">
          <Suspense fallback={<p className="p-3 text-sm text-zinc-400">Carregando prévia…</p>}>
            <MiniaturePreview appearance={a} assimilationLevel={assimilationLevel} />
          </Suspense>
        </div>
        <div className="flex flex-col gap-3">
          <BodyControls a={a} set={onChange} disabled={readOnly} />
          <OutfitControls a={a} set={onChange} disabled={readOnly} />
          <AccessoryControls a={a} set={onChange} disabled={readOnly} />
        </div>
      </div>
    </Section>
  )
}
