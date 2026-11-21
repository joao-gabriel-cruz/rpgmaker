export const BUILDS = ['slim', 'average', 'broad'] as const
export const HAIR_STYLES = ['none', 'short', 'long', 'ponytail', 'mohawk', 'curly'] as const
export const BEARDS = ['none', 'stubble', 'full'] as const
export const TOPS = ['tshirt', 'jacket', 'hoodie', 'vest', 'coat'] as const
export const BOTTOMS = ['pants', 'shorts', 'skirt'] as const
export const SHOES = ['boots', 'sneakers', 'barefoot'] as const
export const WEAPONS = ['none', 'pipe', 'axe', 'rifle', 'bow', 'knife'] as const
export const MASKS = ['none', 'gas', 'bandana'] as const
export const HEADWEAR = ['none', 'cap', 'hood', 'helmet'] as const

export const MIN_HEIGHT = 0.9
export const MAX_HEIGHT = 1.1
export const MAX_INFECTED = 10

export type Build = (typeof BUILDS)[number]
export type HairStyle = (typeof HAIR_STYLES)[number]
export type Beard = (typeof BEARDS)[number]
export type Top = (typeof TOPS)[number]
export type Bottom = (typeof BOTTOMS)[number]
export type Shoes = (typeof SHOES)[number]
export type Weapon = (typeof WEAPONS)[number]
export type Mask = (typeof MASKS)[number]
export type Headwear = (typeof HEADWEAR)[number]

/** Miniature description stored in `sheet.appearance` (validated by the server). */
export interface Appearance {
  body: { build: Build; height: number; skin: string }
  hair: { style: HairStyle; color: string }
  beard: Beard
  outfit: { top: Top; topColor: string; bottom: Bottom; bottomColor: string; shoes: Shoes; shoesColor: string }
  accessories: { backpack: boolean; weapon: Weapon; mask: Mask; headwear: Headwear; glasses: boolean }
  tokenColor: string
  /** Optional 0-10 infection for NPCs, which have no Assimilation level. */
  infected?: number
}

export const DEFAULT_APPEARANCE: Appearance = {
  body: { build: 'average', height: 1, skin: '#e0b48f' },
  hair: { style: 'short', color: '#4a3222' },
  beard: 'none',
  outfit: { top: 'jacket', topColor: '#4b5d3a', bottom: 'pants', bottomColor: '#3b4252', shoes: 'boots', shoesColor: '#2b2420' },
  accessories: { backpack: false, weapon: 'none', mask: 'none', headwear: 'none', glasses: false },
  tokenColor: '#34d399',
}

export const SKIN_COLORS = ['#f5d0b0', '#e0b48f', '#c68e5f', '#9a6a43', '#6b4429', '#3f2a1c']
export const HAIR_COLORS = ['#18120d', '#4a3222', '#8a5a2b', '#c9a24a', '#a8a8a8', '#a33a24', '#3b5bd6']
export const CLOTH_COLORS = ['#4b5d3a', '#3b4252', '#6b3a2b', '#7d8590', '#d6c9a8', '#1f2328', '#a33a24', '#2f5f7a']
export const TOKEN_COLORS = ['#34d399', '#60a5fa', '#f59e0b', '#f472b6', '#a78bfa', '#f87171', '#2dd4bf', '#a3e635']

const LABELS = {
  build: { slim: 'Esguio', average: 'Médio', broad: 'Robusto' },
  hair: { none: 'Careca', short: 'Curto', long: 'Longo', ponytail: 'Rabo de cavalo', mohawk: 'Moicano', curly: 'Cacheado' },
  beard: { none: 'Sem barba', stubble: 'Por fazer', full: 'Cheia' },
  top: { tshirt: 'Camiseta', jacket: 'Jaqueta', hoodie: 'Moletom', vest: 'Colete', coat: 'Casaco' },
  bottom: { pants: 'Calça', shorts: 'Bermuda', skirt: 'Saia' },
  shoes: { boots: 'Botas', sneakers: 'Tênis', barefoot: 'Descalço' },
  weapon: { none: 'Nenhuma', pipe: 'Cano', axe: 'Machado', rifle: 'Rifle', bow: 'Arco', knife: 'Faca' },
  mask: { none: 'Nenhuma', gas: 'Máscara de gás', bandana: 'Bandana' },
  headwear: { none: 'Nenhum', cap: 'Boné', hood: 'Capuz', helmet: 'Capacete' },
} as const
export const LABEL = LABELS

