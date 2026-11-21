import type { Appearance } from './appearance'
import { normalizeAppearance } from './appearance'
import { Beard, Face, Glasses, HairBack, HairFront, Headwear, Mask, Shoulders } from './avatarArt'

const initialsOf = (name: string) => {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return '?'
  return (parts.length === 1 ? parts[0].slice(0, 2) : parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}

export interface CharacterAvatarProps {
  name: string
  /** `sheet.appearance`; anything shape-invalid is normalized. */
  appearance?: unknown
  portraitUrl?: string | null
  /** Ring color; defaults to the appearance token color. */
  ring?: string
  /** Tailwind size classes, e.g. `h-12 w-12`. */
  size?: string
  shape?: 'rounded' | 'circle'
  className?: string
}

/** 2D SVG bust drawn from the appearance (no WebGL), falling back to the portrait URL and then initials. */
export function CharacterAvatar({ name, appearance, portraitUrl, ring, size = 'h-12 w-12', shape = 'rounded', className = '' }: CharacterAvatarProps) {
  const look: Appearance | null = appearance ? normalizeAppearance(appearance) : null
  const color = ring ?? look?.tokenColor
  const base = `${size} shrink-0 overflow-hidden ${shape === 'circle' ? 'rounded-full' : 'rounded-lg'} bg-zinc-800 ${className}`
  const style = color ? { boxShadow: `0 0 0 2px ${color}` } : undefined
  if (look) {
    return (
      <span aria-hidden data-avatar="appearance" className={`${base} inline-block`} style={style}>
        <svg viewBox="0 0 48 48" className="block h-full w-full">
          <rect width="48" height="48" fill="#27272a" />
          <HairBack a={look} />
          <Shoulders a={look} />
          <Face a={look} />
          <Beard a={look} />
          <Mask a={look} />
          <Glasses a={look} />
          <HairFront a={look} />
          <Headwear a={look} />
        </svg>
      </span>
    )
  }
  if (portraitUrl) return <img src={portraitUrl} alt="" data-avatar="portrait" className={`${base} object-cover`} style={style} />
  return (
    <span aria-hidden data-avatar="initials" className={`${base} flex items-center justify-center text-sm font-semibold text-zinc-200`} style={style}>
      {initialsOf(name)}
    </span>
  )
}
