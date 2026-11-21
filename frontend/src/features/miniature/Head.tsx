import type { Appearance } from './appearance'
import { Part, arc, box, cyl, ico } from './shared'

const HEAD_Y = 0.82
const dark = '#202124'

function Hair({ a }: { a: Appearance }) {
  const c = a.hair.color
  const cap = <Part geo={box(0.2, 0.06, 0.2)} color={c} pos={[0, HEAD_Y + 0.085, -0.005]} />
  const back = <Part geo={box(0.2, 0.14, 0.05)} color={c} pos={[0, HEAD_Y + 0.02, -0.085]} />
  switch (a.hair.style) {
    case 'short': return <>{cap}{back}</>
    case 'long': return <>{cap}<Part geo={box(0.21, 0.3, 0.05)} color={c} pos={[0, HEAD_Y - 0.06, -0.09]} /></>
    case 'ponytail': return <>{cap}{back}<Part geo={cyl(0.02, 0.03, 0.2, 5)} color={c} pos={[0, HEAD_Y - 0.03, -0.14]} rot={[0.4, 0, 0]} /></>
    case 'mohawk': return <Part geo={box(0.03, 0.09, 0.22)} color={c} pos={[0, HEAD_Y + 0.11, 0]} />
    case 'curly':
      return <>{[[-0.07, 0], [0.07, 0], [0, 0.07], [0, -0.07], [0, 0]].map(([x, z], i) =>
        <Part key={i} geo={ico(0.055)} color={c} pos={[x, HEAD_Y + 0.09, z]} />)}</>
    default: return null
  }
}

function Beard({ a }: { a: Appearance }) {
  if (a.beard === 'none') return null
  const c = a.hair.color
  if (a.beard === 'stubble') return <Part geo={box(0.15, 0.05, 0.03)} color={c} pos={[0, HEAD_Y - 0.06, 0.085]} />
  return <Part geo={box(0.18, 0.1, 0.06)} color={c} pos={[0, HEAD_Y - 0.06, 0.07]} />
}

function Mask({ a }: { a: Appearance }) {
  if (a.accessories.mask === 'bandana') return <Part geo={box(0.19, 0.075, 0.19)} color="#a33a24" pos={[0, HEAD_Y - 0.055, 0.01]} />
  if (a.accessories.mask !== 'gas') return null
  return (
    <>
      <Part geo={cyl(0.03, 0.045, 0.07, 6)} color="#3a3f45" pos={[0, HEAD_Y - 0.045, 0.12]} rot={[Math.PI / 2, 0, 0]} />
      <Part geo={box(0.19, 0.05, 0.03)} color={dark} pos={[0, HEAD_Y + 0.015, 0.1]} />
      <Part geo={box(0.05, 0.05, 0.02)} color="#7dd3fc" pos={[-0.04, HEAD_Y + 0.015, 0.117]} emissive="#38bdf8" glow={0.4} />
      <Part geo={box(0.05, 0.05, 0.02)} color="#7dd3fc" pos={[0.04, HEAD_Y + 0.015, 0.117]} emissive="#38bdf8" glow={0.4} />
    </>
  )
}

function Glasses({ a }: { a: Appearance }) {
  if (!a.accessories.glasses || a.accessories.mask === 'gas') return null
  return (
    <>
      <Part geo={box(0.055, 0.04, 0.015)} color={dark} pos={[-0.04, HEAD_Y + 0.01, 0.1]} />
      <Part geo={box(0.055, 0.04, 0.015)} color={dark} pos={[0.04, HEAD_Y + 0.01, 0.1]} />
      <Part geo={box(0.03, 0.01, 0.015)} color={dark} pos={[0, HEAD_Y + 0.015, 0.1]} />
    </>
  )
}

function Headwear({ a }: { a: Appearance }) {
  switch (a.accessories.headwear) {
    case 'cap':
      return <><Part geo={box(0.22, 0.07, 0.22)} color={a.outfit.topColor} pos={[0, HEAD_Y + 0.09, 0]} /><Part geo={box(0.2, 0.015, 0.1)} color={a.outfit.topColor} pos={[0, HEAD_Y + 0.07, 0.15]} /></>
    case 'hood':
      return <Part geo={ico(0.125)} color={a.outfit.topColor} pos={[0, HEAD_Y + 0.02, -0.025]} scale={[1, 1.05, 1.05]} />
    case 'helmet':
      return <><Part geo={ico(0.125)} color="#59606a" pos={[0, HEAD_Y + 0.03, 0]} scale={[1, 0.9, 1]} /><Part geo={arc(0.11, 0.012)} color="#3a3f45" pos={[0, HEAD_Y + 0.03, 0.05]} rot={[0, 0, 0]} /></>
    default: return null
  }
}

/** Hair, beard, mask, glasses and headwear as separate meshes around the head. */
export function Head({ a }: { a: Appearance }) {
  const covered = a.accessories.headwear === 'helmet' || a.accessories.headwear === 'hood'
  return (
    <>
      {!covered && <Hair a={a} />}
      <Beard a={a} />
      <Mask a={a} />
      <Glasses a={a} />
      <Headwear a={a} />
    </>
  )
}
