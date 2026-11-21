import { expect, it } from 'vitest'
import { buildBody, buildVein, VOLCANO_RADIUS, volcanoRings } from './volcanoGeometry'

it('keeps the volcano inside the 3x3 footprint with a crater below the rim', () => {
  const rings = volcanoRings()
  for (const ring of rings) for (const [x, , z] of ring) expect(Math.hypot(x, z)).toBeLessThanOrEqual(VOLCANO_RADIUS + 1e-9)
  const outer = Math.max(...rings[rings.length - 1].map(([x, , z]) => Math.hypot(x, z)))
  expect(outer).toBeGreaterThan(1.2)
  expect(Math.min(...rings[2].map((v) => v[1]))).toBeGreaterThan(rings[0][0][1])
})

it('is deterministic and builds body and vein geometries', () => {
  expect(volcanoRings()).toEqual(volcanoRings())
  expect(buildBody().getAttribute('position').count).toBeGreaterThan(100)
  expect(buildVein(3).getIndex()?.count).toBeGreaterThan(0)
})
