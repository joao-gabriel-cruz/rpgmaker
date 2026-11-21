import {
  Apple, Axe, Backpack, Bandage, Battery, Bomb, Bone, Book, Briefcase, Cable, Cog, Cross, Crosshair, Drumstick, Droplet,
  Flame, Flashlight, Gem, Hammer, Key, Leaf, Lock, Map as MapIcon, Milk, Package, Pill, Radio, Scissors, Shield, Shirt,
  Skull, Slice, Sword, Swords, Syringe, Target, Tent, Trees, Wine, Wrench, type LucideIcon,
} from 'lucide-react'

const ICONS: Record<string, LucideIcon> = {
  sword: Sword, swords: Swords, slice: Slice, crosshair: Crosshair, target: Target, hammer: Hammer, axe: Axe, wrench: Wrench,
  key: Key, backpack: Backpack, wine: Wine, milk: Milk, pill: Pill, cross: Cross, bandage: Bandage, flashlight: Flashlight,
  map: MapIcon, radio: Radio, battery: Battery, cable: Cable, drumstick: Drumstick, apple: Apple, droplet: Droplet, trees: Trees,
  gem: Gem, book: Book, skull: Skull, shield: Shield, shirt: Shirt, flame: Flame, package: Package, cog: Cog, scissors: Scissors,
  tent: Tent, lock: Lock, bomb: Bomb, syringe: Syringe, briefcase: Briefcase, bone: Bone, leaf: Leaf,
}

/** Renders a curated lucide icon by name; unknown names fall back to a box. */
export function ItemIcon({ name, color, size = 20, className }: { name?: string; color?: string; size?: number; className?: string }) {
  const Icon = (name && ICONS[name]) || Package
  return <Icon size={size} color={color} aria-hidden className={className} data-icon={name ?? 'package'} />
}
