import type { Appearance, Build } from './appearance'
import { Part, box, cyl, ico, type V3 } from './shared'

export const BUILD_WIDTH: Record<Build, number> = { slim: 0.88, average: 1, broad: 1.18 }
export const ARM_W = 0.075
export const TORSO_TOP = 0.7
const LEG_H = 0.38
const darker = '#202124'

export const armX = (w: number) => 0.165 * w + ARM_W / 2

function Legs({ a, w }: { a: Appearance; w: number }) {
  const { bottom, bottomColor, shoes, shoesColor } = a.outfit
  const skin = a.body.skin
  const x = 0.07 * w + 0.01
  const full = bottom === 'pants'
  const upper = bottom === 'shorts' ? 0.17 : 0
  return (
    <>
      {[-1, 1].map((s) => (
        <group key={s} position={[s * x, 0, 0]}>
          <Part geo={box(0.095 * w, LEG_H, 0.1)} color={full ? bottomColor : skin} pos={[0, LEG_H / 2, 0]} />
          {upper > 0 && <Part geo={box(0.1 * w, upper, 0.105)} color={bottomColor} pos={[0, LEG_H - upper / 2, 0]} />}
          <Shoe kind={shoes} color={shoesColor} skin={skin} w={w} />
        </group>
      ))}
    </>
  )
}

function Shoe({ kind, color, skin, w }: { kind: Appearance['outfit']['shoes']; color: string; skin: string; w: number }) {
  if (kind === 'barefoot') return <Part geo={box(0.095 * w, 0.03, 0.15)} color={skin} pos={[0, 0.015, 0.025]} />
  if (kind === 'sneakers') return <Part geo={box(0.1 * w, 0.06, 0.17)} color={color} pos={[0, 0.03, 0.03]} />
  return (
    <>
      <Part geo={box(0.105 * w, 0.06, 0.17)} color={color} pos={[0, 0.03, 0.03]} />
      <Part geo={box(0.1 * w, 0.1, 0.11)} color={color} pos={[0, 0.1, 0]} />
    </>
  )
}

function Arm({ a, side, w }: { a: Appearance; side: 1 | -1; w: number }) {
  const { top, topColor } = a.outfit
  const skin = a.body.skin
  const sleeve = top === 'tshirt' ? 0.1 : top === 'vest' ? 0 : 0.27
  const lower = 0.27 - sleeve
  return (
    <group position={[side * armX(w), TORSO_TOP - 0.02, 0]}>
      {sleeve > 0 && <Part geo={box(ARM_W, sleeve, ARM_W)} color={topColor} pos={[0, -sleeve / 2, 0]} />}
      {lower > 0 && <Part geo={box(ARM_W, lower, ARM_W)} color={skin} pos={[0, -sleeve - lower / 2, 0]} />}
      <Part geo={box(0.06, 0.06, 0.06)} color={skin} pos={[0, -0.3, 0]} />
    </group>
  )
}

function Torso({ a, w }: { a: Appearance; w: number }) {
  const { top, topColor, bottom, bottomColor } = a.outfit
  const bulk = top === 'jacket' || top === 'coat' ? 1.08 : 1
  return (
    <>
      <Part geo={box(0.3 * w * bulk, 0.3, 0.17 * bulk)} color={topColor} pos={[0, LEG_H + 0.15 + 0.02, 0]} />
      {top === 'coat' && <Part geo={cyl(0.14 * w, 0.17 * w, 0.2, 8)} color={topColor} pos={[0, LEG_H - 0.05, 0]} />}
      {top === 'hoodie' && <Part geo={box(0.2, 0.07, 0.08)} color={topColor} pos={[0, TORSO_TOP + 0.02, -0.09]} />}
      {bottom === 'skirt' && <Part geo={cyl(0.15 * w, 0.2 * w, 0.18, 8)} color={bottomColor} pos={[0, LEG_H - 0.04, 0]} />}
    </>
  )
}

/** Legs, torso, arms, neck and head (without face details), feet at y = 0. */
export function Body({ a }: { a: Appearance }) {
  const w = BUILD_WIDTH[a.body.build]
  const headPos: V3 = [0, 0.82, 0]
  return (
    <>
      <Legs a={a} w={w} />
      <Torso a={a} w={w} />
      <Arm a={a} side={-1} w={w} />
      <Arm a={a} side={1} w={w} />
      <Part geo={box(0.06, 0.05, 0.06)} color={a.body.skin} pos={[0, 0.72, 0]} />
      <Part geo={ico(0.1)} color={a.body.skin} pos={headPos} scale={[1, 1.08, 0.95]} />
      <Part geo={box(0.03, 0.025, 0.02)} color={darker} pos={[-0.04, 0.83, 0.093]} />
      <Part geo={box(0.03, 0.025, 0.02)} color={darker} pos={[0.04, 0.83, 0.093]} />
    </>
  )
}
