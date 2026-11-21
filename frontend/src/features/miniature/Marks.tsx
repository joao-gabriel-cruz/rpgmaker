import { BUILD_WIDTH, ARM_W, TORSO_TOP, armX } from './Body'
import { marksFor, MAX_MARKS } from './assimilationMarks'
import { Part, box, type V3 } from './shared'

interface Slot { pos: V3; size: V3 }

/** Fixed spots on arms, neck and face, in the order they appear as the level rises. */
export function markSlots(w: number): Slot[] {
  const ax = armX(w)
  const off = ARM_W / 2 + 0.004
  const arm = (side: number, y: number, h: number): Slot => ({ pos: [side * (ax + off), TORSO_TOP - y, 0.005], size: [0.012, h, 0.05] })
  const front = (side: number, y: number): Slot => ({ pos: [side * ax, TORSO_TOP - y, ARM_W / 2 + 0.004], size: [0.05, 0.12, 0.012] })
  const slots: Slot[] = [
    arm(1, 0.15, 0.14), arm(-1, 0.2, 0.12), front(1, 0.1), front(-1, 0.14),
    { pos: [0.06, 0.8, 0.098], size: [0.03, 0.07, 0.014] }, { pos: [-0.05, 0.84, 0.098], size: [0.025, 0.05, 0.014] },
    arm(1, 0.28, 0.08), arm(-1, 0.06, 0.1), front(1, 0.22), front(-1, 0.04),
    { pos: [0, 0.725, 0.035], size: [0.05, 0.04, 0.012] }, { pos: [0.092, 0.82, 0.02], size: [0.012, 0.06, 0.05] },
    { pos: [-0.092, 0.84, -0.02], size: [0.012, 0.05, 0.05] }, { pos: [0.015, 0.89, 0.06], size: [0.05, 0.012, 0.03] },
  ]
  return slots.slice(0, MAX_MARKS)
}

/** Glowing veins and blotches that multiply and brighten with the Assimilation level. */
export function Marks({ level, build }: { level: number; build: keyof typeof BUILD_WIDTH }) {
  const { count, intensity } = marksFor(level)
  if (count === 0) return null
  const slots = markSlots(BUILD_WIDTH[build]).slice(0, count)
  return (
    <>
      {slots.map((s, i) => (
        <Part key={i} geo={box(...s.size)} color="#3b0764" emissive={i % 2 ? '#a3e635' : '#a855f7'} glow={0.5 + intensity * 1.8} pos={s.pos} />
      ))}
    </>
  )
}