const HEX = /^#[0-9a-fA-F]{6}$/
const pick = <T extends string>(list: readonly T[], v: unknown, fb: T): T => (list.includes(v as T) ? (v as T) : fb)
const color = (v: unknown, fb: string) => (typeof v === 'string' && HEX.test(v) ? v : fb)
const rec = (v: unknown): Record<string, unknown> => (v && typeof v === 'object' ? (v as Record<string, unknown>) : {})

/** Fills gaps and drops invalid values so rendering never meets a partial or foreign shape. */
export function normalizeAppearance(raw: unknown): Appearance {
  const d = DEFAULT_APPEARANCE
  const r = rec(raw)
  const b = rec(r.body), h = rec(r.hair), o = rec(r.outfit), a = rec(r.accessories)
  const height = typeof b.height === 'number' ? Math.min(MAX_HEIGHT, Math.max(MIN_HEIGHT, b.height)) : d.body.height
  const out: Appearance = {
    body: { build: pick(BUILDS, b.build, d.body.build), height, skin: color(b.skin, d.body.skin) },
    hair: { style: pick(HAIR_STYLES, h.style, d.hair.style), color: color(h.color, d.hair.color) },
    beard: pick(BEARDS, r.beard, d.beard),
    outfit: {
      top: pick(TOPS, o.top, d.outfit.top), topColor: color(o.topColor, d.outfit.topColor),
      bottom: pick(BOTTOMS, o.bottom, d.outfit.bottom), bottomColor: color(o.bottomColor, d.outfit.bottomColor),
      shoes: pick(SHOES, o.shoes, d.outfit.shoes), shoesColor: color(o.shoesColor, d.outfit.shoesColor),
    },
    accessories: {
      backpack: a.backpack === true, glasses: a.glasses === true,
      weapon: pick(WEAPONS, a.weapon, 'none'), mask: pick(MASKS, a.mask, 'none'), headwear: pick(HEADWEAR, a.headwear, 'none'),
    },
    tokenColor: color(r.tokenColor, d.tokenColor),
  }
  if (typeof r.infected === 'number') out.infected = Math.min(MAX_INFECTED, Math.max(0, Math.round(r.infected)))
  return out
}

const oneOf = <T,>(list: readonly T[], rng: () => number): T => list[Math.floor(rng() * list.length) % list.length]

/** Random look; `keep` carries what must survive (token color, infection). */
export function randomAppearance(rng: () => number = Math.random, keep: Pick<Appearance, 'tokenColor' | 'infected'> = DEFAULT_APPEARANCE): Appearance {
  const cloth = () => oneOf(CLOTH_COLORS, rng)
  const out: Appearance = {
    body: { build: oneOf(BUILDS, rng), height: Math.round((MIN_HEIGHT + rng() * (MAX_HEIGHT - MIN_HEIGHT)) * 100) / 100, skin: oneOf(SKIN_COLORS, rng) },
    hair: { style: oneOf(HAIR_STYLES, rng), color: oneOf(HAIR_COLORS, rng) },
    beard: oneOf(BEARDS, rng),
    outfit: { top: oneOf(TOPS, rng), topColor: cloth(), bottom: oneOf(BOTTOMS, rng), bottomColor: cloth(), shoes: oneOf(SHOES, rng), shoesColor: cloth() },
    accessories: {
      backpack: rng() < 0.5, glasses: rng() < 0.3,
      weapon: oneOf(WEAPONS, rng), mask: oneOf(MASKS, rng), headwear: oneOf(HEADWEAR, rng),
    },
    tokenColor: keep.tokenColor,
  }
  if (keep.infected !== undefined) out.infected = keep.infected
  return out
}
