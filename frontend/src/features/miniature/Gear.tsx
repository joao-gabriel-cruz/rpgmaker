import type { Appearance } from './appearance'
import { ARM_W, BUILD_WIDTH, TORSO_TOP, armX } from './Body'
import { Part, arc, box, cyl } from './shared'

const wood = '#6b4a2f'
const metal = '#5f6368'

function Weapon({ a }: { a: Appearance }) {
  const hand: [number, number, number] = [armX(BUILD_WIDTH[a.body.build]) + ARM_W / 2 + 0.01, TORSO_TOP - 0.32, 0.03]
  const at = (pos: [number, number, number]): [number, number, number] => [hand[0] + pos[0], hand[1] + pos[1], hand[2] + pos[2]]
  switch (a.accessories.weapon) {
    case 'pipe': return <Part geo={cyl(0.015, 0.015, 0.5, 6)} color={metal} pos={at([0, 0.1, 0.05])} rot={[0.3, 0, 0]} />
    case 'axe':
      return <>
        <Part geo={cyl(0.015, 0.015, 0.45, 6)} color={wood} pos={at([0, 0.1, 0.05])} rot={[0.3, 0, 0]} />
        <Part geo={box(0.03, 0.09, 0.12)} color={metal} pos={at([0, 0.3, 0.14])} rot={[0.3, 0, 0]} />
      </>
    case 'rifle':
      return <>
        <Part geo={box(0.035, 0.05, 0.55)} color="#2a2d31" pos={at([-0.02, 0.05, 0.15])} />
        <Part geo={box(0.04, 0.08, 0.14)} color={wood} pos={at([-0.02, 0.03, -0.1])} />
      </>
    case 'bow':
      return <Part geo={arc(0.25, 0.012)} color={wood} pos={at([0.02, 0.05, 0.05])} rot={[0, Math.PI / 2, Math.PI / 2]} />
    case 'knife':
      return <Part geo={box(0.02, 0.02, 0.16)} color="#c5c9ce" pos={at([0, 0, 0.1])} />
    default: return null
  }
}

/** Backpack and held weapon. */
export function Gear({ a }: { a: Appearance }) {
  return (
    <>
      {a.accessories.backpack && (
        <>
          <Part geo={box(0.22, 0.28, 0.1)} color="#6b3a2b" pos={[0, 0.54, -0.135]} />
          <Part geo={box(0.18, 0.08, 0.02)} color="#4a2a20" pos={[0, 0.5, -0.19]} />
        </>
      )}
      <Weapon a={a} />
    </>
  )
}
