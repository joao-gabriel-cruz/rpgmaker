import type { Appearance } from './appearance'

const W = 48
const SKIN_SHADE = 'rgba(0,0,0,0.18)'

export function Shoulders({ a }: { a: Appearance }) {
  const c = a.outfit.topColor
  const wide = a.body.build === 'broad' ? 3 : a.body.build === 'slim' ? -3 : 0
  const top = a.outfit.top
  return (
    <g>
      <path d={`M${6 - wide} ${W} C${6 - wide} 36 14 33 24 33 C34 33 ${42 + wide} 36 ${42 + wide} ${W} Z`} fill={c} />
      <rect x="20" y="29" width="8" height="6" rx="2" fill={a.body.skin} />
      <rect x="20" y="29" width="8" height="3" rx="1.5" fill={SKIN_SHADE} />
      {top === 'hoodie' && <path d="M15 35 Q24 42 33 35 Q24 38 15 35Z" fill="rgba(0,0,0,0.25)" />}
      {(top === 'jacket' || top === 'coat') && <path d="M24 34 L24 48" stroke="rgba(0,0,0,0.3)" strokeWidth="1" />}
      {top === 'vest' && <path d="M17 35 L24 48 L31 35" fill="none" stroke="rgba(255,255,255,0.25)" strokeWidth="1" />}
      {a.accessories.backpack && <path d="M13 36 L15 48 M35 36 L33 48" stroke="rgba(0,0,0,0.4)" strokeWidth="2" />}
    </g>
  )
}

export function HairBack({ a }: { a: Appearance }) {
  const c = a.hair.color
  const { style } = a.hair
  return (
    <g fill={c}>
      {style === 'long' && <path d="M12 18 Q12 8 24 8 Q36 8 36 18 L37 34 L30 30 L18 30 L11 34Z" />}
      {style === 'ponytail' && <ellipse cx="35" cy="24" rx="4" ry="8" transform="rotate(-18 35 24)" />}
      {style === 'curly' && [14, 20, 28, 34].map((x) => <circle key={x} cx={x} cy={x === 20 || x === 28 ? 10 : 16} r="5.5" />)}
    </g>
  )
}

export function Face({ a }: { a: Appearance }) {
  const skin = a.body.skin
  return (
    <g>
      <ellipse cx="24" cy="21" rx="9.5" ry="11" fill={skin} />
      <circle cx="14.6" cy="22" r="1.8" fill={skin} />
      <circle cx="33.4" cy="22" r="1.8" fill={skin} />
      {a.accessories.mask !== 'gas' && (
        <>
          <circle cx="20.5" cy="21" r="1.1" fill="#1f2328" />
          <circle cx="27.5" cy="21" r="1.1" fill="#1f2328" />
        </>
      )}
    </g>
  )
}

export function Beard({ a }: { a: Appearance }) {
  if (a.beard === 'none') return null
  const full = a.beard === 'full'
  return <path d="M14.8 24 Q15 33 24 33 Q33 33 33.2 24 Q30 28 24 28 Q18 28 14.8 24Z" fill={a.hair.color} opacity={full ? 1 : 0.4} />
}

export function Mask({ a }: { a: Appearance }) {
  const { mask } = a.accessories
  if (mask === 'bandana') return <path d="M14.6 24 Q24 27 33.4 24 L32 30 Q24 35 16 30Z" fill={a.outfit.topColor} stroke="rgba(0,0,0,0.3)" strokeWidth="0.6" />
  if (mask !== 'gas') return null
  return (
    <g>
      <path d="M14.8 19 Q24 16 33.2 19 L32 30 Q24 35 16 30Z" fill="#2b3036" />
      <circle cx="20.5" cy="21" r="2.6" fill="#9fb4c2" stroke="#111" strokeWidth="0.8" />
      <circle cx="27.5" cy="21" r="2.6" fill="#9fb4c2" stroke="#111" strokeWidth="0.8" />
      <circle cx="24" cy="30" r="2.2" fill="#444b52" stroke="#111" strokeWidth="0.8" />
    </g>
  )
}

export function Glasses({ a }: { a: Appearance }) {
  if (!a.accessories.glasses || a.accessories.mask === 'gas') return null
  return (
    <g fill="none" stroke="#111" strokeWidth="1">
      <circle cx="20.5" cy="21" r="3" />
      <circle cx="27.5" cy="21" r="3" />
      <path d="M23.5 21 L24.5 21" />
    </g>
  )
}

export function HairFront({ a }: { a: Appearance }) {
  const { style, color } = a.hair
  if (style === 'none') return null
  if (style === 'mohawk') return <path d="M21 4 L27 4 L28 17 L20 17Z" fill={color} />
  return <path d="M14 18 Q13 8 24 8 Q35 8 34 18 Q31 13 24 13 Q17 13 14 18Z" fill={color} />
}

export function Headwear({ a }: { a: Appearance }) {
  const w = a.accessories.headwear
  if (w === 'cap') return <g fill={a.outfit.topColor}><path d="M13.5 17 Q14 6 24 6 Q34 6 34.5 17Z" /><path d="M24 15 H40 Q40 18 34 18 H24Z" fill="rgba(0,0,0,0.3)" /></g>
  if (w === 'helmet') return <g><path d="M12.5 20 Q12 5 24 5 Q36 5 35.5 20 L33 17 H15Z" fill="#6b7280" /><path d="M24 5 V15" stroke="rgba(0,0,0,0.25)" /></g>
  if (w === 'hood') return <path d="M10 36 Q8 6 24 5 Q40 6 38 36 L34 24 Q34 12 24 12 Q14 12 14 24Z" fill={a.outfit.topColor} />
  return null
}
