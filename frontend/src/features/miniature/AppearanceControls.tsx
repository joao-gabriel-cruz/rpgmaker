import {
  BEARDS, BOTTOMS, BUILDS, CLOTH_COLORS, HAIR_COLORS, HAIR_STYLES, HEADWEAR, LABEL, MASKS, MAX_HEIGHT, MIN_HEIGHT,
  SHOES, SKIN_COLORS, TOKEN_COLORS, TOPS, WEAPONS, type Appearance,
} from './appearance'
import { ColorField, Group, SelectField, Toggle } from './fields'

interface Props {
  a: Appearance
  set: (next: Appearance) => void
  disabled: boolean
}

export function BodyControls({ a, set, disabled }: Props) {
  const body = (patch: Partial<Appearance['body']>) => set({ ...a, body: { ...a.body, ...patch } })
  const hair = (patch: Partial<Appearance['hair']>) => set({ ...a, hair: { ...a.hair, ...patch } })
  return (
    <>
      <Group title="Corpo">
        <SelectField label="Corpo" value={a.body.build} options={BUILDS} labels={LABEL.build} disabled={disabled} onChange={(build) => body({ build })} />
        <label className="flex flex-col gap-1 text-xs text-zinc-400">
          Altura ({a.body.height.toFixed(2)})
          <input type="range" aria-label="Altura" min={MIN_HEIGHT} max={MAX_HEIGHT} step={0.01} value={a.body.height} disabled={disabled}
            onChange={(e) => body({ height: Number(e.target.value) })} />
        </label>
        <ColorField label="Pele" value={a.body.skin} palette={SKIN_COLORS} disabled={disabled} onChange={(skin) => body({ skin })} />
      </Group>
      <Group title="Cabelo e barba">
        <SelectField label="Cabelo" value={a.hair.style} options={HAIR_STYLES} labels={LABEL.hair} disabled={disabled} onChange={(style) => hair({ style })} />
        <ColorField label="Cor do cabelo" value={a.hair.color} palette={HAIR_COLORS} disabled={disabled} onChange={(color) => hair({ color })} />
        <SelectField label="Barba" value={a.beard} options={BEARDS} labels={LABEL.beard} disabled={disabled} onChange={(beard) => set({ ...a, beard })} />
      </Group>
    </>
  )
}

export function OutfitControls({ a, set, disabled }: Props) {
  const o = (patch: Partial<Appearance['outfit']>) => set({ ...a, outfit: { ...a.outfit, ...patch } })
  return (
    <Group title="Roupa e calçado">
      <SelectField label="Parte de cima" value={a.outfit.top} options={TOPS} labels={LABEL.top} disabled={disabled} onChange={(top) => o({ top })} />
      <ColorField label="Cor de cima" value={a.outfit.topColor} palette={CLOTH_COLORS} disabled={disabled} onChange={(topColor) => o({ topColor })} />
      <SelectField label="Parte de baixo" value={a.outfit.bottom} options={BOTTOMS} labels={LABEL.bottom} disabled={disabled} onChange={(bottom) => o({ bottom })} />
      <ColorField label="Cor de baixo" value={a.outfit.bottomColor} palette={CLOTH_COLORS} disabled={disabled} onChange={(bottomColor) => o({ bottomColor })} />
      <SelectField label="Calçado" value={a.outfit.shoes} options={SHOES} labels={LABEL.shoes} disabled={disabled} onChange={(shoes) => o({ shoes })} />
      <ColorField label="Cor do calçado" value={a.outfit.shoesColor} palette={CLOTH_COLORS} disabled={disabled} onChange={(shoesColor) => o({ shoesColor })} />
    </Group>
  )
}

export function AccessoryControls({ a, set, disabled }: Props) {
  const acc = (patch: Partial<Appearance['accessories']>) => set({ ...a, accessories: { ...a.accessories, ...patch } })
  return (
    <Group title="Acessórios e token">
      <SelectField label="Arma" value={a.accessories.weapon} options={WEAPONS} labels={LABEL.weapon} disabled={disabled} onChange={(weapon) => acc({ weapon })} />
      <SelectField label="Máscara" value={a.accessories.mask} options={MASKS} labels={LABEL.mask} disabled={disabled} onChange={(mask) => acc({ mask })} />
      <SelectField label="Chapéu" value={a.accessories.headwear} options={HEADWEAR} labels={LABEL.headwear} disabled={disabled} onChange={(headwear) => acc({ headwear })} />
      <div className="flex flex-col gap-1">
        <Toggle label="Mochila" checked={a.accessories.backpack} disabled={disabled} onChange={(backpack) => acc({ backpack })} />
        <Toggle label="Óculos" checked={a.accessories.glasses} disabled={disabled} onChange={(glasses) => acc({ glasses })} />
      </div>
      <ColorField label="Cor do token" value={a.tokenColor} palette={TOKEN_COLORS} disabled={disabled} onChange={(tokenColor) => set({ ...a, tokenColor })} />
    </Group>
  )
}
